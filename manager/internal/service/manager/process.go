package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"manager/internal/fetcher/grpc/managerpb"
	"manager/internal/fetcher/grpc/workerpb"
	"manager/pkg/database"
	"manager/pkg/models"
)

// processManager — основная логика одного менеджера
func (s *ManagerService) processManager(ctx context.Context, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, fmt.Errorf("invalid task_id: %w", err)
	}

	managerID, err := uuid.Parse(req.ManagerId)
	if err != nil {
		managerID = uuid.New()
	}

	// Создаём менеджера в БД
	manager := &models.Manager{
		ID:       managerID,
		TaskID:   taskID,
		Role:     req.Role,
		Status:   "active",
		TaskDesc: req.TechnicalDescription,
	}
	database.Db.Create(manager)

	// Парсим метаданные
	metadata := req.Metadata
	tokens := make(map[string]string)
	if tokensJSON, ok := metadata["tokens"]; ok {
		json.Unmarshal([]byte(tokensJSON), &tokens)
	}
	provider := metadata["provider"]
	if provider == "" {
		provider = "openai"
	}
	model := metadata["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Менеджер думает каких воркеров нанять (через agents сервис)
	workerRolesList, err := s.managerThink(ctx, provider, model, tokens, req.TechnicalDescription, req.Role, req.Description)
	if err != nil {
		log.Printf("Manager think error: %v", err)
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("manager think failed: %w", err)
	}

	// Преобразуем роли в proto формат
	workerRoles := make([]*workerpb.WorkerRole, 0, len(workerRolesList))
	for _, role := range workerRolesList {
		workerRoles = append(workerRoles, &workerpb.WorkerRole{
			Role:        role.Role,
			Description: role.Description,
		})
	}

	manager.WorkersCount = int32(len(workerRoles))
	workerRolesJSON, _ := json.Marshal(workerRolesList)
	manager.WorkerRoles = string(workerRolesJSON)
	database.Db.Save(manager)

	// Вызываем Worker сервис
	log.Printf("Manager %s calling workers: %v", req.Role, workerRolesList)

	// Конвертируем results других воркеров для worker сервиса
	otherResults := workerResultsFromManagerpb(req.OtherWorkersResults)

	workerReq := &workerpb.AssignWorkersRequest{
		TaskId:              req.TaskId,
		ManagerId:           req.ManagerId,
		ManagerRole:         req.Role,
		WorkerRoles:         workerRoles,
		TaskMd:              req.TechnicalDescription,
		Metadata:            metadata,
		OtherWorkersResults: otherResults,
	}

	workerResp, err := s.workerClient.AssignWorkersAndWait(ctx, workerReq)
	if err != nil {
		log.Printf("Worker call error: %v", err)
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("worker call failed: %w", err)
	}

	// Manager reviews each worker
	log.Printf("Manager %s reviewing workers...", req.Role)
	reviewSummary := ""
	for _, wr := range workerResp.WorkerResults {
		// Manager reviews worker via AI
		reviewResult, err := s.reviewWorkerResult(ctx, provider, model, tokens, req.Role, wr)
		if err != nil {
			log.Printf("Review error for %s: %v", wr.Role, err)
			reviewSummary += fmt.Sprintf("\n%s: review error - %v", wr.Role, err)
			continue
		}

		if !reviewResult.Approved {
			// Send back for fixes
			log.Printf("Worker %s NOT approved, sending for review...", wr.Role)
			reviewReq := &workerpb.ReviewRequest{
				TaskId:             req.TaskId,
				ManagerId:          req.ManagerId,
				WorkerId:           wr.WorkerId,
				WorkerRole:         wr.Role,
				Feedback:           reviewResult.Feedback,
				OriginalSolutionMd: wr.SolutionMd,
				OriginalFiles:      wr.Files,
				Metadata:           metadata,
			}

			reviewResp, err := s.workerClient.ReviewWorker(ctx, reviewReq)
			if err != nil {
				log.Printf("ReviewWorker call error: %v", err)
				reviewSummary += fmt.Sprintf("\n%s: review call failed - %v", wr.Role, err)
			} else {
				reviewSummary += fmt.Sprintf("\n%s: %s - %s", wr.Role, reviewResp.Status, reviewResp.Feedback)
				// Update worker files with reviewed version
				if reviewResp.Status == "fixed" {
					for path, content := range reviewResp.Files {
						wr.Files[path] = content
					}
					wr.SolutionMd = reviewResp.SolutionMd
					wr.Approved = true
				}
			}
		} else {
			reviewSummary += fmt.Sprintf("\n%s: approved", wr.Role)
			wr.Approved = true
		}
	}

	// Build final ZIP from reviewed worker results
	allFiles := make(map[string]string)
	for _, wr := range workerResp.WorkerResults {
		for path, content := range wr.Files {
			allFiles[path] = content
		}
	}

	finalZip, _ := createZipArchive(allFiles)

	manager.Status = "done"
	database.Db.Save(manager)

	log.Printf("Manager %s completed: %d workers, review: %s", req.Role, len(workerResp.WorkerResults), reviewSummary)

	return &managerpb.ManagerResult{
		TaskId:        req.TaskId,
		ManagerId:     req.ManagerId,
		Role:          req.Role,
		Status:        "success",
		Solution:      finalZip,
		WorkerResults: workerResultsToManagerpb(workerResp.WorkerResults),
		ReviewSummary: reviewSummary,
	}, nil
}
