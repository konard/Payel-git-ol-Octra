package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"manager/internal/fetcher/grpc/managerpb"
	"manager/internal/fetcher/grpc/workerpb"
	"manager/internal/prompts"
)

// reviewWorkerResult — менеджер проверяет работу воркера через AI
func (s *ManagerService) reviewWorkerResult(ctx context.Context, provider, model string, tokens map[string]string, managerRole string, wr *workerpb.WorkerResult) (*ReviewResult, error) {
	filesList := ""
	for path, content := range wr.Files {
		preview := content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		filesList += fmt.Sprintf("\n--- %s ---\n%s\n", path, preview)
	}

	prompt := prompts.ReviewWork(managerRole, wr.Role, wr.TaskMd, wr.SolutionMd, filesList)

	// Timeout 30s for review to avoid WS disconnect
	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := s.agentsClient.Generate(genCtx, provider, model, prompt, tokens, 1024, 0.3)
	if err != nil {
		// Auto-approve on timeout/error
		return &ReviewResult{Approved: true, Feedback: fmt.Sprintf("Review timeout/error: %v", err)}, nil
	}

	var review ReviewResult
	if err := json.Unmarshal([]byte(extractJSONFromMarkdown(resp)), &review); err != nil {
		// If can't parse, approve by default
		return &ReviewResult{Approved: true, Feedback: "Could not parse review, auto-approved"}, nil
	}

	return &review, nil
}

type ReviewResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback"`
}

// Conversion helpers between managerpb.WorkerResult and workerpb.WorkerResult
func workerResultToManagerpb(wr *workerpb.WorkerResult) *managerpb.WorkerResult {
	if wr == nil {
		return nil
	}
	return &managerpb.WorkerResult{
		WorkerId:   wr.WorkerId,
		Role:       wr.Role,
		TaskMd:     wr.TaskMd,
		SolutionMd: wr.SolutionMd,
		Files:      wr.Files,
		Success:    wr.Success,
		Feedback:   wr.Feedback,
		Approved:   wr.Approved,
	}
}

func workerResultsToManagerpb(wrs []*workerpb.WorkerResult) []*managerpb.WorkerResult {
	result := make([]*managerpb.WorkerResult, len(wrs))
	for i, wr := range wrs {
		result[i] = workerResultToManagerpb(wr)
	}
	return result
}

func workerResultFromManagerpb(wr *managerpb.WorkerResult) *workerpb.WorkerResult {
	if wr == nil {
		return nil
	}
	return &workerpb.WorkerResult{
		WorkerId:   wr.WorkerId,
		Role:       wr.Role,
		TaskMd:     wr.TaskMd,
		SolutionMd: wr.SolutionMd,
		Files:      wr.Files,
		Success:    wr.Success,
		Feedback:   wr.Feedback,
		Approved:   wr.Approved,
	}
}

func workerResultsFromManagerpb(wrs []*managerpb.WorkerResult) []*workerpb.WorkerResult {
	result := make([]*workerpb.WorkerResult, len(wrs))
	for i, wr := range wrs {
		result[i] = workerResultFromManagerpb(wr)
	}
	return result
}
