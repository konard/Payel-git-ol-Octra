package gemini

import (
	"agents/pkg/models"
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the AIProvider interface for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	config *models.ProviderConfig
}

// New creates a new Gemini provider
func New(config *models.ProviderConfig) (*GeminiProvider, error) {
	if config.APIKey == "" {
		log.Printf("Gemini provider not configured - missing API key")
		return &GeminiProvider{config: config}, nil
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return nil, err
	}

	return &GeminiProvider{
		client: client,
		config: config,
	}, nil
}

// Generate generates content using Gemini
func (p *GeminiProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	if p.client == nil {
		return "", fmt.Errorf("Gemini provider not configured")
	}

	model := p.config.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	modelClient := p.client.GenerativeModel(model)

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

// IsConfigured checks if the provider is properly configured
func (p *GeminiProvider) IsConfigured() bool {
	return p.client != nil && p.config != nil && p.config.APIKey != ""
}

// Close closes the client
func (p *GeminiProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
