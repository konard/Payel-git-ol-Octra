package gemini

import (
	"agents/pkg/models"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the AIProvider interface for Google Gemini
type GeminiProvider struct {
	mu     sync.Mutex
	client *genai.Client
	config *models.ProviderConfig
}

// New creates a new Gemini provider
func New(config *models.ProviderConfig) (*GeminiProvider, error) {
	return &GeminiProvider{config: config}, nil
}

// getClient returns or creates client using API key from tokens
func (p *GeminiProvider) getClient(tokens map[string]interface{}) (*genai.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	apiKey, _ := tokens["gemini"].(string)
	if apiKey == "" {
		// Try string value
		for _, v := range tokens {
			if s, ok := v.(string); ok && len(s) > 10 {
				apiKey = s
				break
			}
		}
	}

	if apiKey == "" && p.config.APIKey != "" {
		apiKey = p.config.APIKey
	}

	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key not found in tokens")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	p.client = client
	return client, nil
}

// Generate generates content using Gemini
func (p *GeminiProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	client, err := p.getClient(tokens)
	if err != nil {
		return "", err
	}

	model := p.config.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	modelClient := client.GenerativeModel(model)

	// Set temperature
	if p.config.Temperature > 0 {
		modelClient.SetTemperature(p.config.Temperature)
	}

	// Set max tokens
	if p.config.MaxTokens > 0 {
		modelClient.SetMaxOutputTokens(int32(p.config.MaxTokens))
	}

	resp, err := modelClient.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Gemini API error: %v", err)
		return "", err
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no content in Gemini response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no parts in Gemini response")
	}

	// Extract text from the first part
	text := fmt.Sprintf("%v", candidate.Content.Parts[0])
	return text, nil
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// IsConfigured always returns true since API key comes from tokens
func (p *GeminiProvider) IsConfigured() bool {
	return true
}

// Close closes the client
func (p *GeminiProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		err := p.client.Close()
		p.client = nil
		return err
	}
	return nil
}
