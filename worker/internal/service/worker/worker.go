package worker

import (
	"log"
	"os"
	"sync"

	"agents/pkg/fetcher/grpc"
	"worker/internal/fetcher/grpc/workerpb"
)

// WorkerService — worker service
type WorkerService struct {
	workerpb.UnimplementedWorkerServiceServer
	agentsClient *grpc.AgentClient
	mu           sync.Mutex
}

func NewWorkerService() *WorkerService {
	agentsHost := os.Getenv("AGENTS_SERVICE_HOST")
	if agentsHost == "" {
		agentsHost = "localhost:50053"
	}

	agentsClient, err := grpc.NewAgentClient(agentsHost)
	if err != nil {
		log.Printf("Warning: failed to connect to agents service: %v", err)
	}

	return &WorkerService{
		agentsClient: agentsClient,
	}
}
