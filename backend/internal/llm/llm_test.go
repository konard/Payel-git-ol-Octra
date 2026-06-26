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
	var gotPath, gotKey, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-api-key")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"hello "},{"type":"text","text":"world"}]}`)
	}))
	defer srv.Close()

	out, err := New(srv.Client()).Complete(context.Background(), Request{
		APIKey:  "sk-test",
		BaseURL: srv.URL,
		Model:   "claude-sonnet-4-6",
		Prompt:  "hi",
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
	if !strings.Contains(gotBody, `"hi"`) {
		t.Fatalf("prompt not in body: %q", gotBody)
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
