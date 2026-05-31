package search

import (
	"context"
	"strings"
	"testing"
)

// fakeProvider — детерминированная заглушка провайдера для тестов без сети.
type fakeProvider struct {
	byQuery map[string][]Result
	err     error
	calls   []string
}

func (f *fakeProvider) Search(_ context.Context, query string, limit int) ([]Result, error) {
	f.calls = append(f.calls, query)
	if f.err != nil {
		return nil, f.err
	}
	res := f.byQuery[query]
	if limit > 0 && len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}

func TestRankBM25_OrdersByRelevance(t *testing.T) {
	results := []Result{
		{Title: "Cooking pasta", Snippet: "How to boil water for pasta"},
		{Title: "Install httpx on Python", Snippet: "pip install httpx is the way to install the httpx python http client"},
		{Title: "Python news", Snippet: "general python language updates"},
	}
	ranked := RankBM25("install httpx python", results)
	if len(ranked) != 3 {
		t.Fatalf("expected 3 results, got %d", len(ranked))
	}
	if !strings.Contains(ranked[0].Title, "Install httpx") {
		t.Fatalf("expected the httpx doc ranked first, got %q (score %.3f)", ranked[0].Title, ranked[0].Score)
	}
	// Полностью нерелевантный документ должен иметь нулевой счёт.
	last := ranked[len(ranked)-1]
	if last.Score != 0 {
		t.Fatalf("expected the irrelevant doc to have score 0, got %.3f (%q)", last.Score, last.Title)
	}
}

func TestRankBM25_Empty(t *testing.T) {
	if got := RankBM25("anything", nil); got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestClientResearch_DedupesAndRanks(t *testing.T) {
	fp := &fakeProvider{byQuery: map[string][]Result{
		"httpx install": {
			{Title: "Install httpx", URL: "https://python-httpx.org/install/", Snippet: "pip install httpx"},
			{Title: "Dup", URL: "https://dup.example/", Snippet: "duplicate"},
		},
		"httpx python http client": {
			{Title: "Dup again", URL: "https://dup.example", Snippet: "same url, trailing slash differs"},
			{Title: "httpx client", URL: "https://habr.com/httpx", Snippet: "httpx is a python http client"},
		},
	}}
	c := NewClientWithProvider(fp)

	res, err := c.Research(context.Background(), "how to install httpx python http client",
		[]string{"httpx install", "httpx python http client"}, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Два запроса дали 4 результата, но dup.example встречается дважды (с/без
	// завершающего слэша) → ожидаем 3 уникальных.
	if len(res) != 3 {
		t.Fatalf("expected 3 deduped results, got %d", len(res))
	}
	if len(fp.calls) != 2 {
		t.Fatalf("expected 2 provider calls, got %d", len(fp.calls))
	}
}

func TestClientResearch_ProviderErrorPropagates(t *testing.T) {
	fp := &fakeProvider{err: context.DeadlineExceeded}
	c := NewClientWithProvider(fp)
	_, err := c.Research(context.Background(), "topic", []string{"q1"}, 5, 5)
	if err == nil {
		t.Fatal("expected error when provider fails and nothing found")
	}
}

func TestFormatForPrompt(t *testing.T) {
	out := FormatForPrompt([]Result{
		{Title: "T1", URL: "https://a.example", Snippet: "snippet  with   spaces"},
	})
	if !strings.Contains(out, "[1] T1") || !strings.Contains(out, "https://a.example") {
		t.Fatalf("prompt block missing fields: %q", out)
	}
	if strings.Contains(out, "snippet  with") {
		t.Fatalf("expected collapsed spaces in snippet: %q", out)
	}
}

func TestFormatSourcesMarkdown(t *testing.T) {
	out := FormatSourcesMarkdown([]Result{
		{Title: "Title", URL: "https://x.example", Source: "x.example"},
	})
	if !strings.Contains(out, "- [Title](https://x.example) — x.example") {
		t.Fatalf("unexpected markdown: %q", out)
	}
}

func TestEnabled(t *testing.T) {
	t.Setenv("WEB_SEARCH_DISABLED", "1")
	if Enabled() {
		t.Fatal("expected disabled when WEB_SEARCH_DISABLED=1")
	}
	t.Setenv("WEB_SEARCH_DISABLED", "")
	if !Enabled() {
		t.Fatal("expected enabled by default")
	}
}
