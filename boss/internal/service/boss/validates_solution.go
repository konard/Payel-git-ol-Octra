package boss

import (
	"boss/internal/fetcher/grpc/manager/managerpb"
	"boss/internal/service"
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
func (s *BossService) validateSolution(ctx context.Context, agentsClient *service.AgentClientWrapper, provider, model string, tokens map[string]string, decision *BossDecisionResult, managerResults []*managerpb.ManagerResult, zipData []byte) (*ValidationResult, error) {
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

	prompt := `You are the CTO (Chief Technology Officer) reviewing the final deliverable.

ORIGINAL TASK:
Title: ` + tokens["title"] + `

ARCHITECTURE DECISION:
Technical: ` + decision.TechnicalDescription + `
Stack: ` + strings.Join(decision.TechStack, ", ") + `
` + decision.ArchitectureNotes + `

MANAGERS RESULTS:
` + summary + `

GENERATED FILES (` + strconv.Itoa(fileCount) + ` total):
` + fileList + `

ZIP size: ` + strconv.Itoa(len(zipData)) + ` bytes

Review:
1. Does the solution meet the requirements?
2. Is the architecture followed?
3. Are all managers completed their work?
4. Is the file structure reasonable?
5. Any critical issues?

Reply ONLY with JSON:
{
  "approved": true/false,
  "feedback": "detailed feedback"
}`

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
