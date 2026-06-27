package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/api"
	"backend/internal/cli"
	"backend/internal/config"
	"backend/internal/llm"
	"backend/internal/nix"
	"backend/internal/oauth"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/storage"

	"github.com/valyala/fasthttp"
)

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

	nixMgr := nix.NewManager(cfg.EnvironmentsDir, nil)
	cliMgr := cli.NewManager(cli.ExecLauncher{}, cli.NewRedisStateStore(rdb), cfg.CLITTL)
	defer cliMgr.Shutdown()
	llmClient := llm.New(nil)

	billingSvc := service.NewBillingService(users, agents, transactions, usageMetrics)
	authSvc := service.NewAuthServiceWithKeys(users, apiKeys, cfg, transactions)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, nixMgr, billingSvc)
	chatSvc := service.NewChatService(agents, cliMgr, llmClient, nixMgr)
	oauthH := oauth.New(authSvc, cfg)

	handler := api.New(authSvc, envSvc, chatSvc, billingSvc, oauthH).Router().Handler
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
