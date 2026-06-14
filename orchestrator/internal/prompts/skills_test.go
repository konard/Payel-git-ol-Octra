package prompts

import (
	"strings"
	"testing"

	"orchestrator/internal/skills"
)

// TestPlanArchitectureSkillCatalogOptIn — каталог скиллов теперь opt-in (issue #79):
// без запрошенных категорий босс НЕ получает каталог (меньше шума для простых задач),
// а с явными категориями — получает только релевантные фрагменты.
func TestPlanArchitectureSkillCatalogOptIn(t *testing.T) {
	// Без категорий: каталог не должен попадать в промпт.
	noCat := PlanArchitecture("hello world", "напиши hello world на express js", 8, "")
	if strings.Contains(noCat, "SKILL WAREHOUSE CATALOG") || strings.Contains(noCat, "skill categories") {
		t.Errorf("plan prompt must NOT dump the skill catalog when no categories requested:\n%s", noCat)
	}

	// С явными категориями: релевантный фрагмент должен присутствовать.
	withCat := PlanArchitecture("Build a proxy", "write a reverse proxy in Go", 5, "proxy")
	if !strings.Contains(withCat, "skill categories") {
		t.Errorf("plan prompt should mention requested skill categories:\n%s", withCat)
	}
}

// TestWorkerPromptsInjectSkill — экспертный блок скилла должен попадать в промпты
// планирования и генерации файлов воркера.
func TestWorkerPromptsInjectSkill(t *testing.T) {
	skill := skills.Guidance("backend engineer", "go", "build an API")
	if skill == "" {
		t.Fatal("expected non-empty backend guidance")
	}
	plan := WorkerPlanFiles("backend", "API work", "build an API", "", "go", skill)
	if !strings.Contains(plan, "APPLY THIS EXPERT SKILL") {
		t.Errorf("plan prompt should embed skill guidance:\n%s", plan)
	}
	file := WorkerGenerateFile("main.go", "build an API", "backend", "go", skill)
	if !strings.Contains(file, "APPLY THIS EXPERT SKILL") {
		t.Errorf("file prompt should embed skill guidance:\n%s", file)
	}
}

