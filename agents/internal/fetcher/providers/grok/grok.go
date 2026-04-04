package grok

import (
	"context"
	"fmt"
	"log"
	"sync"

	"agents/pkg/models"

	grok "github.com/SimonMorphy/grok-go"
)

// GrokProvider implements the AIProvider interface for Grok (via xAI)
type GrokProvider struct {
	mu     sync.Mutex
	client *grok.Client
	config *models.ProviderConfig
}

// New creates a new Grok provider
func New(config *models.ProviderConfig) (*GrokProvider, error) {
	return &GrokProvider{config: config}, nil
}

// getClient returns or creates client using API key from tokens
func (p *GrokProvider) getClient(tokens map[string]interface{}) (*grok.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	apiKey := extractAPIKey(tokens)
	if apiKey == "" && p.config.APIKey != "" {
		apiKey = p.config.APIKey
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Grok API key not found in tokens")
	}

	client, err := grok.NewClient(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Grok client: %w", err)
	}
	p.client = client
	return client, nil
}

func extractAPIKey(tokens map[string]interface{}) string {
	for _, key := range []string{"grok", "xai", "api_key", "apiKey", "token"} {
		if v, ok := tokens[key]; ok {
			if s, ok := v.(string); ok && len(s) > 10 {
				return s
			}
		}
	}
	for _, v := range tokens {
		if s, ok := v.(string); ok && len(s) > 10 {
			return s
		}
	}
	return ""
}

// Generate generates content using Grok
func (p *GrokProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "grok-2"
	}

	resp, err := grok.CreateChatCompletion(ctx, client, &grok.ChatCompletionRequest{
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

// IsConfigured always returns true since API key comes from tokens
func (p *GrokProvider) IsConfigured() bool {
	return true
}
