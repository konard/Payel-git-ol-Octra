package openrouter

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"agents/pkg/models"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenRouterProvider implements the AIProvider interface for OpenRouter API
type OpenRouterProvider struct {
	mu     sync.Mutex
	client *openai.Client
	config *models.ProviderConfig
}

// New creates a new OpenRouter provider
func New(config *models.ProviderConfig) (*OpenRouterProvider, error) {
	return &OpenRouterProvider{config: config}, nil
}

// getClient returns or creates client using API key from tokens
func (p *OpenRouterProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
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
		return nil, fmt.Errorf("OpenRouter API key not found in tokens")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)
	p.client = &client
	return p.client, nil
}

func extractAPIKey(tokens map[string]interface{}) string {
	for _, key := range []string{"openrouter", "api_key", "apiKey", "token"} {
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

// Generate generates content using OpenRouter API
func (p *OpenRouterProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "openai/gpt-4o-mini" // default fallback
	}

	maxTokens := 1024
	if p.config.MaxTokens > 0 {
		maxTokens = p.config.MaxTokens
	}

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:     model,
		MaxTokens: openai.Int(int64(maxTokens)),
	}

	// Retry logic for transient errors (EOF, connection reset, etc.)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 2 * time.Second
			log.Printf("OpenRouter retry %d/%3 after %v...", attempt, delay)
			time.Sleep(delay)
		}

		response, err := client.Chat.Completions.New(ctx, params)
		if err == nil {
			if len(response.Choices) == 0 {
				return "", fmt.Errorf("no content in OpenRouter response")
			}
			choice := response.Choices[0]
			if choice.Message.Content == "" {
				return "", fmt.Errorf("empty content in OpenRouter response")
			}
			return choice.Message.Content, nil
		}

		lastErr = err
		errMsg := err.Error()

		// Retry on transient errors
		if strings.Contains(errMsg, "EOF") ||
			strings.Contains(errMsg, "connection reset") ||
			strings.Contains(errMsg, "unexpected") ||
			strings.Contains(errMsg, "timeout") {
			log.Printf("OpenRouter transient error (attempt %d): %v", attempt+1, err)
			continue
		}

		// Don't retry on auth/rate limit/errors
		log.Printf("OpenRouter API error: %v", err)
		return "", fmt.Errorf("OpenRouter API error: %w", err)
	}

	return "", fmt.Errorf("OpenRouter failed after 3 attempts: %w", lastErr)
}

// Name returns the provider name
func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

// IsConfigured always returns true since API key comes from tokens
func (p *OpenRouterProvider) IsConfigured() bool {
	return true
}
