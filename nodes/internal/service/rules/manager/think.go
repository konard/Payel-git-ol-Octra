package manager

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"nodes/internal/prompts"
	"nodes/internal/service/util"
	"nodes/pkg/models"
)

// think — менеджер решает каких воркеров нанять
func (s *Service) think(ctx context.Context, provider, model string, tokens map[string]string, taskDesc, role, description, gradeWeight string) ([]models.WorkerRole, error) {
	log.Printf("Manager (%s) thinking about workers... (grade: %s)", role, gradeWeight)

	prompt := prompts.ManagerThink(role, description, taskDesc, gradeWeight)

	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.agentsClient.Generate(genCtx, provider, model, prompt, tokens, 2048, 0.7)
	if err != nil {
		return []models.WorkerRole{{Role: "developer", Description: "fallback due to timeout"}}, nil
	}

	var result struct {
		WorkerRoles []models.WorkerRole `json:"worker_roles"`
	}
	if err := json.Unmarshal([]byte(util.ExtractJSONFromMarkdown(resp)), &result); err != nil {
		return []models.WorkerRole{{Role: "developer", Description: resp}}, nil
	}
	if len(result.WorkerRoles) == 0 {
		return []models.WorkerRole{{Role: "developer", Description: "General developer for " + role}}, nil
	}

	log.Printf("Manager decided to hire: %v", result.WorkerRoles)
	return result.WorkerRoles, nil
}
