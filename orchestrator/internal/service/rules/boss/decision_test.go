package boss

import (
	"encoding/json"
	"testing"
)

// TestDecisionResultUnmarshalsManagerRoles is the regression test for issue #75
// point 4 ("only 3 nodes are created regardless of complexity"). The AI returns
// snake_case keys; without explicit json tags the boss decoded zero manager
// roles and always fell back to a single manager. With the tags in place, the
// AI's manager plan must be honored verbatim.
func TestDecisionResultUnmarshalsManagerRoles(t *testing.T) {
	raw := `{
		"task_type": "code",
		"managers_count": 2,
		"manager_roles": [
			{"role": "frontend", "description": "Build the UI", "priority": 1},
			{"role": "backend", "description": "Build the API", "priority": 2}
		],
		"tech_stack": ["nodejs"],
		"technical_description": "Full-stack web app",
		"architecture_notes": "Separate frontend and backend managers"
	}`

	var d DecisionResult
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if d.ManagersCount != 2 {
		t.Errorf("ManagersCount = %d, want 2", d.ManagersCount)
	}
	if len(d.ManagerRoles) != 2 {
		t.Fatalf("ManagerRoles len = %d, want 2 (the AI plan must not be dropped)", len(d.ManagerRoles))
	}
	if d.ManagerRoles[0].Role != "frontend" || d.ManagerRoles[1].Role != "backend" {
		t.Errorf("ManagerRoles = %+v, want frontend/backend", d.ManagerRoles)
	}
	if len(d.TechStack) != 1 || d.TechStack[0] != "nodejs" {
		t.Errorf("TechStack = %v, want [nodejs]", d.TechStack)
	}
	if d.TechnicalDescription != "Full-stack web app" {
		t.Errorf("TechnicalDescription = %q", d.TechnicalDescription)
	}
	if d.ArchitectureNotes != "Separate frontend and backend managers" {
		t.Errorf("ArchitectureNotes = %q", d.ArchitectureNotes)
	}
}
