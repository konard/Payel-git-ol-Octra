package claude

import (
	"agents/pkg/models"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements the AIProvider interface for Anthropic Claude
type ClaudeProvider struct {
	mu     sync.Mutex
	client *anthropic.Client
	config *models.ProviderConfig
}

// New creates a new Claude provider
func New(config *models.ProviderConfig) (*ClaudeProvider, error) {
	return &ClaudeProvider{config: config}, nil
}

// getClient returns or creates client using API key from tokens
func (p *ClaudeProvider) getClient(tokens map[string]interface{}) (*anthropic.Client, error) {
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
		return nil, fmt.Errorf("Claude API key not found in tokens")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	p.client = &client
	return p.client, nil
}

func extractAPIKey(tokens map[string]interface{}) string {
	for _, key := range []string{"claude", "anthropic", "api_key", "apiKey", "token"} {
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

// Generate generates content using Claude
func (p *ClaudeProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	maxTokens := 1024
	if p.config.MaxTokens > 0 {
		maxTokens = p.config.MaxTokens
	}

	model := p.config.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: int64(maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})

	if err != nil {
		log.Printf("Claude API error: %v", err)
		return "", err
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	contentBlock := message.Content[0]
	return contentBlock.Text, nil
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// IsConfigured always returns true since API key comes from tokens
func (p *ClaudeProvider) IsConfigured() bool {
	return true
}
