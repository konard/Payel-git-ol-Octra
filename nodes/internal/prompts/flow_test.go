package prompts

import (
	"strings"
	"testing"
)

// TestManagerThinkRoutesByTaskType — менеджерский промпт найма воркеров должен
// различаться по типу задачи: код → разработчики, ресёрч → аналитики с углами,
// документ → авторы, презентация → один дизайнер.
func TestManagerThinkRoutesByTaskType(t *testing.T) {
	cases := []struct {
		taskType string
		mustHave []string
	}{
		{"code", []string{"developer"}},
		{"research", []string{"analyst", "angle"}},
		{"document", []string{"writer", "document"}},
		{"presentation", []string{"designer", "exactly 1"}},
	}
	for _, c := range cases {
		got := ManagerThink("lead", "focus", "do the task", "50", c.taskType)
		for _, sub := range c.mustHave {
			if !contains(got, sub) {
				t.Errorf("ManagerThink(%q) missing %q\n---\n%s", c.taskType, sub, got)
			}
		}
	}
}

// TestManagerReviewRoutesByTaskType — ревью кода говорит про code quality/bugs,
// а ревью документов — про факты/структуру/placeholders, а не про код.
func TestManagerReviewRoutesByTaskType(t *testing.T) {
	code := ManagerReviewWork("lead", "dev", "task", "sol", "files", "code")
	if !contains(code, "code quality") {
		t.Errorf("code review must mention code quality:\n%s", code)
	}

	for _, tt := range []string{"research", "document"} {
		doc := ManagerReviewWork("lead", "writer", "task", "sol", "files", tt)
		if contains(doc, "code quality") {
			t.Errorf("%s review must NOT mention code quality:\n%s", tt, doc)
		}
		if !contains(doc, "factually accurate") {
			t.Errorf("%s review must check facts:\n%s", tt, doc)
		}
	}

	pres := ManagerReviewWork("lead", "designer", "task", "sol", "files", "presentation")
	if !contains(pres, "slide") {
		t.Errorf("presentation review must mention slides:\n%s", pres)
	}
}

// TestValidateSolutionRoutesByTaskType — финальная валидация босса различается
// для кода и для документов/ресёрча/презентаций.
func TestValidateSolutionRoutesByTaskType(t *testing.T) {
	code := ValidateSolution("t", "tech", "go", "notes", "sum", "3", "list", "code")
	if !contains(code, "architecture") {
		t.Errorf("code validation must mention architecture:\n%s", code)
	}
	doc := ValidateSolution("t", "tech", "markdown", "notes", "sum", "1", "list", "document")
	if !contains(doc, "fact-check") {
		t.Errorf("document validation must mention fact-check:\n%s", doc)
	}
	if !contains(doc, "editor-in-chief") {
		t.Errorf("document validation should take an editor role:\n%s", doc)
	}
}

func TestPresentationWorkerAsksForVisualsAndSources(t *testing.T) {
	prompt := PresentationWorker("designer", "launch plan", "\ncontext", "")
	for _, want := range []string{"Visual:", "Source:", "image", "web search results", "Image:", "UNIQUE", "VARY THE STRUCTURE"} {
		if !contains(prompt, want) {
			t.Errorf("presentation prompt should ask for %q:\n%s", want, prompt)
		}
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
