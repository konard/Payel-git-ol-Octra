package worker

import (
	"orchestrator/internal/skills"
)

// buildSkillContext resolves skill fragments by slugs (warehouse) or falls
// back to the traditional skills.Guidance() when no slugs were selected.
// This is the single entry point for all code-generation paths — it ensures
// workers only get the skill content they actually need.
func buildSkillContext(selectedSlugs []string, role, techStack, task, description string) string {
	if len(selectedSlugs) > 0 {
		w := skills.NewWarehouse()
		if s := w.Select(selectedSlugs...); s != "" {
			return s
		}
	}
	// Fallback: old monolithic guidance (backward compat)
	return skills.Guidance(role+" "+description, techStack, task)
}
