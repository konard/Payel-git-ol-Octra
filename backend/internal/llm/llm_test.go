package llm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComplete(t *testing.T) {
	var gotPath, gotKey, gotVersion, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"hello "},{"type":"text","text":"world"}]}`)
	}))
	defer srv.Close()

	out, err := New(srv.Client()).Complete(context.Background(), Request{
		Provider: "claude",
		APIKey:   "sk-test",
		BaseURL:  srv.URL,
		Model:    "claude-sonnet-4-6",
		Prompt:   "hi",
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if out != "hello world" {
		t.Fatalf("unexpected output %q", out)
	}
	if gotPath != "/v1/messages" {
		t.Fatalf("unexpected path %q", gotPath)
	}
	if gotKey != "sk-test" {
		t.Fatalf("api key not forwarded: %q", gotKey)
	}
	if gotVersion == "" {
		t.Fatal("anthropic-version header was not set")
	}
	if !strings.Contains(gotBody, `"hi"`) {
		t.Fatalf("prompt not in body: %q", gotBody)
	}
}

func TestCompleteOpenAICompatibleProvider(t *testing.T) {
	var gotPath, gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"openai-compatible reply"}}]}`)
	}))
	defer srv.Close()

	out, err := New(srv.Client()).Complete(context.Background(), Request{
		Provider: "openai",
		APIKey:   "sk-openai",
		BaseURL:  srv.URL + "/v1",
		Model:    "gpt-test",
		Prompt:   "route this",
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if out != "openai-compatible reply" {
		t.Fatalf("unexpected output %q", out)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("unexpected path %q", gotPath)
	}
	if gotAuth != "Bearer sk-openai" {
		t.Fatalf("authorization not forwarded: %q", gotAuth)
	}
	if !strings.Contains(gotBody, `"route this"`) {
		t.Fatalf("prompt not in body: %q", gotBody)
	}
	if !strings.Contains(gotBody, `"model":"gpt-test"`) {
		t.Fatalf("model not in body: %q", gotBody)
	}
}

func TestAPIEndpointDoesNotDuplicateVersionPath(t *testing.T) {
	got := apiEndpoint("https://api.anthropic.com/v1", "v1/messages")
	if got != "https://api.anthropic.com/v1/messages" {
		t.Fatalf("apiEndpoint duplicated version path: %q", got)
	}
}

func TestCompleteErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"bad key"}`)
	}))
	defer srv.Close()

	_, err := New(srv.Client()).Complete(context.Background(), Request{BaseURL: srv.URL, Prompt: "x"})
	if err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}
