package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"nodes/internal/service/rules"
)

// AssignWorkersAndWait — последовательно прогоняет всех воркеров команды менеджера.
// Раньше это был gRPC-вызов manager → worker. Теперь — прямой Go-вызов.
func (s *Service) AssignWorkersAndWait(ctx context.Context, req *rules.AssignWorkersRequest, progress rules.ProgressFunc) (*rules.AssignWorkersResponse, error) {
	log.Printf("Received task from manager %s (%s): %s", req.ManagerId, req.ManagerRole, req.TaskId)

	meta := parseWorkerMetadata(req.Metadata)
	contextSummary := buildContextSummary(req.OtherWorkersResults)
	basePath := resolveBasePath(req)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory %s: %w", basePath, err)
	}
	writeContextFile(basePath, req.TaskId, req.ManagerId, req.ManagerRole, req.WorkerRoles, contextSummary)

	if progress != nil {
		progress(0, "Workers assigned and starting work", map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
		})
	}

	workerResults := make([]*rules.WorkerResult, 0, len(req.WorkerRoles))
	for i, wr := range req.WorkerRoles {
		if progress != nil {
			p := int32(10 + (i * 80 / max1(len(req.WorkerRoles))))
			progress(p, fmt.Sprintf("Starting worker %d/%d: %s", i+1, len(req.WorkerRoles), wr.Role), nil)
		}

		s.mu.Lock()
		accumulatedContext := buildAccumulatedContext(workerResults, contextSummary)
		s.mu.Unlock()

		result, err := s.runOneWorker(ctx, req, wr, meta, basePath, accumulatedContext)
		if err != nil {
			log.Printf("Worker %s error: %v", wr.Role, err)
			continue
		}

		s.mu.Lock()
		workerResults = append(workerResults, result)
		s.mu.Unlock()

		if progress != nil {
			p := int32(20 + ((i + 1) * 80 / max1(len(req.WorkerRoles))))
			fileList := []string{}
			for k := range result.Files {
				fileList = append(fileList, k)
			}
			data := map[string]string{
				"manager_id":   req.ManagerId,
				"manager_role": req.ManagerRole,
				"status":       "processing",
				"files":        strings.Join(fileList, ","),
				"files_count":  strconv.Itoa(len(result.Files)),
			}
			if codeFiles := buildCodeFilesPayload(result.Files, wr.Role, req.ManagerRole); codeFiles != "" {
				data["code_files"] = codeFiles
			}
			progress(p, fmt.Sprintf("Worker %d/%d (%s) completed: %d files", i+1, len(req.WorkerRoles), wr.Role, len(result.Files)), data)
		}
	}

	if progress != nil {
		progress(100, "All workers completed successfully", map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
		})
	}

	return &rules.AssignWorkersResponse{
		TaskId:        req.TaskId,
		Status:        "success",
		WorkerResults: workerResults,
	}, nil
}

// buildAccumulatedContext — собирает контекст из уже завершённых воркеров текущей команды
func buildAccumulatedContext(workerResults []*rules.WorkerResult, externalCtx string) string {
	acc := ""
	for _, prevResult := range workerResults {
		acc += fmt.Sprintf("\n=== Previous worker %s (%s) ===\n", prevResult.WorkerId, prevResult.Role)
		if prevResult.Success {
			for path, content := range prevResult.Files {
				acc += fmt.Sprintf("\n--- File: %s ---\n%s\n", path, content)
			}
		}
		if prevResult.Feedback != "" {
			acc += fmt.Sprintf("\nFeedback: %s\n", prevResult.Feedback)
		}
	}
	if externalCtx != "" {
		acc += "\n=== Context from other worker groups ===\n" + externalCtx
	}
	return acc
}

// max1 — защита от деления на 0 в progress-расчёте
func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
