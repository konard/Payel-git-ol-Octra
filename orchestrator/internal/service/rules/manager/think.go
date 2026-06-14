package manager

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"orchestrator/internal/config"
	"orchestrator/internal/prompts"
	"orchestrator/internal/service/util"
	"orchestrator/pkg/models"

	"github.com/google/uuid"
)

// think — менеджер решает каких воркеров нанять
func (s *Service) think(ctx context.Context, provider, model string, tokens map[string]string, taskDesc, role, description, gradeWeight, taskType string, projectID uuid.UUID, managerID string) ([]models.WorkerRole, error) {
	log.Printf("Manager (%s) thinking about workers... (grade: %s, type: %s)", role, gradeWeight, taskType)

	// Inject project context into prompt
	contextBlock := s.contextClient.GetForPrompt(ctx, projectID, "manager-"+managerID, managerID)
	prompt := prompts.ManagerThink(role, description, taskDesc+contextBlock, gradeWeight, taskType)

	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.agentsClient.Generate(genCtx, provider, model, prompt, tokens, 2048, config.Temperature)
	if err != nil {
		return []models.WorkerRole{fallbackWorkerRole(taskType, "fallback due to timeout")}, nil
	}

	jsonStr := util.ExtractJSONFromMarkdown(resp)

	// Extract and save context from AI response
	if entry := s.contextClient.SaveFromAI(ctx, projectID, "manager-"+managerID, jsonStr); entry != nil {
		log.Printf("[Context] Manager saved: [%s] %s", entry.ContextType, entry.Content)
	}

	var result struct {
		WorkerRoles []models.WorkerRole `json:"worker_roles"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
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

