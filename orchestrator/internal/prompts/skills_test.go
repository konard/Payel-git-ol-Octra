package prompts

import (
	"strings"
	"testing"

	"orchestrator/internal/skills"
)

// TestPlanArchitectureListsSkillCatalog — босс видит каталог системных скиллов,
// чтобы подбирать роли под реальные специальности.
func TestPlanArchitectureListsSkillCatalog(t *testing.T) {
	p := PlanArchitecture("Build a proxy", "write a reverse proxy in Go", 5, "")
	if !strings.Contains(p, "EXPERT SKILLS") {
		t.Errorf("plan prompt should mention expert skills catalog:\n%s", p)
	}
	if !strings.Contains(p, "Proxy") {
		t.Errorf("plan prompt should list the Proxy area:\n%s", p)
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

