package boss

import (
	"boss/internal/fetcher/grpc/manager/managerpb"
	"boss/internal/service"
	"boss/internal/prompts"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

type ValidationResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback"`
}

// validateSolution — Boss validates итоговое решение через AI
func (s *BossService) validateSolution(ctx context.Context, agentsClient *service.AgentClientWrapper, provider, model string, tokens map[string]string, decision *BossDecisionResult, managerResults []*managerpb.ManagerResult) (*ValidationResult, error) {
	// Build summary of all managers work
	summary := ""
	for _, mr := range managerResults {
		summary += "\n=== Manager: " + mr.Role + " ===\n"
		summary += "Status: " + mr.Status + "\n"
		summary += "Review: " + mr.ReviewSummary + "\n"
		summary += "Workers: " + strconv.Itoa(len(mr.WorkerResults)) + "\n"
	}

	// Count total files
	fileCount := 0
	fileList := ""
	for _, mr := range managerResults {
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
		tokens["title"],
		decision.TechnicalDescription,
		strings.Join(decision.TechStack, ", "),
		decision.ArchitectureNotes,
		summary,
		strconv.Itoa(fileCount),
		fileList,
	)

	resp, err := agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSONFromMarkdown(resp)
	var result ValidationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &ValidationResult{Approved: true, Feedback: "Could not parse validation, auto-approved"}, nil
	}

	log.Printf("Boss validation: approved=%v, feedback=%s", result.Approved, result.Feedback)
	return &result, nil
}

// extractJSONFromMarkdown извлекает JSON из markdown блоков ```json ... ``` или ``` ... ```
func extractJSONFromMarkdown(s string) string {
	// Ищем ```json или ``` и закрывающие ```
	start := -1
	for _, marker := range []string{"```json\n", "```\n", "```"} {
		if idx := strings.Index(s, marker); idx >= 0 {
			start = idx + len(marker)
			break
		}
	}

	if start >= 0 {
		// Ищем закрывающие ```
		end := strings.Index(s[start:], "```")
		if end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
		// Если нет закрывающих, берём всё после открывающих
		return strings.TrimSpace(s[start:])
	}

	// Нет markdown — возвращаем как есть
	return s
}

// Helper functions
func marshalString(m map[string]string) string {
	data, _ := json.Marshal(m)
	return string(data)
}
