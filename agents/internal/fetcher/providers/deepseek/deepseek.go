package deepseek

import (
	"agents/pkg/models"
	"context"
	"fmt"
	"log"

	"github.com/cohesion-org/deepseek-go"
)

// DeepSeekProvider implements the AIProvider interface for DeepSeek
type DeepSeekProvider struct {
	client *deepseek.Client
	config *models.ProviderConfig
}

// New creates a new DeepSeek provider
func New(config *models.ProviderConfig) (*DeepSeekProvider, error) {
	if config.APIKey == "" {
		log.Printf("DeepSeek provider not configured - missing API key")
		return &DeepSeekProvider{config: config}, nil
	}

	client := deepseek.NewClient(config.APIKey)

	return &DeepSeekProvider{
		client: client,
		config: config,
	}, nil
}

// Generate generates content using DeepSeek
func (p *DeepSeekProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	if p.client == nil {
		return "", fmt.Errorf("DeepSeek provider not configured")
	}

	model := p.config.Model
	if model == "" {
		model = "deepseek-chat"
	}

	resp, err := p.client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model: model,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
	})

	if err != nil {
		log.Printf("DeepSeek API error: %v", err)
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no content in DeepSeek response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Name returns the provider name
func (p *DeepSeekProvider) Name() string {
	return "deepseek"
}

// IsConfigured checks if the provider is properly configured
func (p *DeepSeekProvider) IsConfigured() bool {
	return p.client != nil && p.config != nil && p.config.APIKey != ""
}
