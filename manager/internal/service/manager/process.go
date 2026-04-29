package manager

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/google/uuid"
  "log"
  "manager/internal/fetcher/grpc/managerpb"
  "manager/internal/fetcher/grpc/workerpb"
  "manager/internal/service/git"
  "manager/pkg/database"
  "manager/pkg/models"
  "os"
  "path/filepath"
  "time"
)

// processManager — основная логика одного менеджера
func (s *ManagerService) processManager(ctx context.Context, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
	return s.processManagerWithProgress(ctx, req, nil)
}

// processManagerWithProgress — основная логика одного менеджера с progress callback
func (s *ManagerService) processManagerWithProgress(ctx context.Context, req *managerpb.AssignManagerRequest, progressCallback func(int, string)) (*managerpb.ManagerResult, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, fmt.Errorf("invalid task_id: %w", err)
	}

	managerID, err := uuid.Parse(req.ManagerId)
	if err != nil {
		managerID = uuid.New()
	}

	// Создаём менеджера в БД (или обновляем если уже существует)
	manager := &models.Manager{
		ID:           managerID,
		TaskID:       taskID,
		Role:         req.Role,
		Status:       "active",
		TaskDesc:     req.TechnicalDescription,
		CustomPrompt: req.CustomPrompt,
	}
	result := database.Db.Create(manager)
	if result.Error != nil {
		log.Printf("Warning: manager create failed, trying update: %v", result.Error)
		database.Db.Save(manager)
	}

  // Create/switch to manager branch
  if req.ProjectPath != "" {
    if err := s.switchToManagerBranch(req.ProjectPath, req.Role); err != nil {
      log.Printf("Warning: failed to create manager branch: %v", err)
    }
  }

  if progressCallback != nil {
    progressCallback(5, fmt.Sprintf("Manager %s initialized", req.Role))
  }

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
	if progressCallback != nil {
		progressCallback(10, fmt.Sprintf("Manager %s thinking about workers...", req.Role))
	}

	workerRolesList, err := s.managerThink(ctx, provider, model, tokens, req.TechnicalDescription, req.Role, req.Description)
	if err != nil {
		log.Printf("Manager think error: %v", err)
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("manager think failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback(20, fmt.Sprintf("Manager %s decided to hire %d workers", req.Role, len(workerRolesList)))
	}

	// Преобразуем роли в proto формат
	workerRoles := make([]*workerpb.WorkerRole, 0, len(workerRolesList))
	for _, role := range workerRolesList {
		workerRoles = append(workerRoles, &workerpb.WorkerRole{
			Role:         role.Role,
			Description:  role.Description,
			CustomPrompt: role.CustomPrompt,
		})
	}

	manager.WorkersCount = int32(len(workerRoles))
	workerRolesJSON, _ := json.Marshal(workerRolesList)
	manager.WorkerRoles = string(workerRolesJSON)
	database.Db.Save(manager)

	// Update .crewai/context.json
	projectPath := req.ProjectPath
	if projectPath != "" {
		crewaiDir := filepath.Join(projectPath, ".crewai")
		contextPath := filepath.Join(crewaiDir, "context.json")
		var contextData map[string]interface{}
		if data, err := os.ReadFile(contextPath); err == nil {
			json.Unmarshal(data, &contextData)
		} else {
			contextData = map[string]interface{}{
				"task_id": req.TaskId,
				"history": []map[string]interface{}{},
			}
		}
		// Add manager decision
		history, ok := contextData["history"].([]map[string]interface{})
		if !ok {
			history = []map[string]interface{}{}
		}
		history = append(history, map[string]interface{}{
			"type":      "manager_decision",
			"manager":   req.Role,
			"workers":   workerRolesList,
			"timestamp": time.Now().Unix(),
		})
		contextData["history"] = history
		contextJSON, _ := json.MarshalIndent(contextData, "", "  ")
		os.MkdirAll(crewaiDir, 0755)
		os.WriteFile(contextPath, contextJSON, 0644)
	}

	// Вызываем Worker сервис
	log.Printf("Manager %s calling workers: %v", req.Role, workerRolesList)

	// Конвертируем results других воркеров для worker сервиса
otherResults := workerResultsFromManagerpb(req.OtherWorkersResults)

  // Add repo path to metadata for Git operations
  workerMetadata := make(map[string]string)
  for k, v := range metadata {
workerMetadata[k] = v
  }
   
  // Use project path as-is (Boss already added title subfolder)
  workerProjectPath := req.ProjectPath
   
  if workerProjectPath != "" {
    workerMetadata["repo_path"] = workerProjectPath
  }
  
  workerReq := &workerpb.AssignWorkersRequest{
    TaskId:              req.TaskId,
    ManagerId:           req.ManagerId,
    ManagerRole:         req.Role,
    WorkerRoles:         workerRoles,
    TaskMd:              req.TechnicalDescription,
    Metadata:            workerMetadata,
    OtherWorkersResults: otherResults,
    ProjectPath:        req.ProjectPath,
  }

	if progressCallback != nil {
		progressCallback(30, fmt.Sprintf("Manager %s starting workers...", req.Role))
	}

	// Use streaming version to get progress updates
	stream, err := s.workerClient.AssignWorkersAndWaitStream(ctx, workerReq)
	if err != nil {
		log.Printf("Worker stream call error: %v", err)
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("worker stream call failed: %w", err)
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Worker stream error: %v", err)
			manager.Status = "error"
			database.Db.Save(manager)
			return nil, fmt.Errorf("worker stream failed: %w", err)
		}

		// Forward progress updates
		if progressCallback != nil && update.Message != "" {
			// Calculate progress based on worker completion
			progress := 30 + int(update.Progress*70/100) // 30-100 range
			progressCallback(progress, update.Message)
		}

		// If this is a completion update, collect results
		if update.Status == "success" || update.Status == "error" {
			// For now, we'll get final results via non-streaming call
			// TODO: Modify worker to send results in stream
			break
		}
	}

	// Get final results
	workerResp, err := s.workerClient.AssignWorkersAndWait(ctx, workerReq)
	if err != nil {
		log.Printf("Worker final call error: %v", err)
		manager.Status = "error"
		database.Db.Save(manager)
		return nil, fmt.Errorf("worker final call failed: %w", err)
	}

	// Merge all worker branches into manager branch before review
	if req.ProjectPath != "" {
		log.Printf("Merging worker branches into manager branch...")
		for _, wr := range workerResp.WorkerResults {
			branchName := fmt.Sprintf("worker-%s", wr.Role)
			mergeMessage := fmt.Sprintf("Merge worker %s branch", wr.Role)
			if err := git.MergeBranch(req.ProjectPath, branchName, mergeMessage); err != nil {
				log.Printf("Failed to merge worker branch %s: %v", branchName, err)
			} else {
				log.Printf("Merged worker branch: %s", branchName)
			}
		}
	}

	// Manager reviews each worker
	log.Printf("Manager %s reviewing workers...", req.Role)
	if progressCallback != nil {
		progressCallback(95, fmt.Sprintf("Manager %s reviewing workers...", req.Role))
	}
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
	manager.Status = "done"
	database.Db.Save(manager)

	log.Printf("Manager %s completed: %d workers, review: %s", req.Role, len(workerResp.WorkerResults), reviewSummary)

	if progressCallback != nil {
		progressCallback(100, fmt.Sprintf("Manager %s completed successfully", req.Role))
	}

  return &managerpb.ManagerResult{
    TaskId:        req.TaskId,
    ManagerId:     req.ManagerId,
    Role:          req.Role,
    Status:        "success",
    WorkerResults: workerResultsToManagerpb(workerResp.WorkerResults),
    ReviewSummary: reviewSummary,
  }, nil
}

// switchToManagerBranch creates and switches to a branch for the manager
func (s *ManagerService) switchToManagerBranch(repoPath, role string) error {
  branchName := fmt.Sprintf("manager-%s", role)

  // Create and switch to branch
  if err := git.CreateBranch(repoPath, branchName); err != nil {
    return err
  }

  log.Printf("✅ Switched to branch: %s", branchName)
  return nil
}
