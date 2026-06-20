package boss

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/util"
	"orchestrator/pkg/models"

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
			bridged(role.Role, 40+idx*5, "Starting manager: "+role.Role, map[string]string{"task_type": decision.TaskType})

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
				PredefinedWorkers:    predefinedWorkersFor(decision, idx, role.Role),
			}

			managerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			defer cancel()

			managerProgress := func(p int32, msg string, data map[string]string) {
				scaled := 40 + idx*10 + int(p)*30/100
				bridged(role.Role, scaled, msg, data)
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
			bridged(role.Role, 70, "Manager completed: "+role.Role, map[string]string{"task_type": decision.TaskType})
		}(i, role)
	}

	wg.Wait()
	if firstErr != nil && len(allResults) == 0 {
		return nil, firstErr
	}
	return allResults, nil
}

func predefinedWorkersFor(decision *DecisionResult, idx int, role string) []*rules.WorkerRole {
	if decision == nil || len(decision.ManagerWorkflows) == 0 {
		return nil
	}

	var workflow *ManagerWorkflow
	if idx >= 0 && idx < len(decision.ManagerWorkflows) {
		workflow = &decision.ManagerWorkflows[idx]
	}
	if workflow == nil || !sameWorkflowRole(workflow.Role, role) {
		for i := range decision.ManagerWorkflows {
			if sameWorkflowRole(decision.ManagerWorkflows[i].Role, role) {
				workflow = &decision.ManagerWorkflows[i]
				break
			}
		}
	}
	if workflow == nil {
		return nil
	}

	workers := make([]*rules.WorkerRole, 0, len(workflow.Workers))
	for _, worker := range workflow.Workers {
		if worker.Role == "" {
			continue
		}
		// Воркер-нода поиска (issue #97) не генерирует файлы — её обрабатывает
		// runSearchNode, поэтому в обычный пул воркеров она не попадает.
		if isSearchNodeRole(worker.Role) {
			continue
		}
		workers = append(workers, &rules.WorkerRole{
			Role:         worker.Role,
			Description:  worker.Description,
			CustomPrompt: worker.CustomPrompt,
		})
	}
	return workers
}

func sameWorkflowRole(a, b string) bool {
	normalize := func(v string) string {
		return strings.ToLower(strings.Join(strings.Fields(v), " "))
	}
	return normalize(a) == normalize(b)
}

// gradeWeight переводит оценку сложности босса (1-10) в шкалу 0-100, которую
// видит менеджер ("grade_weight: N/100"). Раньше здесь был хардкод "10", из-за
// чего босс и менеджер получали рассогласованные сигналы о сложности одной и той
// же задачи (issue #79). Теперь обе стороны исходят из одной оценки: менеджер
// видит grade*10, ровно как босс возвращает в JSON-шаблоне. Для неоценённых
// задач (grade<=0) сохраняем прежний минимум "10".
func gradeWeight(grade int) string {
	if grade <= 0 {
		return "10"
	}
	if grade > 10 {
		grade = 10
	}
	return strconv.Itoa(grade * 10)
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
		"tokens":           util.MarshalJSON(req.Tokens),
		"model":            req.Meta["model"],
		"provider":         req.Meta["provider"],
		"title":            req.Title,
		"description":      req.Description,
		"grade_weight":     gradeWeight(req.Grade),
		"tech_stack":       techStack,
		"task_type":        taskType,
		"skill_categories": req.Meta["skill_categories"],
	}
	for k, v := range req.Tokens {
		metadata[k] = v
	}
	return metadata
}

