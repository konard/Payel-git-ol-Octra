package boss

import (
	"strings"
	"testing"
)

func TestDecisionFromPredefinedWorkflow(t *testing.T) {
	req := &CreateTaskRequest{
		Title:                  "Research quarterly market data",
		Description:            "Find recent sources and produce a short brief",
		Meta:                   map[string]string{"task_type": TaskTypeResearch},
		PredefinedArchitecture: "Custom research workflow",
		PredefinedTechStack:    []string{"web search"},
		PredefinedManagers: []ManagerWorkflow{
			{
				Role:        "Research Lead",
				Description: "Coordinate source gathering",
				Priority:    3,
				Workers: []WorkerWorkflow{
					{Role: "Source Finder", Description: "Collect sources"},
				},
			},
		},
	}

	decision := decisionFromPredefinedWorkflow(req)
	if decision.TaskType != TaskTypeResearch {
		t.Fatalf("task type = %q, want %q", decision.TaskType, TaskTypeResearch)
	}
	if decision.ManagersCount != 1 {
		t.Fatalf("managers count = %d, want 1", decision.ManagersCount)
	}
	if got := decision.ManagerRoles[0].Role; got != "Research Lead" {
		t.Fatalf("manager role = %q, want Research Lead", got)
	}
	if got := len(predefinedWorkersFor(decision, 0, "Research Lead")); got != 1 {
		t.Fatalf("predefined worker count = %d, want 1", got)
	}
	// The technical description must be the real task (what workers implement),
	// not the workflow label. The architecture label lives in ArchitectureNotes.
	if !strings.Contains(decision.TechnicalDescription, "Research quarterly market data") ||
		!strings.Contains(decision.TechnicalDescription, "Find recent sources and produce a short brief") {
		t.Fatalf("technical description = %q, want it to contain the real task", decision.TechnicalDescription)
	}
	if decision.ArchitectureNotes != "Custom research workflow" {
		t.Fatalf("architecture notes = %q, want %q", decision.ArchitectureNotes, "Custom research workflow")
	}
	if len(decision.TechStack) != 1 || decision.TechStack[0] != "web search" {
		t.Fatalf("tech stack = %#v, want [web search]", decision.TechStack)
	}
}

// TestDecisionFromPredefinedWorkflowCarriesRealTask is the regression test for
// issue #87: a custom workflow built in the canvas (Boss → Managers → Workers)
// must hand workers the actual user request, otherwise they get only the
// "Custom workflow with ... managers" label and produce no code.
func TestDecisionFromPredefinedWorkflowCarriesRealTask(t *testing.T) {
	req := &CreateTaskRequest{
		Title:       "Build a proxy server",
		Description: "Write a reverse proxy in Python with load balancing",
		// Architecture is the auto-generated label sent by the frontend canvas.
		PredefinedArchitecture: "Custom workflow with Architect and 2 managers",
		// No tech stack set on the Boss node — must be detected from the task.
		PredefinedManagers: []ManagerWorkflow{
			{
				Role:        "Backend",
				Description: "Backend manager",
				Priority:    1,
				Workers: []WorkerWorkflow{
					{Role: "Developer", Description: "Developer worker"},
				},
			},
			{
				Role:        "Coordinator",
				Description: "Coordinator manager",
				Priority:    2,
				Workers: []WorkerWorkflow{
					{Role: "Developer", Description: "Developer worker"},
					{Role: "Designer", Description: "Designer worker"},
				},
			},
		},
	}

	decision := decisionFromPredefinedWorkflow(req)

	if strings.Contains(decision.TechnicalDescription, "Custom workflow with") {
		t.Fatalf("technical description leaked the workflow label: %q", decision.TechnicalDescription)
	}
	if !strings.Contains(decision.TechnicalDescription, "Build a proxy server") ||
		!strings.Contains(decision.TechnicalDescription, "reverse proxy in Python") {
		t.Fatalf("technical description lost the real task: %q", decision.TechnicalDescription)
	}
	if decision.ManagersCount != 2 {
		t.Fatalf("managers count = %d, want 2", decision.ManagersCount)
	}
	// Tech stack must be detected from the request when the workflow omits it.
	if len(decision.TechStack) == 0 || decision.TechStack[0] != "python" {
		t.Fatalf("tech stack = %#v, want python detected from the task", decision.TechStack)
	}
}

func TestFallbackManagerRoleMatchesTaskType(t *testing.T) {
	cases := []struct {
		taskType string
		wantRole string
	}{
		{TaskTypePresentation, "presentation"},
		{TaskTypeResearch, "research"},
		{TaskTypeDocument, "document"},
		{TaskTypeCode, "development"},
	}

	for _, tc := range cases {
		t.Run(tc.taskType, func(t *testing.T) {
			role := fallbackManagerRole(tc.taskType)
			if role.Role != tc.wantRole {
				t.Fatalf("fallbackManagerRole(%q).Role = %q, want %q", tc.taskType, role.Role, tc.wantRole)
			}
			if role.Description == "" {
				t.Fatalf("fallbackManagerRole(%q) returned empty description", tc.taskType)
			}
		})
	}
}

