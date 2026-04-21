package custom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"agents/pkg/models"
)

// CustomProvider implements the AIProvider interface for custom AI providers
type CustomProvider struct {
	mu     sync.Mutex
	client *http.Client
	config *models.ProviderConfig
}

// New creates a new custom provider
func New(config *models.ProviderConfig) (*CustomProvider, error) {
	return &CustomProvider{
		config: config,
		client: &http.Client{
			Timeout: 0, // No timeout for long-running requests
		},
	}, nil
}

// Generate generates content using custom provider
func (p *CustomProvider) Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error) {
	// Extract API key from tokens
	apiKey := extractAPIKey(tokens)
	if apiKey == "" && p.config.APIKey != "" {
		apiKey = p.config.APIKey
	}
	if apiKey == "" {
		return "", fmt.Errorf("Custom provider API key not found")
	}

	// Prepare request body (OpenAI-compatible format)
	// Model, max_tokens and temperature will be provided by the main application
	requestBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Get base URL from config Tokens
	baseURL := ""
	if p.config.Tokens != nil {
		if urlVal, ok := p.config.Tokens["base_url"].(string); ok {
			baseURL = urlVal
		}
	}
	if baseURL == "" {
		return "", fmt.Errorf("Custom provider base URL not configured")
	}

	url := fmt.Sprintf("%s/chat/completions", baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Custom provider API error: %s", string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (OpenAI-compatible format)
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Choices[0].Message.Content, nil
}

// Name returns the provider name
func (p *CustomProvider) Name() string {
	return "custom"
}

// IsConfigured always returns true since configuration comes from frontend
func (p *CustomProvider) IsConfigured() bool {
	return true
}

func extractAPIKey(tokens map[string]interface{}) string {
	// Try common key names
	for _, key := range []string{"api_key", "apiKey", "token", "key"} {
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
