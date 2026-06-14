package prompts

import (
	"strings"
	"testing"
)

// TestPlanArchitectureEnforcesScopeFidelity — регрессия на issue #38: для задачи
// «мини прокси на go» босс раздувал scope до полноценного REST API с JWT/БД.
// Промпт планирования теперь обязан явно требовать строгого следования задаче.
func TestPlanArchitectureEnforcesScopeFidelity(t *testing.T) {
	p := PlanArchitecture("Mini proxy", "Нужен мини прокси на go", 5, "")
	for _, want := range []string{
		"SCOPE FIDELITY",
		"EXACTLY what the user asked",
		"REST API", // explicitly warns against turning a proxy into a REST API
		"мини",     // recognizes Russian scope-limiting words
	} {
		if !strings.Contains(p, want) {
			t.Errorf("PlanArchitecture must enforce scope fidelity, missing %q:\n%s", want, p)
		}
	}
	// The grade must not be presented as a reason to inflate scope.
	if !strings.Contains(p, "NEVER inflate scope") {
		t.Errorf("PlanArchitecture must decouple complexity grade from scope:\n%s", p)
	}
}

// TestCodeManagerThinkHasWorkerFormula — менеджер кодовой команды использует
// единую формулу расчёта числа воркеров по grade_weight.
func TestCodeManagerThinkHasWorkerFormula(t *testing.T) {
	p := ManagerThink("backend", "Backend development", "Нужен мини прокси на go", "50", "code")
	for _, want := range []string{"Worker count by grade_weight", "1-20:", "21-60:"} {
		if !strings.Contains(p, want) {
			t.Errorf("codeManagerThink must have worker count formula, missing %q:\n%s", want, p)
		}
	}
}

// TestWorkerPlanFilesReturnsFilesAndCommands — воркер планирует файлы и команды
// в одном JSON (вместо отдельного LLM-вызова для команд).
func TestWorkerPlanFilesReturnsFilesAndCommands(t *testing.T) {
	p := WorkerPlanFiles("backend", "build it", "Нужен мини прокси на go", "", "go", "")
	for _, want := range []string{"Return JSON ONLY", "files", "commands"} {
		if !strings.Contains(p, want) {
			t.Errorf("WorkerPlanFiles must return files and commands in JSON, missing %q:\n%s", want, p)
		}
	}
}

// TestWorkerToolCommandsWarnsKeepMinimal — промпт выбора install-флагов напоминает
// о минимальности через «Keep it MINIMAL».
func TestWorkerToolCommandsWarnsKeepMinimal(t *testing.T) {
	p := WorkerToolCommands("backend", "build it", "express hello world server", "", "nodejs")
	for _, want := range []string{"Keep it MINIMAL", "Return JSON ONLY"} {
		if !strings.Contains(p, want) {
			t.Errorf("WorkerToolCommands must encourage minimal install, missing %q:\n%s", want, p)
		}
	}
}

// TestWorkerGenerateFileRequiresCompleteCode — воркер получает команду писать
// полный код без плейсхолдеров.
func TestWorkerGenerateFileRequiresCompleteCode(t *testing.T) {
	p := WorkerGenerateFile("package.json", "express hello world server", "backend", "nodejs", "")
	for _, want := range []string{"Write COMPLETE", "No placeholders. No TODOs"} {
		if !strings.Contains(p, want) {
			t.Errorf("WorkerGenerateFile must require complete code, missing %q:\n%s", want, p)
		}
	}
}

