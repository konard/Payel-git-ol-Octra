package main

import (
	"log"
	"os"

	"nodes/internal/fetcher/grpc"
	"nodes/internal/redis"
	"nodes/internal/service/agents"
	"nodes/internal/service/rules/boss"
	"nodes/internal/service/rules/manager"
	"nodes/internal/service/rules/worker"
	"nodes/pkg/database"
)

// main — единая точка входа для nodes.
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

	port := os.Getenv("NODES_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	log.Printf("Starting nodes gRPC server on port %s (boss+manager+worker in one process)", port)
	if err := grpc.Start(port, bossSvc, redisClient); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
