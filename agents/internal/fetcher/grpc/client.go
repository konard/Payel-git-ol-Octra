package grpc

import (
	"context"
	"fmt"
	"log"

	"agents/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient wraps the gRPC client for Agents service
type AgentClient struct {
	client pb.AgentServiceClient
	conn   *grpc.ClientConn
}

// NewAgentClient creates a new client connection to agents service
func NewAgentClient(host string) (*AgentClient, error) {
	if host == "" {
		host = "localhost:50053"
	}

	conn, err := grpc.NewClient(host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024),
			grpc.MaxCallSendMsgSize(100*1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agents service at %s: %w", host, err)
	}

	log.Printf("✅ Connected to Agents service at %s", host)

	return &AgentClient{
		client: pb.NewAgentServiceClient(conn),
		conn:   conn,
	}, nil
}

// Generate calls the Generate RPC on the agents service
func (c *AgentClient) Generate(ctx context.Context, provider, model, prompt string, tokens map[string]string, maxTokens int32, temperature float32) (string, error) {
	// Convert tokens map back to gRPC format
	grpcTokens := make(map[string]string)
	for k, v := range tokens {
		grpcTokens[k] = v
	}

	req := &pb.GenerateRequest{
		Provider:    provider,
		Model:       model,
		Prompt:      prompt,
		Tokens:      grpcTokens,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	resp, err := c.client.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate with %s: %w", provider, err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("agent service error: %s (code: %s)", resp.Error, resp.ErrorCode)
	}

	return resp.Content, nil
}

// Close closes the client connection
func (c *AgentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
