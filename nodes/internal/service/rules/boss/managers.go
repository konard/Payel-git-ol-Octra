package boss

import (
	"context"
	"log"
	"sync"
	"time"

	"nodes/internal/service/rules"
	"nodes/internal/service/util"
	"nodes/pkg/models"

	"github.com/google/uuid"
)

// assignManagersParallel — параллельно запускает всех менеджеров.
// Раньше каждый вызов уходил по gRPC в отдельный manager-сервис.
// Теперь это прямой Go-вызов s.manager.AssignManager в той же программе.
func (s *Service) assignManagersParallel(
	ctx context.Context,
	taskID string,
	decision *DecisionResult,
	req *CreateTaskRequest,
	progress rules.ProgressFunc,
	projectPath string,
) ([]*rules.ManagerResult, error) {
	metadata := buildManagerMetadata(req, decision)

	var (
		mu         sync.Mutex
		allResults []*rules.ManagerResult
		firstErr   error
	)

	var wg sync.WaitGroup
	bridged := s.bridge(progress)

	for i, role := range decision.ManagerRoles {
		wg.Add(1)
		go func(idx int, role models.ManagerRole) {
			defer wg.Done()
			bridged(role.Role, 40+idx*5, "Starting manager: "+role.Role)

			if idx > 0 {
				select {
				case <-time.After(time.Duration(idx) * 5 * time.Second):
				case <-ctx.Done():
					return
				}
			}

			mu.Lock()
			var contextResults []*rules.WorkerResult
			for _, mr := range allResults {
				contextResults = append(contextResults, mr.WorkerResults...)
			}
			mu.Unlock()

			managerReq := &rules.AssignManagerRequest{
				TaskId:               taskID,
				ManagerId:            uuid.New().String(),
				Role:                 role.Role,
				Description:          role.Description,
				CustomPrompt:         role.CustomPrompt,
				TechnicalDescription: decision.TechnicalDescription,
				ProjectPath:          projectPath,
				Metadata:             metadata,
				OtherWorkersResults:  contextResults,
			}

			managerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			defer cancel()

			managerProgress := func(p int32, msg string, data map[string]string) {
				scaled := 40 + idx*10 + int(p)*30/100
				bridged(role.Role, scaled, msg)
			}

			log.Printf("Calling Manager #%d: %s", idx+1, role.Role)
			result, err := s.manager.AssignManager(managerCtx, managerReq, managerProgress)
			if err != nil {
				log.Printf("Manager %s failed: %v", role.Role, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			allResults = append(allResults, result)
			mu.Unlock()
			bridged(role.Role, 70, "Manager completed: "+role.Role)
		}(i, role)
	}

	wg.Wait()
	if firstErr != nil && len(allResults) == 0 {
		return nil, firstErr
	}
	return allResults, nil
}

// buildManagerMetadata — формирует метадату для менеджеров
func buildManagerMetadata(req *CreateTaskRequest, decision *DecisionResult) map[string]string {
	techStack := "go"
	if len(decision.TechStack) > 0 {
		techStack = decision.TechStack[0]
	}
	taskType := decision.TaskType
	if taskType == "" {
		taskType = "code"
	}
	metadata := map[string]string{
		"tokens":       util.MarshalJSON(req.Tokens),
		"model":        req.Meta["model"],
		"provider":     req.Meta["provider"],
		"title":        req.Title,
		"description":  req.Description,
		"grade_weight": "10",
		"tech_stack":   techStack,
		"task_type":    taskType,
	}
	for k, v := range req.Tokens {
		metadata[k] = v
	}
	return metadata
}
