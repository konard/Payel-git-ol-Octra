package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"worker/internal/fetcher/grpc/workerpb"
	"worker/internal/service/git"
	"worker/pkg/database"
	"worker/pkg/models"

	"github.com/google/uuid"
)

// AssignWorkersAndWait accepts task, generates code via AI agents and returns ZIP
func (s *WorkerService) AssignWorkersAndWait(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	return s.assignWorkersAndWaitWithProgress(ctx, req, nil)
}

// AssignWorkersAndWaitStream sends progress updates while processing workers
func (s *WorkerService) AssignWorkersAndWaitStream(req *workerpb.AssignWorkersRequest, stream workerpb.WorkerService_AssignWorkersAndWaitStreamServer) error {
	ctx := stream.Context()

	// Send initial progress
	stream.Send(&workerpb.TaskUpdate{
		TaskId:    req.TaskId,
		Message:   "Workers assigned and starting work",
		Progress:  0,
		Status:    "processing",
		Timestamp: 0,
		Data: map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
		},
	})

	progressCallback := func(progress int, message string, files map[string]string) {
		data := map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
			"status":       "processing",
		}
		// Add file list (not contents - too large for streaming)
		if len(files) > 0 {
			fileList := []string{}
			for k := range files {
				fileList = append(fileList, k)
			}
			data["files"] = strings.Join(fileList, ",")
			data["files_count"] = strconv.Itoa(len(files))
		}
		stream.Send(&workerpb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   message,
			Progress:  int32(progress),
			Status:    "processing",
			Timestamp: 0,
			Data:      data,
		})
	}

	_, err := s.assignWorkersAndWaitWithProgress(ctx, req, progressCallback)
	if err != nil {
		stream.Send(&workerpb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   fmt.Sprintf("Workers failed: %v", err),
			Progress:  0,
			Status:    "error",
			Timestamp: 0,
		})
		return err
	}

	// Send final success
	stream.Send(&workerpb.TaskUpdate{
		TaskId:    req.TaskId,
		Message:   "All workers completed successfully",
		Progress:  100,
		Status:    "success",
		Timestamp: 0,
		Data: map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
		},
	})

	return nil
}

// assignWorkersAndWaitWithProgress is the core implementation with optional progress callback
func (s *WorkerService) assignWorkersAndWaitWithProgress(ctx context.Context, req *workerpb.AssignWorkersRequest, progressCallback func(int, string, map[string]string)) (*workerpb.AssignWorkersResponse, error) {
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

	basePath := req.ProjectPath
	if basePath == "" {
		// Check metadata for repo_path (set by manager)
		if repoPath, ok := req.Metadata["repo_path"]; ok && repoPath != "" {
			basePath = repoPath
		} else {
			projectsDir := os.Getenv("PROJECTS_DIR")
			if projectsDir == "" {
				projectsDir = "/workspace/projects"
			}
			// If we only have projects dir, append taskID to get project-specific path
			if filepath.Base(projectsDir) == "projects" || projectsDir == "/workspace/projects" {
				basePath = filepath.Join(projectsDir, req.TaskId)
			} else {
				basePath = projectsDir
			}
		}
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory %s: %w", basePath, err)
	}

	// Create .crewai/context.json
	contextData := map[string]interface{}{
		"task_id": req.TaskId,
		"manager": map[string]string{
			"id":   req.ManagerId,
			"role": req.ManagerRole,
		},
		"workers": req.WorkerRoles,
		"context": contextSummary,
		"history": []map[string]interface{}{},
	}
	contextJSON, _ := json.MarshalIndent(contextData, "", "  ")
	crewaiDir := filepath.Join(basePath, ".crewai")
	os.MkdirAll(crewaiDir, 0755)
	os.WriteFile(filepath.Join(crewaiDir, "context.json"), contextJSON, 0644)

	// Process each worker role SEQUENTIALLY so they can see each other's results
	for i, wr := range req.WorkerRoles {
		role := wr.Role
		description := wr.Description

		if progressCallback != nil {
			progress := 10 + (i * 80 / len(req.WorkerRoles))
			progressCallback(progress, fmt.Sprintf("Starting worker %d/%d: %s", i+1, len(req.WorkerRoles), role), nil)
		}

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

		// Create worker branch if repo path is provided
		if req.ProjectPath != "" {
			branchName := fmt.Sprintf("worker-%s", role)
			if err := git.CreateBranch(req.ProjectPath, branchName); err != nil {
				log.Printf("Failed to create worker branch %s: %v", branchName, err)
			} else {
				log.Printf("Created worker branch: %s", branchName)
			}
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
		var commands []string
		if workerMode == "multypass" {
			files, commands, err = s.generateCodeMultiPass(ctx, provider, model, tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext)
		} else {
			// Fallback to legacy N+1 approach
			files, commands, err = s.generateCode(ctx, provider, model, tokens, taskMD, role, description, req.ManagerRole, basePath, accumulatedContext)
		}

		if err != nil {
			log.Printf("AI error (generate): %v", err)
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		// Execute commands in project path
		projectPath := basePath
		if err := os.MkdirAll(projectPath, 0755); err != nil {
			log.Printf("Warning: failed to ensure project path %s: %v", projectPath, err)
		}
		for _, cmd := range commands {
			if cmd == "" {
				continue
			}
			parts := strings.Fields(cmd)
			if len(parts) == 0 {
				continue
			}
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Dir = projectPath
			cmd.Run() // ignore errors for now
		}

		// Write generated files to disk
		for path, content := range files {
			fullPath := filepath.Join(projectPath, path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				log.Printf("Warning: failed to create dir for %s: %v", path, err)
				continue
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				log.Printf("Warning: failed to write file %s: %v", path, err)
			}
		}
		log.Printf("Worker wrote %d files to %s", len(files), projectPath)

		// Git operations - commit to current branch
		isRepo, err := git.IsGitRepo(projectPath)
		if err != nil {
			worker.Status = "error"
			database.Db.Save(worker)
			return nil, fmt.Errorf("failed to validate git repository at %s: %w", projectPath, err)
		}
		if !isRepo {
			worker.Status = "error"
			database.Db.Save(worker)
			return nil, fmt.Errorf("git repository not found at project path: %s", projectPath)
		}

		commitMessage := fmt.Sprintf("Worker %s: %s", role, taskMD)
		if err := git.Add(projectPath); err != nil {
			worker.Status = "error"
			database.Db.Save(worker)
			return nil, fmt.Errorf("git add failed at %s: %w", projectPath, err)
		}
		if err := git.Commit(projectPath, commitMessage); err != nil {
			worker.Status = "error"
			database.Db.Save(worker)
			return nil, fmt.Errorf("git commit failed at %s: %w", projectPath, err)
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
			Commands:   commands,
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

		if progressCallback != nil {
			progress := 20 + ((i + 1) * 80 / len(req.WorkerRoles))
			progressCallback(progress, fmt.Sprintf("Worker %d/%d (%s) completed: %d files", i+1, len(req.WorkerRoles), role, len(files)), files)
		}
	}

	return &workerpb.AssignWorkersResponse{
		TaskId:        req.TaskId,
		Status:        "success",
		WorkerResults: workerResults,
	}, nil
}
