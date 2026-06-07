package boss

import "testing"

// Issue #19 adapts the system for web search. Search-style requests must be
// classified as "research" so the boss → manager → worker flow runs a real
// web search, rather than falling through to the default "code" pipeline.
func TestClassifyTaskType_Search(t *testing.T) {
	cases := []struct {
		name  string
		title string
		desc  string
		want  string
	}{
		{"search the web", "Search the web for httpx install on Python", "", TaskTypeResearch},
		{"web search phrase", "Need a web search about BM25 ranking", "", TaskTypeResearch},
		{"look up", "Look up the latest news on Go 1.26", "", TaskTypeResearch},
		{"russian internet search", "Поиск в интернете про установку httpx", "", TaskTypeResearch},
		{"russian google", "Погугли свежие новости про Go", "", TaskTypeResearch},
		{"explicit research", "Research the market for AI agents", "", TaskTypeResearch},
		{"presentation still wins", "Make a presentation and search the web for facts", "", TaskTypePresentation},
		{"plain code untouched", "Implement a REST API in Go", "", TaskTypeCode},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyTaskType(tc.title, tc.desc); got != tc.want {
				t.Fatalf("classifyTaskType(%q) = %q, want %q", tc.title, got, tc.want)
			}
		})
	}
}

