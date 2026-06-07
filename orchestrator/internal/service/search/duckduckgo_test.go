package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleDDGHTML = `<!DOCTYPE html><html><body>
<div class="result results_links">
  <div class="result__body">
    <a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fwww.python-httpx.org%2F&rut=abc">HTTPX official site</a>
    <a class="result__snippet">HTTPX is a fully featured HTTP client for Python 3.</a>
  </div>
</div>
<div class="result results_links">
  <div class="result__body">
    <a class="result__a" href="https://habr.com/ru/articles/httpx/">Установка httpx</a>
    <a class="result__snippet">pip install httpx — самый простой способ.</a>
  </div>
</div>
</body></html>`

func TestParseDuckDuckGoHTML(t *testing.T) {
	results := parseDuckDuckGoHTML(strings.NewReader(sampleDDGHTML))
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	first := results[0]
	if first.Title != "HTTPX official site" {
		t.Errorf("unexpected title: %q", first.Title)
	}
	if first.URL != "https://www.python-httpx.org/" {
		t.Errorf("uddg redirect not decoded, got %q", first.URL)
	}
	if first.Source != "python-httpx.org" {
		t.Errorf("unexpected domain: %q", first.Source)
	}
	if !strings.Contains(first.Snippet, "HTTP client for Python") {
		t.Errorf("unexpected snippet: %q", first.Snippet)
	}

	second := results[1]
	if second.URL != "https://habr.com/ru/articles/httpx/" {
		t.Errorf("direct href not preserved, got %q", second.URL)
	}
	if second.Source != "habr.com" {
		t.Errorf("unexpected domain: %q", second.Source)
	}
}

func TestDuckDuckGoProvider_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") == "" {
			t.Error("expected q query parameter")
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(sampleDDGHTML))
	}))
	defer srv.Close()

	p := NewDuckDuckGoProvider()
	p.endpoint = srv.URL + "/html/"

	results, err := p.Search(context.Background(), "install httpx python", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected limit to cap at 1 result, got %d", len(results))
	}
}

func TestDuckDuckGoProvider_EmptyQuery(t *testing.T) {
	p := NewDuckDuckGoProvider()
	if _, err := p.Search(context.Background(), "  ", 5); err == nil {
		t.Fatal("expected error for empty query")
	}
}

