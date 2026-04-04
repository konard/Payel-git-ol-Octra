package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"manager/pkg/models"
)

// managerThink — менеджер решает каких воркеров нанять (через agents сервис)
func (s *ManagerService) managerThink(ctx context.Context, provider, model string, tokens map[string]string, taskDesc, role, description string) ([]models.WorkerRole, error) {
	log.Printf("Manager (%s) thinking about workers...", role)

	prompt := fmt.Sprintf(`You are a project manager for a %s team ONLY. Your team's focus: %s

Task:
%s

You are ONLY responsible for the %s aspect of this project.
What workers do you need on YOUR team? Think about specific sub-roles within your domain.

IMPORTANT: Only list workers that belong to YOUR team (%s).
- If you are "frontend" manager → only list frontend workers (React developers, UI designers, etc.)
- If you are "backend" manager → only list backend workers (API developers, DB engineers, etc.)
- If you are "devops/qa" manager → only list devops/qa workers (CI/CD, testing, monitoring, etc.)

Do NOT list workers from other teams. Each team has its own manager.

Reply ONLY with JSON:
{
  "worker_roles": [
    {"role": "role1", "description": "What this role does"},
    {"role": "role2", "description": "What this role does"}
  ]
}`, role, description, taskDesc, role, role)

	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 2048, 0.7)
	if err != nil {
		return nil, err
	}

	var result struct {
		WorkerRoles []models.WorkerRole `json:"worker_roles"`
	}

	if err := json.Unmarshal([]byte(extractJSONFromMarkdown(resp)), &result); err != nil {
		// Fallback: create single worker role from response
		return []models.WorkerRole{{Role: "developer", Description: resp}}, nil
	}

	if len(result.WorkerRoles) == 0 {
		return []models.WorkerRole{{Role: "developer", Description: "General developer for " + role}}, nil
	}

	log.Printf("Manager decided to hire: %v", result.WorkerRoles)
	return result.WorkerRoles, nil
}
