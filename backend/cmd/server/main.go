package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/api"
	"backend/internal/cli"
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/llm"
	"backend/internal/nix"
	"backend/internal/oauth"
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
	apiKeys := repository.NewAPIKeyRepository(db)
	dashboardEnvs := repository.NewDashboardEnvironmentRepository(db)
	clis := repository.NewCLIRepository(db)

	nixMgr := nix.NewManager(cfg.EnvironmentsDir, nil)

	// Provision all built-in CLIs into the system Nix profile so they are
	// available to every user environment without per-user installation.
	if nix.Available() {
		if err := nixMgr.ProvisionSystem(ctx); err != nil {
			log.Printf("nix provision (non-fatal): %v", err)
		} else {
			log.Println("built-in CLIs provisioned into system profile")
		}
	}

	cliMgr := cli.NewManager(cli.ExecLauncher{}, cli.NewRedisStateStore(rdb), cfg.CLITTL)
	defer cliMgr.Shutdown()
	llmClient := llm.New(nil)

	billingSvc := service.NewBillingService(users, agents, transactions, usageMetrics)
	authSvc := service.NewAuthServiceWithKeys(users, apiKeys, cfg, transactions)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, nixMgr, billingSvc)
	chatSvc := service.NewChatService(agents, cliMgr, llmClient, nixMgr)
	oauthH := oauth.New(authSvc, cfg)

	var tsClient *ts.Client
	if cfg.TypesenseHost != "" {
		tsClient = ts.New(cfg.TypesenseHost, cfg.TypesenseAPIKey)
		if err := tsClient.EnsureCollection(ctx); err != nil {
			log.Printf("typesense: %v (search will be unavailable)", err)
			tsClient = nil
		} else if err := tsClient.EnsureCLICollection(ctx); err != nil {
			log.Printf("typesense cli: %v", err)
		}
	}

	if tsClient != nil && os.Getenv("SKILL_SYNC_DISABLED") == "" {
		syncSvc := service.NewSkillSyncService(skills, tsClient, 24*time.Hour, cfg.BaseURLGetSkills)
		syncSvc.Start(ctx)
	}

	// Seed CLI catalogue from registry into DB + Typesense.
	seedCLIs(ctx, clis, tsClient)

	handler := loggingMiddleware(api.New(authSvc, envSvc, chatSvc, billingSvc, oauthH, dashboardEnvs, clis, tsClient).Router().Handler)
	server := &fasthttp.Server{Handler: handler, Name: "octra"}

	go func() {
		log.Printf("octra backend listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(cfg.HTTPAddr); err != nil {
			log.Fatalf("http server: %v", err)
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
