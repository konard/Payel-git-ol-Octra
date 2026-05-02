package manager

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"manager/pkg/models"
	"manager/internal/prompts"
)

// managerThink — менеджер решает каких воркеров нанять (через agents сервис)
func (s *ManagerService) managerThink(ctx context.Context, provider, model string, tokens map[string]string, taskDesc, role, description, gradeWeight string) ([]models.WorkerRole, error) {
	log.Printf("Manager (%s) thinking about workers... (grade: %s)", role, gradeWeight)

	prompt := prompts.Think(role, description, taskDesc, gradeWeight)

	// Timeout 30s to avoid WS disconnect
	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.agentsClient.Generate(genCtx, provider, model, prompt, tokens, 2048, 0.7)
	if err != nil {
		// Fallback: single worker on timeout
		return []models.WorkerRole{{Role: "developer", Description: "fallback due to timeout"}}, nil
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
