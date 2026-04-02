package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// Config — конфигурация LLM клиента
type Config struct {
	Provider string
	APIKey   string
	BaseURL  string // например: https://openrouter.ai/api/v1 или https://generativelanguage.googleapis.com
	Model    string // например: meta-llama/llama-3.3-70b-instruct:free или gemini-3-flash-preview
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

// NewLLMClient — фабрика для создания нужного LLM клиента
// Определяет тип провайдера по URL и создаёт подходящий клиент
func NewLLMClient(baseURL, model string, tokens []string) Client {
	// Определяем провайдера по URL (проверяем ДО установки дефолтов)
	if strings.Contains(baseURL, "generativelanguage.googleapis.com") ||
		strings.Contains(baseURL, "gemini") ||
		strings.HasSuffix(baseURL, ":generateContent") {
		return NewGeminiClient(baseURL, model, tokens)
	}

	if strings.Contains(baseURL, "openrouter.ai") || baseURL == "" {
		if baseURL == "" {
			baseURL = "https://openrouter.ai/api/v1"
		}
		if model == "" {
			model = "meta-llama/llama-3.3-70b-instruct:free"
		}
		return NewOpenRouterClient(baseURL, model, tokens)
	}

	// По умолчанию используем OpenRouter формат (совместимый с большинством OpenAI-like API)
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct:free"
	}
	return NewOpenRouterClient(baseURL, model, tokens)
}

// NewOpenRouterClient — явное создание OpenRouter клиента
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

// NewGeminiClient — явное создание Gemini клиента
func NewGeminiClient(baseURL, model string, tokens []string) *GeminiClient {
	// Если baseURL это полный путь с :generateContent, оставляем как есть
	if !strings.Contains(baseURL, ":generateContent") {
		if baseURL == "" {
			baseURL = "https://generativelanguage.googleapis.com/v1beta"
		}
	}
	if model == "" {
		model = "gemini-3-flash-preview"
	}

	return &GeminiClient{
		baseURL:      baseURL,
		model:        model,
		httpClient:   &http.Client{},
		tokenRotator: NewTokenRotator(tokens),
	}
}

// ===== OPENROUTER CLIENT =====

// OpenRouterClient — клиент для OpenRouter API (и совместимых с OpenAI API)
type OpenRouterClient struct {
	baseURL      string
	model        string
	httpClient   *http.Client
	tokenRotator *TokenRotator
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

	// Пробуем с разными токенами при ошибке (ротация)
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

// ===== GEMINI CLIENT =====

// GeminiClient — клиент для Google Gemini API
type GeminiClient struct {
	baseURL      string
	model        string
	httpClient   *http.Client
	tokenRotator *TokenRotator
}

type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	return c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (c *GeminiClient) Chat(ctx context.Context, messages []Message) (string, error) {
	var lastErr error

	// Пробуем с разными токенами (ключами) при ошибке
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

func (c *GeminiClient) doRequest(ctx context.Context, apiKey string, messages []Message) (string, error) {
	// Преобразуем сообщения в формат Gemini
	var contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}

	for _, msg := range messages {
		content := struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			Parts: []struct {
				Text string `json:"text"`
			}{
				{Text: msg.Content},
			},
		}
		contents = append(contents, content)
	}

	reqBody := GeminiRequest{
		Contents: contents,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	// Формируем URL для Gemini API - обрабатываем оба варианта
	var url string
	if strings.Contains(c.baseURL, ":generateContent") {
		// Полный URL уже передан (например: https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent)
		url = fmt.Sprintf("%s?key=%s", c.baseURL, apiKey)
	} else {
		// Базовый URL без модели (например: https://generativelanguage.googleapis.com/v1beta)
		url = fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	// Проверяем HTTP статус
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp GeminiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("unmarshal error: %w (response: %s)", err, string(body))
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Candidates) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	if len(apiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in Gemini response")
	}

	return apiResp.Candidates[0].Content.Parts[0].Text, nil
}
