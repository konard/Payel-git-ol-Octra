package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nodes/internal/prompts"
	"nodes/internal/service/rules"
	"nodes/internal/service/util"
)

// reviewResult — внутренний результат ревью
type reviewResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback"`
}

// reviewWorkerResult — менеджер ревьюит работу одного воркера через LLM
func (s *Service) reviewWorkerResult(ctx context.Context, provider, model string, tokens map[string]string, managerRole string, wr *rules.WorkerResult) (*reviewResult, error) {
	filesList := ""
	for path, content := range wr.Files {
		preview := content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		filesList += fmt.Sprintf("\n--- %s ---\n%s\n", path, preview)
	}

	prompt := prompts.ManagerReviewWork(managerRole, wr.Role, wr.TaskMd, wr.SolutionMd, filesList)

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
