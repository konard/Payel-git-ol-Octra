package main

import (
	"log"
	"os"

	"orchestrator/internal/fetcher/grpc"
	"orchestrator/internal/redis"
	"orchestrator/internal/service/agents"
	"orchestrator/internal/service/rules/boss"
	"orchestrator/internal/service/rules/manager"
	"orchestrator/internal/service/rules/worker"
	"orchestrator/pkg/database"
)

// main — единая точка входа для orchestrator.
// Раньше boss, manager и worker запускались отдельными процессами и общались
// через gRPC; теперь все три роли живут в одном процессе и зовут друг друга
// напрямую через Go.
func main() {
	database.InitDb()

	agentsClient, err := agents.NewClient()
	if err != nil {
		log.Fatalf("Failed to init agents client: %v", err)
	}
	defer agentsClient.Close()

	redisClient := redis.NewClient()

	workerSvc := worker.NewService(agentsClient)
	managerSvc := manager.NewService(agentsClient, workerSvc)
	bossSvc := boss.NewService(agentsClient, managerSvc, redisClient)

	port := os.Getenv("ORCHESTRATOR_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	log.Printf("Starting orchestrator gRPC server on port %s (boss+manager+worker in one process)", port)
	if err := grpc.Start(port, bossSvc, redisClient); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

