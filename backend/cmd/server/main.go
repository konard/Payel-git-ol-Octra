package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/api"
	"backend/internal/cli"
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/nix"
	"backend/internal/oauth"
	"backend/internal/provider"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/storage"
	ts "backend/internal/typesense"

	"github.com/valyala/fasthttp"
)

func seedCLIs(ctx context.Context, cliRepo repository.CLIRepository, tsClient *ts.Client) {
	builtins := cli.BuiltinCLIs()
	for _, p := range builtins {
		record := &model.CLI{
			Name:       p.Name,
			NixAttr:    p.NixAttr,
			InstallCmd: p.InstallCmd,
		}
		if err := cliRepo.Upsert(ctx, record); err != nil {
			log.Printf("seed cli %s: %v", p.Name, err)
		}
	}
	if tsClient == nil {
		return
	}
	all, err := cliRepo.List(ctx)
	if err != nil {
		log.Printf("seed cli list: %v", err)
		return
	}
	docs := make([]ts.CLIDocument, len(all))
	for i, cli := range all {
		docs[i] = ts.CLIDocument{
			ID:         cli.ID.String(),
			Name:       cli.Name,
			NixAttr:    cli.NixAttr,
			InstallCmd: cli.InstallCmd,
		}
	}
	if err := tsClient.IndexCLIs(ctx, docs); err != nil {
		log.Printf("seed cli index: %v", err)
	} else {
		log.Printf("seeded %d CLIs into Typesense", len(docs))
	}
}

func seedProviders(ctx context.Context, providerRepo repository.ProviderRepository, tsClient *ts.Client) {
	builtins := provider.BuiltinProviders()
	for _, p := range builtins {
		record := &model.Provider{
			Key:          p.Key,
			Name:         p.Name,
			BaseURL:      p.BaseURL,
			AuthEnv:      p.AuthEnv,
			DefaultModel: p.DefaultModel,
			Description:  p.Description,
		}
		if err := providerRepo.Upsert(ctx, record); err != nil {
			log.Printf("seed provider %s: %v", p.Key, err)
		}
	}
	if tsClient == nil {
		return
	}
	all, err := providerRepo.List(ctx)
	if err != nil {
		log.Printf("seed provider list: %v", err)
		return
	}
	docs := make([]ts.ProviderDocument, len(all))
	for i, p := range all {
		docs[i] = ts.ProviderDocument{
			ID:           p.ID.String(),
			Key:          p.Key,
			Name:         p.Name,
			BaseURL:      p.BaseURL,
			AuthEnv:      p.AuthEnv,
			DefaultModel: p.DefaultModel,
			Description:  p.Description,
		}
	}
	if err := tsClient.IndexProviders(ctx, docs); err != nil {
		log.Printf("seed provider index: %v", err)
	} else {
		log.Printf("seeded %d providers into Typesense", len(docs))
	}
}

func loggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		next(ctx)
		log.Printf("%s %s %d %s", ctx.Method(), ctx.Path(), ctx.Response.StatusCode(), time.Since(start))
	}
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := storage.OpenPostgres(cfg.DBDsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	rdb, err := storage.OpenRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	users := repository.NewUserRepository(db)
	agents := repository.NewAgentRepository(db)
	skills := repository.NewSkillRepository(db)
	userSkills := repository.NewUserSkillRepository(db)
	transactions := repository.NewTransactionRepository(db)
	usageMetrics := repository.NewUsageMetricsRepository(db)
	requestMetrics := repository.NewRequestMetricsRepository(db)
	apiKeys := repository.NewAPIKeyRepository(db)
	dashboardEnvs := repository.NewDashboardEnvironmentRepository(db)
	canvasNodes := repository.NewCanvasNodeRepository(db)
	clis := repository.NewCLIRepository(db)
	providers := repository.NewProviderRepository(db)

	nixMgr := nix.NewManager(cfg.EnvironmentsDir, nil)

	ocaweAddr := os.Getenv("OCAWE_ADDR")
	var ocaweLauncher cli.Launcher
	if ocaweAddr != "" {
		parsed, err := url.Parse(ocaweAddr)
		if err != nil {
			log.Fatalf("invalid OCAWE_ADDR: %v", err)
		}
		ocaweLauncher = cli.RemoteOcaweLauncher{BaseURL: parsed}
	} else {
		ocaweLauncher = cli.OcaweLauncher{}
	}

	cliMgr := cli.NewManager(ocaweLauncher, cli.NewRedisStateStore(rdb), cfg.CLITTL)
	defer cliMgr.Shutdown()

	billingSvc := service.NewBillingService(users, agents, transactions, usageMetrics)
	metricsSvc := service.NewMetricsService(requestMetrics, dashboardEnvs)
	authSvc := service.NewAuthServiceWithKeys(users, apiKeys, cfg, transactions)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, nixMgr, billingSvc)
	chatSvc := service.NewChatService(agents, cliMgr, nixMgr).
		WithEnvironmentRepos(dashboardEnvs, canvasNodes).
		WithBaseURL(ocaweAddr)
	oauthH := oauth.New(authSvc, cfg)

	var tsClient *ts.Client
	if cfg.TypesenseHost != "" {
		tsClient = ts.New(cfg.TypesenseHost, cfg.TypesenseAPIKey)
		if err := tsClient.EnsureCollection(ctx); err != nil {
			log.Printf("typesense: %v (search will be unavailable)", err)
			tsClient = nil
		}
		if tsClient != nil {
			if err := tsClient.EnsureCLICollection(ctx); err != nil {
				log.Printf("typesense cli collection: %v", err)
				tsClient = nil
			}
		}
		if tsClient != nil {
			if err := tsClient.EnsureProviderCollection(ctx); err != nil {
				log.Printf("typesense provider collection: %v", err)
				tsClient = nil
			}
		}
	}

	if tsClient != nil && os.Getenv("SKILL_SYNC_DISABLED") == "" {
		syncSvc := service.NewSkillSyncService(skills, tsClient, 24*time.Hour, cfg.BaseURLGetSkills)
		syncSvc.Start(ctx)
	}

	// Seed CLI catalogue from registry into DB + Typesense.
	seedCLIs(ctx, clis, tsClient)
	seedProviders(ctx, providers, tsClient)

	handler := loggingMiddleware(api.New(authSvc, envSvc, chatSvc, billingSvc, metricsSvc, oauthH, dashboardEnvs, canvasNodes, skills, clis, providers, tsClient, cliMgr).Router().Handler)
	server := &fasthttp.Server{Handler: handler, Name: "octra"}

	go func() {
		log.Printf("octra backend listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(cfg.HTTPAddr); err != nil {
			log.Fatalf("http server: %v", err)
		}
	}()

	// Provision built-in CLIs in the background so the HTTP server starts
	// accepting requests immediately instead of blocking on Nix downloads.
	if nix.Available() {
		go func() {
			if err := nixMgr.ProvisionSystem(ctx); err != nil {
				log.Printf("nix provision (non-fatal): %v", err)
			} else {
				log.Println("built-in CLIs provisioned into system profile")
			}
		}()
	}

	// Sync unbuilt environments to ocawe in the background.
	go func() {
		if err := chatSvc.SyncEnvironmentBuilds(ctx); err != nil {
			log.Printf("env sync: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	if err := server.Shutdown(); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
