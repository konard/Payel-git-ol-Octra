// Package llm implements the proxy-mode path: when a user's agent has no CLI
// configured, Octra forwards the prompt straight to the configured LLM using
// the provider's HTTP API.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	providerCatalog "backend/internal/provider"
)

// Client talks to Anthropic Messages and OpenAI-compatible chat APIs.
type Client struct {
	http *http.Client
}

// New returns a Client. If httpClient is nil, http.DefaultClient is used.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{http: httpClient}
}

// Request describes a single completion call.
type Request struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
	Prompt   string
}

type messagesRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []messagePayload `json:"messages"`
}

type messagePayload struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type chatCompletionRequest struct {
	Model     string                  `json:"model"`
	MaxTokens int                     `json:"max_tokens,omitempty"`
	Messages  []chatCompletionMessage `json:"messages"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Complete sends the prompt and returns the assistant's text reply.
func (c *Client) Complete(ctx context.Context, req Request) (string, error) {
	baseURL := strings.TrimRight(req.BaseURL, "/")
	provider := normalizeProvider(req.Provider)
	if baseURL == "" {
		baseURL = providerBaseURL(provider)
	}
	if useAnthropicMessages(provider, baseURL) {
		return c.completeAnthropic(ctx, req, baseURL)
	}
	return c.completeChatCompletions(ctx, req, baseURL, provider)
}

func (c *Client) completeAnthropic(ctx context.Context, req Request, baseURL string) (string, error) {
	model := req.Model
	if model == "" {
		model = providerDefaultModel("claude")
	}
	body, err := json.Marshal(messagesRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages:  []messagePayload{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		return "", err
	}

	url := apiEndpoint(baseURL, "v1/messages")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm responded %d: %s", resp.StatusCode, string(data))
	}

	var parsed messagesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}

	var sb strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}

func (c *Client) completeChatCompletions(ctx context.Context, req Request, baseURL, provider string) (string, error) {
	model := req.Model
	if model == "" {
		model = providerDefaultModel(provider)
	}
	if model == "" {
		model = "gpt-4.1"
	}
	body, err := json.Marshal(chatCompletionRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages:  []chatCompletionMessage{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint(baseURL, "chat/completions"), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if req.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm responded %d: %s", resp.StatusCode, string(data))
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}
	var sb strings.Builder
	for _, choice := range parsed.Choices {
		sb.WriteString(choice.Message.Content)
	}
	return sb.String(), nil
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "anthropic" {
		return "claude"
	}
	return provider
}

func useAnthropicMessages(provider, baseURL string) bool {
	if provider == "claude" {
		return true
	}
	if provider == "" {
		return baseURL == "" || strings.Contains(strings.ToLower(baseURL), "anthropic.com")
	}
	return strings.Contains(strings.ToLower(baseURL), "anthropic.com")
}

func providerBaseURL(provider string) string {
	for _, p := range providerCatalog.BuiltinProviders() {
		if normalizeProvider(p.Key) == provider {
			return strings.TrimRight(p.BaseURL, "/")
		}
	}
	return ""
}

func providerDefaultModel(provider string) string {
	for _, p := range providerCatalog.BuiltinProviders() {
		if normalizeProvider(p.Key) == provider {
			return p.DefaultModel
		}
	}
	if provider == "" {
		return providerDefaultModel("claude")
	}
	return ""
}

func apiEndpoint(baseURL, suffix string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return "/" + suffix
	}
	lower := strings.ToLower(baseURL)
	if strings.HasSuffix(lower, "/messages") || strings.HasSuffix(lower, "/chat/completions") {
		return baseURL
	}
	if strings.HasSuffix(lower, "/v1") && strings.HasPrefix(strings.TrimLeft(suffix, "/"), "v1/") {
		return baseURL + strings.TrimPrefix(strings.TrimLeft(suffix, "/"), "v1")
	}
	return baseURL + "/" + strings.TrimLeft(suffix, "/")
}
