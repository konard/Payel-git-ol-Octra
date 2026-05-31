package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"nodes/internal/service/rules"
	"nodes/pkg/database"
	"nodes/pkg/models"

	"github.com/google/uuid"
)

// AssignManager — основной метод менеджера: think → workers → review.
// Раньше это было gRPC-эндпоинтом, теперь — обычная Go-функция,
// вызываемая боссом из того же процесса.
func (s *Service) AssignManager(ctx context.Context, req *rules.AssignManagerRequest, progress rules.ProgressFunc) (*rules.ManagerResult, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, fmt.Errorf("invalid task_id: %w", err)
	}
	managerID, err := uuid.Parse(req.ManagerId)
	if err != nil {
		managerID = uuid.New()
	}

	manager := &models.Manager{
		ID:           managerID,
		TaskID:       taskID,
		Role:         req.Role,
		Status:       "active",
		TaskDesc:     req.TechnicalDescription,
		CustomPrompt: req.CustomPrompt,
	}
	if res := database.Db.Create(manager); res.Error != nil {
		log.Printf("Warning: manager create failed, trying update: %v", res.Error)
		database.Db.Save(manager)
	}

	if req.ProjectPath != "" {
		if err := switchToManagerBranch(req.ProjectPath, req.Role); err != nil {
			log.Printf("Warning: failed to create manager branch: %v", err)
		}
	}
	emit(progress, 5, fmt.Sprintf("Manager %s initialized", req.Role), nil)

	provider, model, tokens, gradeWeight, taskType := parseMetadata(req.Metadata)

	emit(progress, 10, fmt.Sprintf("Manager %s thinking about workers...", req.Role), nil)

	workerRolesList, err := s.think(ctx, provider, model, tokens, req.TechnicalDescription, req.Role, req.Description, gradeWeight, taskType)
	if err != nil {
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("manager think failed: %w", err)
	}

	emit(progress, 20, fmt.Sprintf("Manager %s decided to hire %d workers", req.Role, len(workerRolesList)), nil)

	workerRoles := make([]*rules.WorkerRole, 0, len(workerRolesList))
	for _, r := range workerRolesList {
		workerRoles = append(workerRoles, &rules.WorkerRole{
			Role:         r.Role,
			Description:  r.Description,
			CustomPrompt: r.CustomPrompt,
		})
	}
	manager.WorkersCount = int32(len(workerRoles))
	workerRolesJSON, _ := json.Marshal(workerRolesList)
	manager.WorkerRoles = string(workerRolesJSON)
	database.Db.Save(manager)

	updateContextFile(req.ProjectPath, req.TaskId, req.Role, workerRolesList)

	workerMetadata := make(map[string]string, len(req.Metadata)+1)
	for k, v := range req.Metadata {
		workerMetadata[k] = v
	}
	if req.ProjectPath != "" {
		workerMetadata["repo_path"] = req.ProjectPath
	}

	workerReq := &rules.AssignWorkersRequest{
		TaskId:              req.TaskId,
		ManagerId:           req.ManagerId,
		ManagerRole:         req.Role,
		WorkerRoles:         workerRoles,
		TaskMd:              req.TechnicalDescription,
		Metadata:            workerMetadata,
		OtherWorkersResults: req.OtherWorkersResults,
		ProjectPath:         req.ProjectPath,
	}
	emit(progress, 30, fmt.Sprintf("Manager %s starting workers...", req.Role), nil)

	workerProgress := func(p int32, msg string, data map[string]string) {
		if progress != nil {
			scaled := int32(30 + int(p)*70/100)
			progress(scaled, msg, data)
		}
	}
	workerResp, err := s.worker.AssignWorkersAndWait(ctx, workerReq, workerProgress)
	if err != nil {
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("worker call failed: %w", err)
	}

	if req.ProjectPath != "" {
		mergeWorkerBranches(req.ProjectPath, workerResp.WorkerResults)
	}

	emit(progress, 95, fmt.Sprintf("Manager %s reviewing workers...", req.Role), nil)
	reviewSummary := s.reviewAll(ctx, req, provider, model, tokens, taskType, workerResp.WorkerResults)

	manager.Status = "done"
	database.Db.Save(manager)
	emit(progress, 100, fmt.Sprintf("Manager %s completed successfully", req.Role), nil)

	return &rules.ManagerResult{
		TaskId:        req.TaskId,
		ManagerId:     req.ManagerId,
		Role:          req.Role,
		Status:        "success",
		WorkerResults: workerResp.WorkerResults,
		ReviewSummary: reviewSummary,
	}, nil
}

// emit — отправляет апдейт прогресса, если callback задан
func emit(progress rules.ProgressFunc, p int32, msg string, data map[string]string) {
	if progress != nil {
		progress(p, msg, data)
	}
}

// parseMetadata — стандартный парсинг metadata менеджера
func parseMetadata(metadata map[string]string) (provider, model string, tokens map[string]string, gradeWeight, taskType string) {
	tokens = make(map[string]string)
	if tokensJSON, ok := metadata["tokens"]; ok {
		json.Unmarshal([]byte(tokensJSON), &tokens)
	}
	provider = metadata["provider"]
	if provider == "" {
		provider = "openai"
	}
	model = metadata["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}
	gradeWeight = "10"
	if gw, ok := metadata["grade_weight"]; ok && gw != "" {
		gradeWeight = gw
	}
	taskType = metadata["task_type"]
	if taskType == "" {
		taskType = "code"
	}
	return
}
