package main

import (
	"log"
	"os"

	"agents/pkg/fetcher/grpc"
	"agents/internal/service"
)

func main() {
	// Initialize agent service
	// NOTE: API keys are provided per-request via tokens field
	// NOT loaded from environment
	agentService := service.NewAgentService()

	// Load optional environmental configuration (model defaults, feature flags, etc.)
	availableProviders := []string{"claude", "gemini", "openai", "deepseek", "grok"}
	log.Printf("✅ Available AI providers: %v", availableProviders)

	// Start gRPC server
	port := os.Getenv("AGENTS_PORT")
	if port == "" {
		port = "50053"
	}

	log.Printf("Starting Agents gRPC server on port %s...", port)
	if err := grpc.Start(port, agentService); err != nil {
		log.Fatalf("❌ Failed to start Agents server: %v", err)
	}
}
