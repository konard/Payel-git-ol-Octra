package prompts

import (
	"strings"
	"testing"
)

func TestPromptRank(t *testing.T) {
	cases := map[string]string{
		"gpt-4o-mini":             RankLight,
		"claude-3-haiku-20240307": RankLight,
		"gemini-pro":              RankLight,
		"gemini-1.5-flash":        RankLight,
		"gpt-4o":                  RankFull,
		"claude-3-5-sonnet":       RankFull,
		"":                        RankFull,
	}
	for model, want := range cases {
		if got := PromptRank(model); got != want {
			t.Errorf("PromptRank(%q) = %q, want %q", model, got, want)
		}
	}
}

// TestPlanArchitectureLightDropsSkillCatalog — слабые модели не должны получать
// громоздкий каталог скиллов (план фикса, пункт 9).
func TestPlanArchitectureLightDropsSkillCatalog(t *testing.T) {
	full := PlanArchitectureRanked("Build a proxy", "write a proxy", 5, "", RankFull)
	light := PlanArchitectureRanked("Build a proxy", "write a proxy", 5, "", RankLight)

	if len(light) >= len(full) {
		t.Fatalf("light prompt (%d) should be shorter than full (%d)", len(light), len(full))
	}
	// Both must still ask for JSON and classify task_type.
	for _, sub := range []string{"task_type", "Reply ONLY with JSON"} {
		if !strings.Contains(light, sub) {
			t.Errorf("light prompt missing %q", sub)
		}
	}
}
