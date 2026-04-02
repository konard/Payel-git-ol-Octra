package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"agents/pkg/fetcher/grpc"
)

// AgentClientWrapper wraps the agents gRPC client
type AgentClientWrapper struct {
	client *grpc.AgentClient
}

// NewAgentClientWrapper creates a new agent client wrapper
func NewAgentClientWrapper() (*AgentClientWrapper, error) {
	// Get agents service host from environment, default to localhost:50053
	agentsHost := os.Getenv("AGENTS_SERVICE_HOST")
	if agentsHost == "" {
		agentsHost = "localhost:50053"
	}

	client, err := grpc.NewAgentClient(agentsHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent client: %w", err)
	}

	return &AgentClientWrapper{client: client}, nil
}

// GenerateFromTask calls the agents service to generate content
func (w *AgentClientWrapper) GenerateFromTask(ctx context.Context, provider, model, prompt string, tokens map[string]string) (string, error) {
	log.Printf("🤖 Calling agents service (provider=%s, model=%s)", provider, model)

	content, err := w.client.Generate(ctx, provider, model, prompt, tokens, 4096, 0.7)
	if err != nil {
		return "", err
	}

	return content, nil
}

// Close closes the client connection
func (w *AgentClientWrapper) Close() error {
	if w.client != nil {
		return w.client.Close()
	}
	return nil
}
