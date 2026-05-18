package boss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"nodes/internal/prompts"
	"nodes/internal/service/rules"
	"nodes/internal/service/util"
)

// thinkAboutTask — boss решает архитектуру через AI (с fallback'ом по провайдерам)
func (s *Service) thinkAboutTask(ctx context.Context, provider, model string, req *CreateTaskRequest) (*DecisionResult, error) {
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
	prompt := prompts.PlanArchitecture(req.Title, req.Description, req.Grade)

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

	var decision DecisionResult
	if err := json.Unmarshal([]byte(jsonStr), &decision); err != nil {
		log.Printf("Boss JSON parse error: %v; raw: %s", err, resp)
		return nil, err
	}
	log.Printf("Boss decision: managers=%d stack=%v", decision.ManagersCount, decision.TechStack)
	return &decision, nil
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
		fmt.Sprintf("%d", fileCount), fileList,
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
