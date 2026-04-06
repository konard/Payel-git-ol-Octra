package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"worker/internal/fetcher/grpc/workerpb"
	"worker/pkg/database"
	"worker/pkg/models"

	"github.com/google/uuid"
)

// AssignWorkersAndWait accepts task, generates code via AI agents and returns ZIP
func (s *WorkerService) AssignWorkersAndWait(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	log.Printf("Received task from manager %s (%s): %s", req.ManagerId, req.ManagerRole, req.TaskId)
	log.Printf("Worker roles: %v", req.WorkerRoles)

	metadata := req.Metadata
	provider := metadata["provider"]
	if provider == "" {
		provider = "openai"
	}
	model := metadata["model"]
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Parse tokens
	tokens := make(map[string]string)
	if tokensJSON, ok := metadata["tokens"]; ok {
		json.Unmarshal([]byte(tokensJSON), &tokens)
	}
	// Also pass individual provider key
	if apiKey, ok := metadata[provider]; ok {
		tokens[provider] = apiKey
	}

	// Build context from other workers results
	contextSummary := ""
	for _, wr := range req.OtherWorkersResults {
		contextSummary += fmt.Sprintf("\n--- Worker %s (%s) result ---\n", wr.WorkerId, wr.Role)
		contextSummary += fmt.Sprintf("Task: %s\n", wr.TaskMd)
		if wr.Success {
			contextSummary += fmt.Sprintf("Solution: %s\n", wr.SolutionMd)
		}
		if wr.Feedback != "" {
			contextSummary += fmt.Sprintf("Feedback: %s\n", wr.Feedback)
		}
	}

	// Collect all project files
	allFiles := make(map[string]string)
	workerResults := make([]*workerpb.WorkerResult, 0)

	projectName := sanitizeProjectName(req.ManagerRole)
	basePath := fmt.Sprintf("project/%s", projectName)

	// Process each worker role SEQUENTIALLY so they can see each other's results
	for i, wr := range req.WorkerRoles {
		role := wr.Role
		description := wr.Description

		// Create worker in DB
		workerID := uuid.New()
		worker := &models.Worker{
			ID:        workerID,
			TaskID:    uuid.MustParse(req.TaskId),
			ManagerID: uuid.MustParse(req.ManagerId),
			Role:      role,
			Status:    "thinking",
			TaskMD:    req.TaskMd,
		}

		if err := database.Db.Create(worker).Error; err != nil {
			log.Printf("Error creating worker: %v", err)
			continue
		}

		// Build accumulated context from previous workers in THIS team
		s.mu.Lock()
		accumulatedContext := ""
		for _, prevResult := range workerResults {
			accumulatedContext += fmt.Sprintf("\n=== Previous worker %s (%s) ===\n", prevResult.WorkerId, prevResult.Role)
			if prevResult.Success {
				for path, content := range prevResult.Files {
					accumulatedContext += fmt.Sprintf("\n--- File: %s ---\n%s\n", path, content)
				}
			}
			if prevResult.Feedback != "" {
				accumulatedContext += fmt.Sprintf("\nFeedback: %s\n", prevResult.Feedback)
			}
		}
		s.mu.Unlock()

		// Add external workers context (from OTHER managers)
		if contextSummary != "" {
			accumulatedContext += "\n=== Context from other worker groups ===\n" + contextSummary
		}

		// Worker thinks about task - creates TASK.md via agents service
		taskMD, err := s.createTaskMD(ctx, provider, model, tokens, role, description, req.TaskMd, accumulatedContext, basePath)
		if err != nil {
			log.Printf("AI error (think): %v", err)
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		worker.Status = "coding"
		database.Db.Save(worker)

		// Generate code via agents service
		// Check WORKER_MODE env var: "multypass" (optimized) or "nplus1" (legacy)
		workerMode := os.Getenv("WORKER_MODE")
		if workerMode == "" {
			workerMode = "multypass" // Default to optimized mode
		}

		var files map[string]string
		if workerMode == "multypass" {
			files, err = s.generateCodeMultiPass(ctx, provider, model, tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext)
		} else {
			// Fallback to legacy N+1 approach
			files, err = s.generateCode(ctx, provider, model, tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext)
		}

		if err != nil {
			log.Printf("AI error (generate): %v", err)
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		// Collect files - add worker prefix to avoid collisions
		for path, content := range files {
			// Each worker's files go to their own subdirectory
			// e.g., frontend/react/, frontend/ui/, backend/api/
			allFiles[path] = content
		}

		// Build worker result
		filesJSON := make(map[string]string)
		for k, v := range files {
			filesJSON[k] = v
		}

		workerResult := &workerpb.WorkerResult{
			WorkerId:   workerID.String(),
			Role:       role,
			TaskMd:     taskMD,
			SolutionMd: fmt.Sprintf("Generated %d files for role %s", len(files), role),
			Files:      filesJSON,
			Success:    true,
			Approved:   false,
		}

		worker.SolutionMD = workerResult.SolutionMd
		worker.Files = marshalJSON(files)
		worker.Status = "done"
		worker.Success = true
		database.Db.Save(worker)

		s.mu.Lock()
		workerResults = append(workerResults, workerResult)
		s.mu.Unlock()

		log.Printf("Worker %d/%d (%s) completed: %d files", i+1, len(req.WorkerRoles), role, len(files))
	}

	// Create ZIP archive
	zipData, err := createZipArchive(allFiles)
	if err != nil {
		log.Printf("Error creating ZIP: %v", err)
		zipData = []byte{}
	}

	log.Printf("Created ZIP archive: %d bytes", len(zipData))

	return &workerpb.AssignWorkersResponse{
		TaskId:        req.TaskId,
		Status:        "success",
		Solution:      zipData,
		WorkerResults: workerResults,
	}, nil
}
