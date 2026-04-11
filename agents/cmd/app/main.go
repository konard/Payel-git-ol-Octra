package main

import (
	"log"
	"os"

	"agents/internal/service"
	"agents/pkg/fetcher/grpc"
	"agents/pkg/models"
)

func main() {
	// Initialize agent service
	agentService := service.NewAgentService()

	// API keys will be provided per-request via tokens field
	providers := map[string]*models.ProviderConfig{
		"claude":     {Model: "claude-opus-4-6"},
		"gemini":     {Model: "gemini-2.5-flash"},
		"openai":     {Model: "gpt-4o"},
		"deepseek":   {Model: "deepseek-chat"},
		"grok":       {Model: "grok-3"},
		"openrouter": {Model: "qwen/qwen3-coder"},
		"qwen":       {Model: "qwen-turbo"},
		"zai":        {Model: "glm-4.5-air"},
	}

	// Initialize and register each provider
	InitProvider(providers, agentService)

	availableProviders := agentService.GetAvailableProviders()
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
