package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"orchestrator/internal/service/search"
)

func TestResearchUsesModelProviderWhenWebSearchDisabled(t *testing.T) {
	t.Setenv("WEB_SEARCH_DISABLED", "1")

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Fatalf("Authorization header = %q, want Bearer sk-test", got)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got := payload["model"]; got != "search-model" {
			t.Fatalf("model = %v, want search-model", got)
		}
		if _, ok := payload["messages"].([]interface{}); !ok {
			t.Fatalf("expected OpenAI-compatible chat messages payload, got %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"model answer"}}]}`))
	}))
	defer server.Close()

	svc := &Service{}
	results, err := svc.research(context.Background(), "topic", []string{"topic search"}, &search.ModelConfig{
		Provider: "custom",
		Model:    "search-model",
		BaseURL:  server.URL + "/v1",
		APIKey:   "sk-test",
	})
	if err != nil {
		t.Fatalf("research returned error: %v", err)
	}
	if !called {
		t.Fatal("expected model provider to be called")
	}
	if len(results) != 1 {
		t.Fatalf("results len = %d, want 1", len(results))
	}
	if got := results[0].Snippet; got != "model answer" {
		t.Fatalf("snippet = %q, want model answer", got)
	}
}
