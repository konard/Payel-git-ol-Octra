package qwen

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"agents/pkg/models"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// CLIProxyProvider implements the AIProvider interface for CLIProxyAPI (Qwen Code via OAuth)
type CLIProxyProvider struct {
	mu     sync.Mutex
	client *openai.Client
	config *models.ProviderConfig
}

// New creates a new CLIProxy provider for Qwen Code
func New(config *models.ProviderConfig) (*CLIProxyProvider, error) {
	return &CLIProxyProvider{config: config}, nil
}

// getClient returns or creates client using CLIProxyAPI base URL
func (p *CLIProxyProvider) getClient(tokens map[string]interface{}) (*openai.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	// Get CLIProxyAPI URL from environment or use default
	baseURL := os.Getenv("CLIPROXY_API_URL")
	if baseURL == "" {
		baseURL = "http://host.docker.internal:8317/v1"
	}

	// API key from config (CLIProxyAPI auth key, not LLM key)
	apiKey := p.config.APIKey
	if apiKey == "" {
		apiKey = "cliproxy-dev-key-change-me"
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	p.client = &client
	return p.client, nil
}

// Generate generates content using CLIProxyAPI (Qwen Code via OAuth)
func (p *CLIProxyProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "qwen-code" // default alias configured in CLIProxyAPI
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

	// Retry logic with cooldown handling
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 5 * time.Second
			log.Printf("CLIProxy (Qwen) retry %d/%d after %v...", attempt, 3, delay)
			time.Sleep(delay)
		}

		response, err := client.Chat.Completions.New(ctx, params)
		if err == nil {
			if len(response.Choices) == 0 {
				return "", fmt.Errorf("no content in CLIProxy response")
			}
			choice := response.Choices[0]
			if choice.Message.Content == "" {
				return "", fmt.Errorf("empty content in CLIProxy response")
			}
			return choice.Message.Content, nil
		}

		lastErr = err
		errMsg := err.Error()

		// Check for cooldown (429)
		if strings.Contains(errMsg, "429") || strings.Contains(errMsg, "Too Many Requests") {
			// Extract reset_seconds if available
			if strings.Contains(errMsg, "reset_seconds") {
				log.Printf("CLIProxy cooldown detected: %v", err)
				return "", fmt.Errorf("CLIProxy cooldown for model %s: %w", model, err)
			}
			// Generic rate limit — wait and retry
			delay := 30 * time.Second
			log.Printf("CLIProxy rate limit, waiting %v...", delay)
			time.Sleep(delay)
			continue
		}

		// Retry on transient errors
		if strings.Contains(errMsg, "EOF") ||
			strings.Contains(errMsg, "connection reset") ||
			strings.Contains(errMsg, "unexpected") ||
			strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "connection refused") {
			log.Printf("CLIProxy transient error (attempt %d): %v", attempt+1, err)
			continue
		}

		// Don't retry on other errors
		log.Printf("CLIProxy API error: %v", err)
		return "", fmt.Errorf("CLIProxy API error: %w", err)
	}

	return "", fmt.Errorf("CLIProxy failed after 3 attempts: %w", lastErr)
}

// Name returns the provider name
func (p *CLIProxyProvider) Name() string {
	return "cliproxy"
}

// IsConfigured always returns true since CLIProxyAPI uses OAuth
func (p *CLIProxyProvider) IsConfigured() bool {
	return true
}
