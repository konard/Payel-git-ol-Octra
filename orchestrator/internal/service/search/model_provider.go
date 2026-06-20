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

// ProgressFunc — callback для поисковых событий. progress — процент выполнения
// (0–100), message — человекочитаемое описание, data — дополнительные ключи.
type ProgressFunc func(progress int32, message string, data map[string]string)

// ModelProvider turns an AI search model response into the same Result shape as
// classic web search providers. Apodex uses the Responses API; custom providers
// can point either to /v1/responses or to an OpenAI-compatible chat completions
// endpoint.
type ModelProvider struct {
	cfg      ModelConfig
	client   *http.Client
	progress ProgressFunc
}

func NewModelProvider(cfg ModelConfig) *ModelProvider {
	return &ModelProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

// SetProgressReporter устанавливает callback для получения промежуточных событий
// поиска (reasoning steps: thinking, web_search, fetch_url_content).
func (p *ModelProvider) SetProgressReporter(fn ProgressFunc) {
	p.progress = fn
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
		text, err = p.readModelStream(resp.Body)
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

func (p *ModelProvider) readModelStream(r io.Reader) (string, error) {
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
		if p.progress != nil {
			p.emitReasoningStep([]byte(data))
		}
		sb.WriteString(extractModelStreamDelta([]byte(data)))
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (p *ModelProvider) emitReasoningStep(data []byte) {
	if p.progress == nil {
		return
	}
	var event struct {
		Type          string                 `json:"type"`
		ReasoningStep map[string]interface{} `json:"reasoning_step"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}
	if event.Type != "reasoning_step" || event.ReasoningStep == nil {
		return
	}

	stepType, _ := event.ReasoningStep["type"].(string)
	switch stepType {
	case "thinking":
		thought, _ := event.ReasoningStep["thought"].(string)
		if thought == "" {
			return
		}
		p.progress(20, "Thinking: "+shorten(thought, 120), map[string]string{
			"search_step":    "thinking",
			"search_thought": thought,
		})

	case "web_search":
		keywords, _ := event.ReasoningStep["search_keywords"].(string)
		if keywords == "" {
			return
		}
		rawResults, _ := event.ReasoningStep["search_results"].([]interface{})
		msg := "Searching: " + keywords
		data := map[string]string{
			"search_step":     "web_search",
			"search_keywords": keywords,
		}
		if len(rawResults) > 0 {
			var urls []string
			for _, r := range rawResults {
				if m, ok := r.(map[string]interface{}); ok {
					if u, ok := m["url"].(string); ok {
						urls = append(urls, u)
					}
				}
			}
			if len(urls) > 0 {
				data["search_urls"] = strings.Join(urls, "\n")
			}
			data["search_result_count"] = fmt.Sprintf("%d", len(rawResults))
			msg = fmt.Sprintf("Search found %d result(s) for: %s", len(rawResults), keywords)
		}
		p.progress(20, msg, data)

	case "fetch_url_content":
		url, _ := event.ReasoningStep["url"].(string)
		if url == "" {
			return
		}
		data := map[string]string{
			"search_step": "fetch_url",
			"search_url":  url,
		}
		if resp, ok := event.ReasoningStep["response"].(map[string]interface{}); ok {
			if code, ok := resp["status_code"].(float64); ok {
				data["search_status"] = fmt.Sprintf("%.0f", code)
			}
		}
		p.progress(20, "Fetching page: "+url, data)

	case "agent_summary":
		summary, _ := event.ReasoningStep["summary"].(string)
		if summary != "" {
			p.progress(20, "Search agent summary: "+shorten(summary, 200), map[string]string{
				"search_step":    "agent_summary",
				"search_summary": summary,
			})
		}
	}
}

func shorten(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
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
