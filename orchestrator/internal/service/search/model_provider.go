package search

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ModelConfig struct {
	Provider  string
	Model     string
	BaseURL   string
	APIKey    string
	Streaming bool
}

// ModelProvider turns an AI search model response into the same Result shape as
// classic web search providers. Apodex uses the Responses API; custom providers
// can point either to /v1/responses or to an OpenAI-compatible chat completions
// endpoint.
type ModelProvider struct {
	cfg    ModelConfig
	client *http.Client
}

func NewModelProvider(cfg ModelConfig) *ModelProvider {
	return &ModelProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 45 * time.Second},
	}
}

func (p *ModelProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}
	if p == nil || strings.TrimSpace(p.cfg.BaseURL) == "" || strings.TrimSpace(p.cfg.APIKey) == "" || strings.TrimSpace(p.cfg.Model) == "" {
		return nil, fmt.Errorf("search model provider is not configured")
	}

	endpoint, style := modelEndpoint(p.cfg)
	body, err := json.Marshal(modelRequestPayload(p.cfg, style, query))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(p.cfg.APIKey))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("search model returned status %d: %s", resp.StatusCode, collapseSpaces(string(msg)))
	}

	var text string
	if p.cfg.Streaming {
		text, err = readModelStream(resp.Body)
	} else {
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if readErr != nil {
			err = readErr
		} else {
			text = extractModelResponseText(data)
		}
	}
	if err != nil {
		return nil, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("search model returned an empty response")
	}

	provider := strings.TrimSpace(p.cfg.Provider)
	if provider == "" {
		provider = "custom"
	}
	title := "AI search result: " + query
	if len(title) > 120 {
		title = title[:120]
	}

	return []Result{{
		Title:   title,
		URL:     strings.TrimSpace(p.cfg.BaseURL),
		Snippet: text,
		Source:  provider,
	}}, nil
}

func modelEndpoint(cfg ModelConfig) (endpoint, style string) {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	lower := strings.ToLower(base)
	switch {
	case strings.Contains(lower, "/responses"):
		return base, "responses"
	case strings.Contains(lower, "/chat/completions"):
		return base, "chat"
	case strings.EqualFold(strings.TrimSpace(cfg.Provider), "apodex") || strings.Contains(lower, "apodex"):
		if strings.HasSuffix(lower, "/v1") {
			return base + "/responses", "responses"
		}
		return base + "/v1/responses", "responses"
	case strings.HasSuffix(lower, "/v1"):
		return base + "/chat/completions", "chat"
	default:
		return base + "/v1/chat/completions", "chat"
	}
}

func modelRequestPayload(cfg ModelConfig, style, query string) map[string]interface{} {
	payload := map[string]interface{}{
		"model": strings.TrimSpace(cfg.Model),
	}
	if cfg.Streaming {
		payload["stream"] = true
	}
	if style == "responses" {
		payload["input"] = query
		return payload
	}
	payload["messages"] = []map[string]string{{"role": "user", "content": query}}
	return payload
}

func readModelStream(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var sb strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		sb.WriteString(extractModelStreamDelta([]byte(data)))
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func extractModelStreamDelta(data []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	if choices, ok := payload["choices"].([]interface{}); ok {
		for _, choice := range choices {
			if c, ok := choice.(map[string]interface{}); ok {
				if delta, ok := c["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						return content
					}
				}
			}
		}
	}
	if delta, ok := payload["delta"].(string); ok {
		return delta
	}
	if outputText, ok := payload["output_text"].(string); ok {
		return outputText
	}
	if text, ok := payload["text"].(string); ok {
		return text
	}
	return extractModelResponseText(data)
}

func extractModelResponseText(data []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	if outputText, ok := payload["output_text"].(string); ok {
		return outputText
	}
	if text := extractChoicesText(payload); text != "" {
		return text
	}
	if text := extractResponsesOutputText(payload); text != "" {
		return text
	}
	if text, ok := payload["text"].(string); ok {
		return text
	}
	return ""
}

func extractChoicesText(payload map[string]interface{}) string {
	choices, ok := payload["choices"].([]interface{})
	if !ok {
		return ""
	}
	var parts []string
	for _, choice := range choices {
		c, ok := choice.(map[string]interface{})
		if !ok {
			continue
		}
		if message, ok := c["message"].(map[string]interface{}); ok {
			if content, ok := message["content"].(string); ok {
				parts = append(parts, content)
			}
		}
		if delta, ok := c["delta"].(map[string]interface{}); ok {
			if content, ok := delta["content"].(string); ok {
				parts = append(parts, content)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func extractResponsesOutputText(payload map[string]interface{}) string {
	output, ok := payload["output"].([]interface{})
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range output {
		outputItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := outputItem["content"].([]interface{})
		if !ok {
			continue
		}
		for _, contentItem := range content {
			contentMap, ok := contentItem.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := contentMap["text"].(string); ok {
				parts = append(parts, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
