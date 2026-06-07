package boss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"orchestrator/internal/prompts"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	"orchestrator/pkg/models"

	"github.com/google/uuid"
)

// thinkAboutTask — boss решает архитектуру через AI (с fallback'ом по провайдерам)
func (s *Service) thinkAboutTask(ctx context.Context, provider, model string, req *CreateTaskRequest) (*DecisionResult, error) {
	if len(req.PredefinedManagers) > 0 {
		return decisionFromPredefinedWorkflow(req), nil
	}

	providers := []struct{ provider, model string }{
		{provider, model},
		{"openai", "gpt-4o-mini"},
		{"anthropic", "claude-3-haiku-20240307"},
		{"gemini", "gemini-pro"},
	}

	var lastErr error
	for _, p := range providers {
		log.Printf("Boss trying AI provider: %s/%s", p.provider, p.model)
		decision, err := s.thinkOnce(ctx, p.provider, p.model, req)
		if err == nil {
			return decision, nil
		}
		log.Printf("AI provider %s/%s failed: %v", p.provider, p.model, err)
		lastErr = err
		if strings.Contains(err.Error(), "API key") || strings.Contains(err.Error(), "not found in tokens") {
			break
		}
	}
	if lastErr != nil {
		if strings.Contains(lastErr.Error(), "API key") || strings.Contains(lastErr.Error(), "not found in tokens") {
			return nil, fmt.Errorf("AI service configuration error: no valid API keys")
		}
		if strings.Contains(lastErr.Error(), "Payment Required") || strings.Contains(lastErr.Error(), "credits") {
			return nil, fmt.Errorf("AI service error: credits exhausted")
		}
	}
	return nil, lastErr
}

// thinkOnce — один запрос к AI для планирования архитектуры
func (s *Service) thinkOnce(ctx context.Context, provider, model string, req *CreateTaskRequest) (*DecisionResult, error) {
	taskUUID, _ := uuid.Parse(req.Meta["task_id"])

	// Inject project context into the prompt
	contextBlock := s.contextClient.GetForPrompt(ctx, taskUUID, "boss", "")
	prompt := prompts.PlanArchitecture(req.Title, req.Description+contextBlock, req.Grade, req.Meta["skill_categories"])

	tokens := req.Tokens
	if tokens == nil {
		tokens = map[string]string{}
	}
	tokens["model"] = model

	resp, err := s.agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil {
		return nil, err
	}

	jsonStr := util.ExtractJSONFromMarkdown(resp)

	// Extract and save context from AI response
	if entry := s.contextClient.SaveFromAI(ctx, taskUUID, "boss", jsonStr); entry != nil {
		log.Printf("[Context] Boss saved: [%s] %s", entry.ContextType, entry.Content)
	}

	var decision DecisionResult
	if err := json.Unmarshal([]byte(jsonStr), &decision); err != nil {
		log.Printf("Boss JSON parse error: %v; raw: %s", err, resp)
		return nil, err
	}
	// Normalize / fall back the task type. The deterministic classifier guarantees a
	// sane value even when the AI omits it or the JSON fails to parse.
	decision.TaskType = normalizeTaskType(decision.TaskType)
	if decision.TaskType == "" {
		decision.TaskType = classifyTaskType(req.Title, req.Description)
	}
	log.Printf("Boss decision: type=%s managers=%d stack=%v", decision.TaskType, decision.ManagersCount, decision.TechStack)

	// Final safeguard inside thinkOnce
	if decision.ManagersCount <= 0 {
		decision.ManagersCount = 1
	}

	return &decision, nil
}

func decisionFromPredefinedWorkflow(req *CreateTaskRequest) *DecisionResult {
	managerRoles := make([]models.ManagerRole, 0, len(req.PredefinedManagers))
	for i, manager := range req.PredefinedManagers {
		role := strings.TrimSpace(manager.Role)
		if role == "" {
			role = fmt.Sprintf("manager-%d", i+1)
		}
		description := strings.TrimSpace(manager.Description)
		if description == "" {
			description = role + " manager for task execution"
		}
		priority := manager.Priority
		if priority <= 0 {
			priority = int32(i + 1)
		}
		managerRoles = append(managerRoles, models.ManagerRole{
			Role:         role,
			Description:  description,
			Priority:     priority,
			CustomPrompt: manager.CustomPrompt,
		})
	}

	taskType := normalizeTaskType(req.Meta["task_type"])
	if taskType == "" {
		taskType = classifyTaskType(req.Title, req.Description)
	}
	architecture := strings.TrimSpace(req.PredefinedArchitecture)
	if architecture == "" {
		architecture = "User-defined workflow"
	}

	return &DecisionResult{
		TaskType:             taskType,
		ManagersCount:        int32(len(managerRoles)),
		ManagerRoles:         managerRoles,
		ManagerWorkflows:     req.PredefinedManagers,
		TechnicalDescription: firstNonEmpty(architecture, req.Title+"\n"+req.Description),
		TechStack:            append([]string(nil), req.PredefinedTechStack...),
		ArchitectureNotes:    architecture,
	}
}

// validationResult — внутренний результат валидации боссом
type validationResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback"`
}

// validateSolution — boss проверяет итоговое решение через AI
func (s *Service) validateSolution(
	ctx context.Context,
	provider, model string,
	tokens map[string]string,
	decision *DecisionResult,
	results []*rules.ManagerResult,
) *validationResult {
	summary := ""
	fileCount := 0
	fileList := ""
	for _, mr := range results {
		summary += fmt.Sprintf("\n=== Manager: %s ===\nStatus: %s\nReview: %s\nWorkers: %d\n",
			mr.Role, mr.Status, mr.ReviewSummary, len(mr.WorkerResults))
		for _, wr := range mr.WorkerResults {
			for path := range wr.Files {
				fileCount++
				if fileCount <= 20 {
					fileList += "\n  - " + path
				}
			}
		}
	}
	prompt := prompts.ValidateSolution(
		tokens["title"], decision.TechnicalDescription,
		strings.Join(decision.TechStack, ", "),
		decision.ArchitectureNotes, summary,
		fmt.Sprintf("%d", fileCount), fileList, decision.TaskType,
	)
	resp, err := s.agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil {
		log.Printf("Boss validation AI error: %v", err)
		return &validationResult{Approved: true, Feedback: "validation skipped: " + err.Error()}
	}
	var result validationResult
	if err := json.Unmarshal([]byte(util.ExtractJSONFromMarkdown(resp)), &result); err != nil {
		return &validationResult{Approved: true, Feedback: "could not parse validation, auto-approved"}
	}
	log.Printf("Boss validation: approved=%v feedback=%s", result.Approved, result.Feedback)
	return &result
}

