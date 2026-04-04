package openai

import (
	"context"
	"fmt"
	"log"
	"sync"

	"agents/pkg/models"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenAIProvider implements the AIProvider interface for OpenAI
type OpenAIProvider struct {
	mu     sync.Mutex
	client *openai.Client
	config *models.ProviderConfig
}

// New creates a new OpenAI provider
func New(config *models.ProviderConfig) (*OpenAIProvider, error) {
	return &OpenAIProvider{config: config}, nil
}

// getClient returns or creates client using API key from tokens
func (p *OpenAIProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
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
		return nil, fmt.Errorf("OpenAI API key not found in tokens")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	p.client = &client
	return p.client, nil
}

func extractAPIKey(tokens map[string]interface{}) string {
	// Try common key names
	for _, key := range []string{"openai", "api_key", "apiKey", "token"} {
		if v, ok := tokens[key]; ok {
			if s, ok := v.(string); ok && len(s) > 10 {
				return s
			}
		}
	}
	// Fallback: first string value that looks like a key
	for _, v := range tokens {
		if s, ok := v.(string); ok && len(s) > 10 {
			return s
		}
	}
	return ""
}

// Generate generates content using OpenAI
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	maxTokens := 1024
	if p.config.MaxTokens > 0 {
		maxTokens = p.config.MaxTokens
	}

	response, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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

// IsConfigured always returns true since API key comes from tokens
func (p *OpenAIProvider) IsConfigured() bool {
	return true
}
