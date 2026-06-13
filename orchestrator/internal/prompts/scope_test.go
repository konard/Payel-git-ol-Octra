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

// TestCodeManagerThinkEnforcesScopeFidelity — менеджер кодовой команды не должен
// нанимать лишних воркеров под незапрошенные фичи.
func TestCodeManagerThinkEnforcesScopeFidelity(t *testing.T) {
	p := ManagerThink("backend", "Backend development", "Нужен мини прокси на go", "50", "code")
	for _, want := range []string{"SCOPE FIDELITY", "EXACTLY what the task asks"} {
		if !strings.Contains(p, want) {
			t.Errorf("codeManagerThink must enforce scope fidelity, missing %q:\n%s", want, p)
		}
	}
}

// TestWorkerPlanFilesEnforcesScopeFidelity — воркер должен планировать минимальный
// набор файлов и не добавлять файлы под незапрошенные фичи.
func TestWorkerPlanFilesEnforcesScopeFidelity(t *testing.T) {
	p := WorkerPlanFiles("backend", "build it", "Нужен мини прокси на go", "", "go", "")
	for _, want := range []string{"SCOPE FIDELITY", "MINIMUM set of files"} {
		if !strings.Contains(p, want) {
			t.Errorf("WorkerPlanFiles must enforce scope fidelity, missing %q:\n%s", want, p)
		}
	}
}

// TestWorkerToolCommandsEnforcesScopeFidelity — регрессия на issue #75 п.5: для
// «express hello world server» воркер ставил prisma/jwt/bcrypt/zod/winston/React.
// Промпт выбора install-флагов теперь обязан запрещать незапрошенные флаги.
func TestWorkerToolCommandsEnforcesScopeFidelity(t *testing.T) {
	p := WorkerToolCommands("backend", "build it", "express hello world server", "", "nodejs")
	for _, want := range []string{"SCOPE FIDELITY", "prisma", "jwt", "bcrypt", "zod", "winston", "React"} {
		if !strings.Contains(p, want) {
			t.Errorf("WorkerToolCommands must warn against over-installing, missing %q:\n%s", want, p)
		}
	}
}

// TestWorkerGenerateFileEnforcesScopeFidelity — содержимое файла (в т.ч. манифеста)
// не должно тянуть незапрошенные зависимости (issue #75 п.5).
func TestWorkerGenerateFileEnforcesScopeFidelity(t *testing.T) {
	p := WorkerGenerateFile("package.json", "express hello world server", "backend", "nodejs", "")
	for _, want := range []string{"SCOPE FIDELITY", "ONLY the dependencies actually used"} {
		if !strings.Contains(p, want) {
			t.Errorf("WorkerGenerateFile must enforce scope fidelity, missing %q:\n%s", want, p)
		}
	}
}

