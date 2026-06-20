package prompts

import (
	"strings"
	"testing"
)

func TestUniversalNode_CodeTask(t *testing.T) {
	p := UniversalNode("hello world", "напиши hello world на python", "code", "python", "")
	mustContain(t, p, "UNIVERSAL NODE")
	mustContain(t, p, "keep it minimal")
	mustContain(t, p, `{"files":`)
	mustContain(t, p, "source-file extension")
	mustContain(t, p, "python")
	if !strings.Contains(p, "solution/answer.md") {
		t.Errorf("code task prompt should mention solution/answer.md as the format for questions")
	}
}

func TestUniversalNode_SearchBlock(t *testing.T) {
	block := "[1] Example title — https://example.com\nSnippet about the topic."
	p := UniversalNode("q", "what is the latest news", "research", "markdown", block)
	mustContain(t, p, "WEB SEARCH RESULTS")
	mustContain(t, p, block)
	mustContain(t, p, "do not invent facts")
}

func TestUniversalNode_NoSearchBlock(t *testing.T) {
	p := UniversalNode("q", "what is 2+2", "research", "markdown", "")
	if strings.Contains(p, "WEB SEARCH RESULTS") {
		t.Errorf("prompt should not contain a web-search section when no results were provided")
	}
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("prompt missing %q", needle)
	}
}
