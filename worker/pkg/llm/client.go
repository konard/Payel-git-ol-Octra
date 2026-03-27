package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Config — конфигурация LLM клиента
type Config struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// Client — универсальный LLM клиент
type Client interface {
	Generate(ctx context.Context, prompt string) (string, error)
	Chat(ctx context.Context, messages []Message) (string, error)
}

// Message — сообщение для чата
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TokenRotator — ротатор токенов
type TokenRotator struct {
	tokens []string
	index  int
	mu     sync.Mutex
}

func NewTokenRotator(tokens []string) *TokenRotator {
	return &TokenRotator{
		tokens: tokens,
		index:  0,
	}
}

func (r *TokenRotator) Next() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.tokens) == 0 {
		return ""
	}

	token := r.tokens[r.index]
	r.index = (r.index + 1) % len(r.tokens)
	return token
}

// OpenRouterClient — клиент для OpenRouter API (универсальный)
type OpenRouterClient struct {
	baseURL      string
	model        string
	httpClient   *http.Client
	tokenRotator *TokenRotator
}

func NewOpenRouterClient(baseURL, model string, tokens []string) *OpenRouterClient {
	// Дефолтные значения если не указаны
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct:free"
	}

	return &OpenRouterClient{
		baseURL:      baseURL,
		model:        model,
		httpClient:   &http.Client{},
		tokenRotator: NewTokenRotator(tokens),
	}
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenRouterClient) Generate(ctx context.Context, prompt string) (string, error) {
	return c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (c *OpenRouterClient) Chat(ctx context.Context, messages []Message) (string, error) {
	var lastErr error

	for i := 0; i < 3; i++ {
		token := c.tokenRotator.Next()
		if token == "" {
			break
		}

		result, err := c.doRequest(ctx, token, messages)
		if err == nil {
			return result, nil
		}

		lastErr = err
	}

	return "", lastErr
}

func (c *OpenRouterClient) doRequest(ctx context.Context, token string, messages []Message) (string, error) {
	reqBody := OpenRouterRequest{
		Model:    c.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	var apiResp OpenRouterResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("unmarshal error: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return apiResp.Choices[0].Message.Content, nil
}
