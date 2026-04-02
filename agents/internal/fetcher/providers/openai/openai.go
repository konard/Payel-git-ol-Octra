package openai

import (
	"context"
	"fmt"
	"log"

	"agents/pkg/models"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenAIProvider implements the AIProvider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	config *models.ProviderConfig
}

// New creates a new OpenAI provider
func New(config *models.ProviderConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		log.Printf("OpenAI provider not configured - missing API key")
		return &OpenAIProvider{config: config}, nil
	}

	client := openai.NewClient(option.WithAPIKey(config.APIKey))

	return &OpenAIProvider{
		&client,
		config,
	}, nil
}

// Generate generates content using OpenAI
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	if p.client == nil {
		return "", fmt.Errorf("OpenAI provider not configured")
	}

	model := p.config.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	maxTokens := 1024
	if p.config.MaxTokens > 0 {
		maxTokens = p.config.MaxTokens
	}

	response, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:     model,
		MaxTokens: openai.Int(int64(maxTokens)),
	})

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no content in OpenAI response")
	}

	choice := response.Choices[0]
	if choice.Message.Content == "" {
		return "", fmt.Errorf("empty content in OpenAI response")
	}

	return choice.Message.Content, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// IsConfigured checks if the provider is properly configured
func (p *OpenAIProvider) IsConfigured() bool {
	return p.client != nil && p.config != nil && p.config.APIKey != ""
}
