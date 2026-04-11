package zai

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

// ZAIProvider implements the AIProvider interface for Z.AI API
type ZAIProvider struct {
	mu     sync.Mutex
	client *openai.Client
	config *models.ProviderConfig
}

// New creates a new Z.AI provider
func New(config *models.ProviderConfig) (*ZAIProvider, error) {
	return &ZAIProvider{config: config}, nil
}

// extractAPIKey extracts API key from tokens
func extractAPIKey(tokens map[string]interface{}) string {
	for _, key := range []string{"zai", "z.ai", "api_key", "apiKey", "token"} {
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

// getClient returns or creates client using Z.AI API endpoint
func (p *ZAIProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
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
		return nil, fmt.Errorf("Z.AI API key not found in tokens")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.z.ai/api/v1"),
	)
	p.client = &client
	return p.client, nil
}

// Generate generates content using Z.AI API
func (p *ZAIProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "glm-4.5-air"
	}

	maxTokens := 4096
	if p.config.MaxTokens > 0 {
		maxTokens = p.config.MaxTokens
	}

	temperature := 0.5
	if p.config.Temperature > 0 {
		temperature = float64(p.config.Temperature)
	}

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model:       model,
		MaxTokens:   openai.Int(int64(maxTokens)),
		Temperature: openai.Float(temperature),
	}

	// Retry logic for transient errors
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 2 * time.Second
			log.Printf("Z.AI retry %d/%d after %v...", attempt, 3, delay)
			time.Sleep(delay)
		}

		response, err := client.Chat.Completions.New(ctx, params)
		if err == nil {
			if len(response.Choices) == 0 {
				return "", fmt.Errorf("no content in Z.AI response")
			}
			choice := response.Choices[0]
			if choice.Message.Content == "" {
				return "", fmt.Errorf("empty content in Z.AI response")
			}
			return choice.Message.Content, nil
		}

		lastErr = err
		errMsg := err.Error()

		// Retry on transient errors
		if strings.Contains(errMsg, "EOF") ||
			strings.Contains(errMsg, "connection reset") ||
			strings.Contains(errMsg, "unexpected") ||
			strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "500") ||
			strings.Contains(errMsg, "502") ||
			strings.Contains(errMsg, "503") {
			log.Printf("Z.AI transient error (attempt %d): %v", attempt+1, err)
			continue
		}

		// Don't retry on auth/rate limit errors
		log.Printf("Z.AI API error: %v", err)
		return "", fmt.Errorf("Z.AI API error: %w", err)
	}

	return "", fmt.Errorf("Z.AI failed after 3 attempts: %w", lastErr)
}

// Name returns the provider name
func (p *ZAIProvider) Name() string {
	return "zai"
}

// IsConfigured always returns true since API key comes from tokens
func (p *ZAIProvider) IsConfigured() bool {
	return true
}
