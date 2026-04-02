package grok

import (
	"context"
	"fmt"
	"log"

	"agents/pkg/models"

	grok "github.com/SimonMorphy/grok-go"
)

// GrokProvider implements the AIProvider interface for Grok (via xAI)
type GrokProvider struct {
	client *grok.Client
	config *models.ProviderConfig
}

// New creates a new Grok provider
func New(config *models.ProviderConfig) (*GrokProvider, error) {
	if config.APIKey == "" {
		log.Printf("Grok provider not configured - missing API key")
		return &GrokProvider{config: config}, nil
	}

	client, err := grok.NewClient(config.APIKey)
	if err != nil {
		log.Printf("Failed to create Grok client: %v", err)
		return nil, err
	}

	return &GrokProvider{
		client: client,
		config: config,
	}, nil
}

// Generate generates content using Grok
func (p *GrokProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	if p.client == nil {
		return "", fmt.Errorf("Grok provider not configured")
	}

	model := p.config.Model
	if model == "" {
		model = "grok-2"
	}

	resp, err := grok.CreateChatCompletion(ctx, p.client, &grok.ChatCompletionRequest{
		Model: model,
		Messages: []grok.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: float64(p.config.Temperature),
	})

	if err != nil {
		log.Printf("Grok API error: %v", err)
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no content in Grok response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Name returns the provider name
func (p *GrokProvider) Name() string {
	return "grok"
}

// IsConfigured checks if the provider is properly configured
func (p *GrokProvider) IsConfigured() bool {
	return p.client != nil && p.config != nil && p.config.APIKey != ""
}
