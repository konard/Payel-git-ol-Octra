// Package llm implements the proxy-mode path: when a user's agent has no CLI
// configured, Octra forwards the prompt straight to the configured LLM using
// the Anthropic Messages API shape.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client talks to an Anthropic-compatible Messages API.
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
	APIKey  string
	BaseURL string
	Model   string
	Prompt  string
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

// Complete sends the prompt and returns the assistant's text reply.
func (c *Client) Complete(ctx context.Context, req Request) (string, error) {
	model := req.Model
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	body, err := json.Marshal(messagesRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages:  []messagePayload{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(req.BaseURL, "/") + "/v1/messages"
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
