package agents

import (
	"context"
	"fmt"
	"log"
	"os"

	"orchestrator/internal/config"

	"agents/pkg/fetcher/grpc"
)

// Client — единый клиент к сервису agents для всех узлов (boss/manager/worker)
type Client struct {
	rpc *grpc.AgentClient
}

// NewClient — устанавливает соединение к agents через AGENTS_SERVICE_HOST
func NewClient() (*Client, error) {
	host := os.Getenv("AGENTS_SERVICE_HOST")
	if host == "" {
		host = "localhost:50053"
	}
	rpc, err := grpc.NewAgentClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent client: %w", err)
	}
	return &Client{rpc: rpc}, nil
}

// Generate — низкоуровневый вызов LLM через agents
func (c *Client) Generate(ctx context.Context, provider, model, prompt string, tokens map[string]string, maxTokens int32, temperature float32) (string, error) {
	if c == nil || c.rpc == nil {
		return "", fmt.Errorf("agents client is not initialized")
	}
	return c.rpc.Generate(ctx, provider, model, prompt, tokens, maxTokens, temperature)
}

// GenerateFromTask — обёртка с параметрами по умолчанию (4096 токенов, единая
// детерминированная температура config.Temperature)
func (c *Client) GenerateFromTask(ctx context.Context, provider, model, prompt string, tokens map[string]string) (string, error) {
	log.Printf("Calling agents service (provider=%s, model=%s)", provider, model)
	return c.Generate(ctx, provider, model, prompt, tokens, 4096, config.Temperature)
}

// Close — закрывает соединение
func (c *Client) Close() error {
	if c != nil && c.rpc != nil {
		return c.rpc.Close()
	}
	return nil
}

