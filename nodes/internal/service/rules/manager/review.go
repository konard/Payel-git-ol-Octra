package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"nodes/internal/prompts"
	"nodes/internal/service/rules"
	"nodes/internal/service/util"
)

// isBinaryReviewPath — файлы, которые нельзя показывать ревьюеру/чинить как текст
// (например, .pptx собирается из Markdown билдером, а не правится построчно).
func isBinaryReviewPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pptx", ".docx", ".xlsx", ".pdf", ".png", ".jpg", ".jpeg", ".gif", ".zip":
		return true
	default:
		return false
	}
}

// reviewResult — внутренний результат ревью
type reviewResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback"`
}

// reviewWorkerResult — менеджер ревьюит работу одного воркера через LLM
func (s *Service) reviewWorkerResult(ctx context.Context, provider, model string, tokens map[string]string, managerRole, taskType string, wr *rules.WorkerResult) (*reviewResult, error) {
	filesList := ""
	for path, content := range wr.Files {
		// Бинарные артефакты (например, .pptx) не показываем ревьюеру как текст —
		// он оценивает исходный Markdown, а не байты презентации.
		if isBinaryReviewPath(path) {
			filesList += fmt.Sprintf("\n--- %s (binary artifact, not shown) ---\n", path)
			continue
		}
		preview := content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		filesList += fmt.Sprintf("\n--- %s ---\n%s\n", path, preview)
	}

	prompt := prompts.ManagerReviewWork(managerRole, wr.Role, wr.TaskMd, wr.SolutionMd, filesList, taskType)

	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.agentsClient.Generate(genCtx, provider, model, prompt, tokens, 1024, 0.3)
	if err != nil {
		return &reviewResult{Approved: true, Feedback: fmt.Sprintf("Review timeout/error: %v", err)}, nil
	}

	var review reviewResult
	if err := json.Unmarshal([]byte(util.ExtractJSONFromMarkdown(resp)), &review); err != nil {
		return &reviewResult{Approved: true, Feedback: "Could not parse review, auto-approved"}, nil
	}
	return &review, nil
}
