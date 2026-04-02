package claude

import (
	"agents/pkg/models"
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements the AIProvider interface for Anthropic Claude
type ClaudeProvider struct {
	client *anthropic.Client
	config *models.ProviderConfig
}

// New creates a new Claude provider
func New(config *models.ProviderConfig) (*ClaudeProvider, error) {
	if config.APIKey == "" {
		log.Printf("Claude provider not configured - missing API key")
		return &ClaudeProvider{config: config}, nil
	}

	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))

	return &ClaudeProvider{
		client: &client,
		config: config,
	}, nil
}

// Generate generates content using Claude
func (p *ClaudeProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	if p.client == nil {
		return "", fmt.Errorf("Claude provider not configured")
	}

	maxTokens := 1024
	if p.config.MaxTokens > 0 {
		maxTokens = p.config.MaxTokens
	}

	model := p.config.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
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

	// The first content block should have the text
	// In anthropic SDK, ContentBlockUnion can directly contain Text field
	contentBlock := message.Content[0]

	// Try to access Text directly (union type flattening)
	return contentBlock.Text, nil
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// IsConfigured checks if the provider is properly configured
func (p *ClaudeProvider) IsConfigured() bool {
	return p.client != nil && p.config != nil && p.config.APIKey != ""
}
