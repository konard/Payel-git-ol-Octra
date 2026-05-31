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
func (s *Service) think(ctx context.Context, provider, model string, tokens map[string]string, taskDesc, role, description, gradeWeight, taskType string) ([]models.WorkerRole, error) {
	log.Printf("Manager (%s) thinking about workers... (grade: %s, type: %s)", role, gradeWeight, taskType)

	prompt := prompts.ManagerThink(role, description, taskDesc, gradeWeight, taskType)

	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.agentsClient.Generate(genCtx, provider, model, prompt, tokens, 2048, 0.7)
	if err != nil {
		return []models.WorkerRole{fallbackWorkerRole(taskType, "fallback due to timeout")}, nil
	}

	var result struct {
		WorkerRoles []models.WorkerRole `json:"worker_roles"`
	}
	if err := json.Unmarshal([]byte(util.ExtractJSONFromMarkdown(resp)), &result); err != nil {
		return []models.WorkerRole{fallbackWorkerRole(taskType, resp)}, nil
	}
	if len(result.WorkerRoles) == 0 {
		return []models.WorkerRole{fallbackWorkerRole(taskType, "General worker for "+role)}, nil
	}

	log.Printf("Manager decided to hire: %v", result.WorkerRoles)
	return result.WorkerRoles, nil
}

// fallbackWorkerRole — дефолтный воркер под тип задачи, когда AI не вернул валидный план.
func fallbackWorkerRole(taskType, description string) models.WorkerRole {
	role := "developer"
	switch taskType {
	case "research":
		role = "analyst"
	case "document":
		role = "writer"
	case "presentation":
		role = "designer"
	}
	return models.WorkerRole{Role: role, Description: description}
}
