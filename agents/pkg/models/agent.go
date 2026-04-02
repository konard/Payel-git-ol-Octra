package models

import "context"

// AIProvider defines the interface for different AI providers
type AIProvider interface {
	// Generate generates content based on the prompt
	Generate(ctx context.Context, prompt string, tokens map[string]interface{}) (string, error)

	// Name returns the provider name
	Name() string

	// IsConfigured checks if the provider is properly configured
	IsConfigured() bool
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	Model       string                 `json:"model"`
	APIKey      string                 `json:"api_key"`
	Tokens      map[string]interface{} `json:"tokens"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float32                `json:"temperature"`
}

// AgentRequest is a request to generate content
type AgentRequest struct {
	Provider    string                 `json:"provider"`    // claude, gemini, deepseek, openai, grok, llama
	Model       string                 `json:"model"`       // specific model name
	Prompt      string                 `json:"prompt"`      // the prompt
	Tokens      map[string]interface{} `json:"tokens"`      // authentication tokens
	MaxTokens   int                    `json:"max_tokens"`  // response max tokens
	Temperature float32                `json:"temperature"` // creativity parameter
}

// AgentResponse is the response from an agent
type AgentResponse struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Content   string `json:"content"`
	Tokens    int    `json:"tokens"` // tokens used
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}
