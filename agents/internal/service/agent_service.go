package service

import (
	"context"
	"fmt"
	"log"

	"agents/internal/fetcher/providers/claude"
	"agents/internal/fetcher/providers/custom"
	"agents/internal/fetcher/providers/deepseek"
	"agents/internal/fetcher/providers/gemini"
	"agents/internal/fetcher/providers/grok"
	"agents/internal/fetcher/providers/openai"
	"agents/internal/fetcher/providers/qwen"
	"agents/internal/fetcher/providers/zai"
	"agents/pkg/models"
)

// AgentService routes AI requests to the appropriate provider
type AgentService struct {
	providers map[string]models.AIProvider
}

// NewAgentService creates a new agent service
func NewAgentService() *AgentService {
	return &AgentService{
		providers: make(map[string]models.AIProvider),
	}
}

// RegisterProvider registers a provider
func (s *AgentService) RegisterProvider(name string, provider models.AIProvider) {
	s.providers[name] = provider
	log.Printf("✅ Registered AI provider: %s", name)
}

// RegisterCustomProvider registers a custom provider for a specific user
func (s *AgentService) RegisterCustomProvider(userId, providerName string, config *models.ProviderConfig) error {
	// Set base URL in tokens for custom provider
	if config.Tokens == nil {
		config.Tokens = make(map[string]interface{})
	}
	// Note: base_url should be set when creating the config

	provider, err := custom.New(config)
	if err != nil {
		return fmt.Errorf("failed to create custom provider: %w", err)
	}

	// Use user-specific provider name to avoid conflicts
	fullProviderName := fmt.Sprintf("custom-%s-%s", userId, providerName)
	s.RegisterProvider(fullProviderName, provider)

	return nil
}

// InitializeAllProviders initializes all providers from environment/config
func (s *AgentService) InitializeAllProviders(configs map[string]*models.ProviderConfig) error {
	// Initialize Claude
	if cfg, ok := configs["claude"]; ok {
		if p, err := claude.New(cfg); err == nil {
			s.RegisterProvider("claude", p)
		} else {
			log.Printf("⚠️  Failed to initialize Claude: %v", err)
		}
	}

	// Initialize Gemini
	if cfg, ok := configs["gemini"]; ok {
		if p, err := gemini.New(cfg); err == nil {
			s.RegisterProvider("gemini", p)
		} else {
			log.Printf("⚠️  Failed to initialize Gemini: %v", err)
		}
	}

	// Initialize OpenAI
	if cfg, ok := configs["openai"]; ok {
		if p, err := openai.New(cfg); err == nil {
			s.RegisterProvider("openai", p)
		} else {
			log.Printf("⚠️  Failed to initialize OpenAI: %v", err)
		}
	}

	// Initialize DeepSeek
	if cfg, ok := configs["deepseek"]; ok {
		if p, err := deepseek.New(cfg); err == nil {
			s.RegisterProvider("deepseek", p)
		} else {
			log.Printf("⚠️  Failed to initialize DeepSeek: %v", err)
		}
	}

	// Initialize Grok
	if cfg, ok := configs["grok"]; ok {
		if p, err := grok.New(cfg); err == nil {
			s.RegisterProvider("grok", p)
		} else {
			log.Printf("⚠️  Failed to initialize Grok: %v", err)
		}
	}

	// Initialize Qwen
	if cfg, ok := configs["qwen"]; ok {
		if p, err := qwen.New(cfg); err == nil {
			s.RegisterProvider("qwen", p)
		} else {
			log.Printf("⚠️  Failed to initialize Qwen: %v", err)
		}
	}

	// Initialize Z.AI
	if cfg, ok := configs["zai"]; ok {
		if p, err := zai.New(cfg); err == nil {
			s.RegisterProvider("zai", p)
		} else {
			log.Printf("⚠️  Failed to initialize Z.AI: %v", err)
		}
	}

	return nil
}

// Generate generates content using the specified provider
func (s *AgentService) Generate(ctx context.Context, req *models.AgentRequest) (*models.AgentResponse, error) {
	provider, ok := s.providers[req.Provider]
	if !ok {
		return &models.AgentResponse{
			Provider:  req.Provider,
			Model:     req.Model,
			Error:     fmt.Sprintf("provider '%s' not found", req.Provider),
			ErrorCode: "PROVIDER_NOT_FOUND",
		}, fmt.Errorf("provider '%s' not found", req.Provider)
	}

	if !provider.IsConfigured() {
		return &models.AgentResponse{
			Provider:  req.Provider,
			Model:     req.Model,
			Error:     fmt.Sprintf("provider '%s' not configured", req.Provider),
			ErrorCode: "PROVIDER_NOT_CONFIGURED",
		}, fmt.Errorf("provider '%s' not configured", req.Provider)
	}

	log.Printf("🤖 [%s] Generating content with model: %s", provider.Name(), req.Model)

	content, err := provider.Generate(ctx, req.Prompt, req.Tokens)
	if err != nil {
		log.Printf("❌ Error generating content: %v", err)
		return &models.AgentResponse{
			Provider:  req.Provider,
			Model:     req.Model,
			Error:     err.Error(),
			ErrorCode: "GENERATION_ERROR",
		}, err
	}

	return &models.AgentResponse{
		Provider: req.Provider,
		Model:    req.Model,
		Content:  content,
	}, nil
}

// GetAvailableProviders returns a list of available providers
func (s *AgentService) GetAvailableProviders() []string {
	var providers []string
	for name, p := range s.providers {
		if p.IsConfigured() {
			providers = append(providers, name)
		}
	}
	return providers
}
