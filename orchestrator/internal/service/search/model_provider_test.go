package search

import (
	"strings"
	"testing"
)

func TestModelEndpoint(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ModelConfig
		wantEndpoint string
		wantStyle    string
	}{
		{
			name:         "apodex root",
			cfg:          ModelConfig{Provider: "apodex", BaseURL: "https://api.apodex.ai"},
			wantEndpoint: "https://api.apodex.ai/v1/responses",
			wantStyle:    "responses",
		},
		{
			name:         "explicit responses",
			cfg:          ModelConfig{Provider: "custom", BaseURL: "https://search.example.com/v1/responses"},
			wantEndpoint: "https://search.example.com/v1/responses",
			wantStyle:    "responses",
		},
		{
			name:         "custom v1",
			cfg:          ModelConfig{Provider: "custom", BaseURL: "https://search.example.com/v1"},
			wantEndpoint: "https://search.example.com/v1/chat/completions",
			wantStyle:    "chat",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotEndpoint, gotStyle := modelEndpoint(tc.cfg)
			if gotEndpoint != tc.wantEndpoint || gotStyle != tc.wantStyle {
				t.Fatalf("modelEndpoint() = (%q, %q), want (%q, %q)", gotEndpoint, gotStyle, tc.wantEndpoint, tc.wantStyle)
			}
		})
	}
}

func TestExtractModelResponseText(t *testing.T) {
	if got := extractModelResponseText([]byte(`{"output_text":"apodex answer"}`)); got != "apodex answer" {
		t.Fatalf("output_text = %q", got)
	}
	if got := extractModelResponseText([]byte(`{"choices":[{"message":{"content":"chat answer"}}]}`)); got != "chat answer" {
		t.Fatalf("chat content = %q", got)
	}
	if got := extractModelResponseText([]byte(`{"output":[{"content":[{"text":"responses answer"}]}]}`)); got != "responses answer" {
		t.Fatalf("responses output content = %q", got)
	}
}

func TestReadModelStream(t *testing.T) {
	p := NewModelProvider(ModelConfig{Streaming: true})
	stream := strings.NewReader("event: response.output_text.delta\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\" world\"}\n\n" +
		"data: [DONE]\n\n")

	got, err := p.readModelStream(stream)
	if err != nil {
		t.Fatalf("readModelStream returned error: %v", err)
	}
	if got != "hello world" {
		t.Fatalf("readModelStream = %q, want %q", got, "hello world")
	}
}

func TestReadModelStreamWithReasoning(t *testing.T) {
	var events []string
	p := NewModelProvider(ModelConfig{Streaming: true})
	p.SetProgressReporter(func(_ int32, msg string, data map[string]string) {
		if step := data["search_step"]; step != "" {
			events = append(events, step+": "+msg)
		}
	})
	stream := strings.NewReader(
		"data: {\"type\":\"reasoning_step\",\"reasoning_step\":{\"type\":\"thinking\",\"thought\":\"I need to search for this\"}}\n\n" +
			"data: {\"type\":\"reasoning_step\",\"reasoning_step\":{\"type\":\"web_search\",\"search_keywords\":\"test query\",\"search_results\":[{\"title\":\"Result 1\",\"url\":\"https://example.com\"}]}}\n\n" +
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"answer text\"}\n\n" +
			"data: [DONE]\n\n")

	got, err := p.readModelStream(stream)
	if err != nil {
		t.Fatalf("readModelStream returned error: %v", err)
	}
	if got != "answer text" {
		t.Fatalf("readModelStream = %q, want %q", got, "answer text")
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 progress events, got %d: %v", len(events), events)
	}
}
