package boss

import (
	"os"
	"testing"

	gh "orchestrator/internal/service/github"
	"orchestrator/internal/service/rules/universal"
)

// TestShouldUseUniversalNode — быстрый путь для тривиальных задач (issue #91).
// Решение опирается ТОЛЬКО на оценку сложности (1-10) и контекст задачи, а не на
// триггер-слова: тривиальная задача идёт через одну универсальную ноду, всё
// остальное — через полный конвейер.
func TestShouldUseUniversalNode(t *testing.T) {
	base := func() (*CreateTaskRequest, *DecisionResult) {
		return &CreateTaskRequest{
				Title:       "hello world",
				Description: "напиши hello world на python",
				Grade:       1,
			}, &DecisionResult{
				TaskType:  TaskTypeCode,
				TechStack: []string{"python"},
			}
	}

	t.Run("trivial code task takes the fast path", func(t *testing.T) {
		req, _ := base()
		if !shouldUseUniversalNode(req, nil) {
			t.Fatalf("grade %d should use the universal node", req.Grade)
		}
	})

	t.Run("grade at the threshold still trivial", func(t *testing.T) {
		req, _ := base()
		req.Grade = 2
		if !shouldUseUniversalNode(req, nil) {
			t.Fatalf("grade %d (== threshold) should use the universal node", req.Grade)
		}
	})

	t.Run("complex task uses the full pipeline", func(t *testing.T) {
		req, _ := base()
		req.Grade = 5
		if shouldUseUniversalNode(req, nil) {
			t.Fatalf("grade %d should NOT use the universal node", req.Grade)
		}
	})

	t.Run("ungraded task uses the full pipeline", func(t *testing.T) {
		req, _ := base()
		req.Grade = 0
		if shouldUseUniversalNode(req, nil) {
			t.Fatal("ungraded task must not use the universal node")
		}
	})

	t.Run("predefined custom workflow is respected", func(t *testing.T) {
		req, _ := base()
		req.PredefinedManagers = []ManagerWorkflow{{Role: "lead"}}
		if shouldUseUniversalNode(req, nil) {
			t.Fatal("user-defined workflow must not be replaced by the universal node")
		}
	})

	t.Run("refinement of an existing repo uses the full pipeline", func(t *testing.T) {
		req, _ := base()
		req.IsRefinement = true
		if shouldUseUniversalNode(req, nil) {
			t.Fatal("refinement must not use the universal node")
		}
		req, _ = base()
		req.ExistingRepoUrl = "https://github.com/acme/repo"
		if shouldUseUniversalNode(req, nil) {
			t.Fatal("existing-repo task must not use the universal node")
		}
	})

	t.Run("github issue task keeps its own pipeline", func(t *testing.T) {
		req, _ := base()
		if shouldUseUniversalNode(req, &gh.IssueTarget{IssueURL: "https://github.com/a/b/issues/1"}) {
			t.Fatal("github issue task must not use the universal node")
		}
	})

	t.Run("disabled via env never triggers", func(t *testing.T) {
		t.Setenv("OCTRA_DISABLE_UNIVERSAL_NODE", "true")
		req, _ := base()
		if shouldUseUniversalNode(req, nil) {
			t.Fatal("OCTRA_DISABLE_UNIVERSAL_NODE=true must disable the fast path")
		}
		os.Unsetenv("OCTRA_DISABLE_UNIVERSAL_NODE")
	})

	t.Run("nil inputs are safe", func(t *testing.T) {
		if shouldUseUniversalNode(nil, nil) {
			t.Fatal("nil inputs must not crash or trigger the fast path")
		}
	})
}

func TestUniversalDecisionFromRequest(t *testing.T) {
	t.Run("code task has no boss managers", func(t *testing.T) {
		decision := universalDecisionFromRequest(&CreateTaskRequest{
			Title:       "hello world",
			Description: "напиши hello world на python",
		})
		if decision.TaskType != TaskTypeCode {
			t.Fatalf("task type = %q, want %q", decision.TaskType, TaskTypeCode)
		}
		if decision.ManagersCount != 0 || len(decision.ManagerRoles) != 0 {
			t.Fatalf("universal direct path must not invent boss managers: %#v", decision)
		}
		if len(decision.TechStack) != 1 || decision.TechStack[0] != "python" {
			t.Fatalf("tech stack = %#v, want [python]", decision.TechStack)
		}
	})

	t.Run("math or logic task becomes direct answer", func(t *testing.T) {
		decision := universalDecisionFromRequest(&CreateTaskRequest{
			Title:       "math",
			Description: "Сколько будет 2 + 2?",
		})
		if decision.TaskType != TaskTypeDocument {
			t.Fatalf("task type = %q, want %q", decision.TaskType, TaskTypeDocument)
		}
		if decision.ManagersCount != 0 {
			t.Fatalf("managers count = %d, want 0", decision.ManagersCount)
		}
	})

	t.Run("code request phrased as a question stays code", func(t *testing.T) {
		decision := universalDecisionFromRequest(&CreateTaskRequest{
			Title:       "small function",
			Description: "Can you write a function that adds two numbers?",
		})
		if decision.TaskType != TaskTypeCode {
			t.Fatalf("task type = %q, want %q", decision.TaskType, TaskTypeCode)
		}
	})

	t.Run("concept question stays direct document answer", func(t *testing.T) {
		decision := universalDecisionFromRequest(&CreateTaskRequest{
			Title:       "API question",
			Description: "What is an API?",
		})
		if decision.TaskType != TaskTypeDocument {
			t.Fatalf("task type = %q, want %q", decision.TaskType, TaskTypeDocument)
		}
	})
}

// TestParseFiles — разбор строгого JSON-контракта универсальной ноды.
func TestParseUniversalFiles(t *testing.T) {
	t.Run("plain json", func(t *testing.T) {
		r := universal.ParseResponse(`{"files": {"main.py": "print('Hello, World!')"}}`)
		if r == nil || len(r.Files) != 1 || r.Files["main.py"] != "print('Hello, World!')" {
			t.Fatalf("unexpected files: %#v", r)
		}
	})

	t.Run("json wrapped in markdown fences", func(t *testing.T) {
		raw := "```json\n{\"files\": {\"main.go\": \"package main\"}}\n```"
		r := universal.ParseResponse(raw)
		if r == nil || r.Files["main.go"] != "package main" {
			t.Fatalf("unexpected files: %#v", r)
		}
	})

	t.Run("normalizes ./ prefix", func(t *testing.T) {
		r := universal.ParseResponse(`{"files": {"./src/app.js": "x"}}`)
		if r == nil || r.Files["src/app.js"] != "x" {
			t.Fatalf("path not normalized: %#v", r)
		}
	})

	t.Run("rejects traversal and absolute paths", func(t *testing.T) {
		r := universal.ParseResponse(`{"files": {"../escape": "x", "/etc/passwd": "y", "ok.txt": "z"}}`)
		if r == nil || len(r.Files) != 1 || r.Files["ok.txt"] != "z" {
			t.Fatalf("expected only the safe file, got: %#v", r)
		}
	})

	t.Run("skips empty content", func(t *testing.T) {
		r := universal.ParseResponse(`{"files": {"empty.txt": "   ", "real.txt": "data"}}`)
		if r == nil || len(r.Files) != 1 || r.Files["real.txt"] != "data" {
			t.Fatalf("unexpected files: %#v", r)
		}
	})

	t.Run("parses dependencies flag", func(t *testing.T) {
		r := universal.ParseResponse(`{"files": {"main.py": "x"}, "dependencies": true}`)
		if r == nil || !r.Dependencies {
			t.Fatal("expected dependencies=true")
		}
	})

	t.Run("defaults dependencies to false", func(t *testing.T) {
		r := universal.ParseResponse(`{"files": {"main.py": "x"}}`)
		if r == nil || r.Dependencies {
			t.Fatal("expected dependencies=false (default)")
		}
	})

	t.Run("invalid json returns nil", func(t *testing.T) {
		if r := universal.ParseResponse("not json at all"); r != nil {
			t.Fatalf("expected nil, got: %#v", r)
		}
	})
}
