package deepseek

import (
	"agents/pkg/models"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cohesion-org/deepseek-go"
)

// DeepSeekProvider implements the AIProvider interface for DeepSeek
type DeepSeekProvider struct {
	mu     sync.Mutex
	client *deepseek.Client
	config *models.ProviderConfig
}

// New creates a new DeepSeek provider
func New(config *models.ProviderConfig) (*DeepSeekProvider, error) {
	return &DeepSeekProvider{config: config}, nil
}

// getClient returns or creates client using API key from tokens
func (p *DeepSeekProvider) getClient(tokens map[string]interface{}) (*deepseek.Client, error) {
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
		return nil, fmt.Errorf("DeepSeek API key not found in tokens")
	}

	client := deepseek.NewClient(apiKey)
	p.client = client
	return client, nil
}

func extractAPIKey(tokens map[string]interface{}) string {
	for _, key := range []string{"deepseek", "api_key", "apiKey", "token"} {
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

// Generate generates content using DeepSeek
func (p *DeepSeekProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "deepseek-chat"
	}

	resp, err := client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
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

// IsConfigured always returns true since API key comes from tokens
func (p *DeepSeekProvider) IsConfigured() bool {
	return true
}
