package boss

import "testing"

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
	if decision.TechnicalDescription != "Custom research workflow" {
		t.Fatalf("technical description = %q", decision.TechnicalDescription)
	}
	if len(decision.TechStack) != 1 || decision.TechStack[0] != "web search" {
		t.Fatalf("tech stack = %#v, want [web search]", decision.TechStack)
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
