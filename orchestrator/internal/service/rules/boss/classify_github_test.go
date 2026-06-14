package boss

import (
	"testing"

	"orchestrator/pkg/models"
)

func TestClampToSingleManager(t *testing.T) {
	d := &DecisionResult{
		ManagersCount: 3,
		ManagerRoles: []models.ManagerRole{
			{Role: "backend"}, {Role: "frontend"}, {Role: "devops"},
		},
		ManagerWorkflows: []ManagerWorkflow{{Role: "backend"}, {Role: "frontend"}},
	}
	clampToSingleManager(d)
	if d.ManagersCount != 1 {
		t.Fatalf("ManagersCount = %d, want 1", d.ManagersCount)
	}
	if len(d.ManagerRoles) != 1 || d.ManagerRoles[0].Role != "backend" {
		t.Fatalf("ManagerRoles = %#v, want single backend", d.ManagerRoles)
	}
	if len(d.ManagerWorkflows) != 1 {
		t.Fatalf("ManagerWorkflows len = %d, want 1", len(d.ManagerWorkflows))
	}
	clampToSingleManager(nil) // must not panic
}

func TestNormalizeTaskTypeGitHub(t *testing.T) {
	for _, in := range []string{"github", "GitHub", "issue", "pull_request", "pr"} {
		if got := normalizeTaskType(in); got != TaskTypeGitHub {
			t.Errorf("normalizeTaskType(%q) = %q, want %q", in, got, TaskTypeGitHub)
		}
	}
}

func TestIsCodeLikeTask(t *testing.T) {
	codeLike := []string{"", TaskTypeCode, TaskTypeGitHub}
	for _, tt := range codeLike {
		if !isCodeLikeTask(tt) {
			t.Errorf("isCodeLikeTask(%q) = false, want true", tt)
		}
	}
	notCode := []string{TaskTypeResearch, TaskTypeDocument, TaskTypePresentation}
	for _, tt := range notCode {
		if isCodeLikeTask(tt) {
			t.Errorf("isCodeLikeTask(%q) = true, want false", tt)
		}
	}
}
