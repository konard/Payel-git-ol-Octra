package boss

import (
	"os"
	"testing"

	gh "orchestrator/internal/service/github"
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
		req, dec := base()
		if !shouldUseUniversalNode(req, dec, nil) {
			t.Fatalf("grade %d should use the universal node", req.Grade)
		}
	})

	t.Run("grade at the threshold still trivial", func(t *testing.T) {
		req, dec := base()
		req.Grade = 2
		if !shouldUseUniversalNode(req, dec, nil) {
			t.Fatalf("grade %d (== threshold) should use the universal node", req.Grade)
		}
	})

	t.Run("complex task uses the full pipeline", func(t *testing.T) {
		req, dec := base()
		req.Grade = 5
		if shouldUseUniversalNode(req, dec, nil) {
			t.Fatalf("grade %d should NOT use the universal node", req.Grade)
		}
	})

	t.Run("ungraded task uses the full pipeline", func(t *testing.T) {
		req, dec := base()
		req.Grade = 0
		if shouldUseUniversalNode(req, dec, nil) {
			t.Fatal("ungraded task must not use the universal node")
		}
	})

	t.Run("predefined custom workflow is respected", func(t *testing.T) {
		req, dec := base()
		req.PredefinedManagers = []ManagerWorkflow{{Role: "lead"}}
		if shouldUseUniversalNode(req, dec, nil) {
			t.Fatal("user-defined workflow must not be replaced by the universal node")
		}
	})

	t.Run("refinement of an existing repo uses the full pipeline", func(t *testing.T) {
		req, dec := base()
		req.IsRefinement = true
		if shouldUseUniversalNode(req, dec, nil) {
			t.Fatal("refinement must not use the universal node")
		}
		req, dec = base()
		req.ExistingRepoUrl = "https://github.com/acme/repo"
		if shouldUseUniversalNode(req, dec, nil) {
			t.Fatal("existing-repo task must not use the universal node")
		}
	})

	t.Run("github issue task keeps its own pipeline", func(t *testing.T) {
		req, dec := base()
		if shouldUseUniversalNode(req, dec, &gh.IssueTarget{IssueURL: "https://github.com/a/b/issues/1"}) {
			t.Fatal("github issue task must not use the universal node")
		}
	})

	t.Run("disabled via env never triggers", func(t *testing.T) {
		t.Setenv("OCTRA_DISABLE_UNIVERSAL_NODE", "true")
		req, dec := base()
		if shouldUseUniversalNode(req, dec, nil) {
			t.Fatal("OCTRA_DISABLE_UNIVERSAL_NODE=true must disable the fast path")
		}
		os.Unsetenv("OCTRA_DISABLE_UNIVERSAL_NODE")
	})

	t.Run("nil inputs are safe", func(t *testing.T) {
		if shouldUseUniversalNode(nil, nil, nil) {
			t.Fatal("nil inputs must not crash or trigger the fast path")
		}
	})
}

// TestParseUniversalFiles — разбор строгого JSON-контракта универсальной ноды.
func TestParseUniversalFiles(t *testing.T) {
	t.Run("plain json", func(t *testing.T) {
		files := parseUniversalFiles(`{"files": {"main.py": "print('Hello, World!')"}}`)
		if len(files) != 1 || files["main.py"] != "print('Hello, World!')" {
			t.Fatalf("unexpected files: %#v", files)
		}
	})

	t.Run("json wrapped in markdown fences", func(t *testing.T) {
		raw := "```json\n{\"files\": {\"main.go\": \"package main\"}}\n```"
		files := parseUniversalFiles(raw)
		if files["main.go"] != "package main" {
			t.Fatalf("unexpected files: %#v", files)
		}
	})

	t.Run("normalizes ./ prefix", func(t *testing.T) {
		files := parseUniversalFiles(`{"files": {"./src/app.js": "x"}}`)
		if _, ok := files["src/app.js"]; !ok {
			t.Fatalf("path not normalized: %#v", files)
		}
	})

	t.Run("rejects traversal and absolute paths", func(t *testing.T) {
		files := parseUniversalFiles(`{"files": {"../escape": "x", "/etc/passwd": "y", "ok.txt": "z"}}`)
		if len(files) != 1 {
			t.Fatalf("expected only the safe file, got: %#v", files)
		}
		if files["ok.txt"] != "z" {
			t.Fatalf("safe file missing: %#v", files)
		}
	})

	t.Run("skips empty content", func(t *testing.T) {
		files := parseUniversalFiles(`{"files": {"empty.txt": "   ", "real.txt": "data"}}`)
		if len(files) != 1 || files["real.txt"] != "data" {
			t.Fatalf("unexpected files: %#v", files)
		}
	})

	t.Run("invalid json returns nil", func(t *testing.T) {
		if files := parseUniversalFiles("not json at all"); files != nil {
			t.Fatalf("expected nil, got: %#v", files)
		}
	})
}
