// Command server is the entrypoint for the Octra monolith: a single backend
// that aggregates AI CLIs and exposes a personal MCP endpoint per user.
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

	// Repositories.
	users := repository.NewUserRepository(db)
	agents := repository.NewAgentRepository(db)
	skills := repository.NewSkillRepository(db)
	userSkills := repository.NewUserSkillRepository(db)

	// Infrastructure.
	nixMgr := nix.NewManager(cfg.EnvironmentsDir, nil)
	cliMgr := cli.NewManager(cli.ExecLauncher{}, cli.NewRedisStateStore(rdb), cfg.CLITTL)
	defer cliMgr.Shutdown()
	llmClient := llm.New(nil)

	// Services.
	authSvc := service.NewAuthService(users)
	envSvc := service.NewEnvironmentService(agents, skills, userSkills, nixMgr)
	chatSvc := service.NewChatService(agents, cliMgr, llmClient, nixMgr)

	// HTTP.
	handler := api.New(authSvc, envSvc, chatSvc).Router().Handler
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
