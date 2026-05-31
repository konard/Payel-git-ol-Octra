package manager

import (
	"context"
	"fmt"
	"log"

	"nodes/internal/service/rules"
)

// reviewAll — менеджер ревьюит результаты всех воркеров.
// Если ревью не одобрено, менеджер просит воркера исправить решение.
// Возвращает текстовое резюме ревью для итогового ManagerResult.
func (s *Service) reviewAll(
	ctx context.Context,
	req *rules.AssignManagerRequest,
	provider, model string,
	tokens map[string]string,
	taskType string,
	results []*rules.WorkerResult,
) string {
	summary := ""
	for _, wr := range results {
		review, err := s.reviewWorkerResult(ctx, provider, model, tokens, req.Role, taskType, wr)
		if err != nil {
			log.Printf("Review of worker %s failed: %v", wr.Role, err)
			summary += fmt.Sprintf("\n%s: review_error - %v", wr.Role, err)
			continue
		}

		wr.Approved = review.Approved
		wr.Feedback = review.Feedback

		if review.Approved {
			summary += fmt.Sprintf("\n%s: approved - %s", wr.Role, review.Feedback)
			continue
		}

		log.Printf("Worker %s not approved, asking for fix: %s", wr.Role, review.Feedback)
		reviewReq := &rules.ReviewRequest{
			TaskId:             req.TaskId,
			ManagerId:          req.ManagerId,
			WorkerId:           wr.WorkerId,
			WorkerRole:         wr.Role,
			Feedback:           review.Feedback,
			OriginalSolutionMd: wr.SolutionMd,
			OriginalFiles:      wr.Files,
			Metadata:           req.Metadata,
		}
		resp, err := s.worker.ReviewWorker(ctx, reviewReq)
		if err != nil {
			log.Printf("Worker %s fix failed: %v", wr.Role, err)
			summary += fmt.Sprintf("\n%s: fix_error - %v", wr.Role, err)
			continue
		}
		if resp.Status == "fixed" {
			if len(resp.Files) > 0 {
				wr.Files = resp.Files
			}
			if resp.SolutionMd != "" {
				wr.SolutionMd = resp.SolutionMd
			}
			wr.Approved = true
			summary += fmt.Sprintf("\n%s: fixed - %s", wr.Role, resp.Feedback)
		} else {
			summary += fmt.Sprintf("\n%s: not_fixed - %s", wr.Role, resp.Feedback)
		}
	}
	return summary
}
