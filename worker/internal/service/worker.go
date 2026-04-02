package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"worker/internal/fetcher/grpc/workerpb"
	"worker/pkg/database"
	"worker/pkg/llm"
	"worker/pkg/models"

	"github.com/google/uuid"
)

// WorkerService — worker service
type WorkerService struct {
	workerpb.UnimplementedWorkerServiceServer
}

func NewWorkerService() *WorkerService {
	return &WorkerService{}
}

// AssignWorkersAndWait accepts task, generates code via AI and returns ZIP
func (s *WorkerService) AssignWorkersAndWait(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	log.Printf("Received task from manager %s: %s", req.ManagerId, req.ManagerRole)
	log.Printf("Workers: %v", req.WorkerRoles)

	tokens := parseTokens(req.Metadata["tokens"])
	model := req.Metadata["model"]
	modelURL := req.Metadata["modelUrl"]

	if model == "" {
		model = "openai/gpt-3.5-turbo"
	}
	if modelURL == "" {
		modelURL = "https://openrouter.ai/api/v1"
	}

	llmClient := llm.NewLLMClient(modelURL, model, tokens)

	// Collect all project files
	allFiles := make(map[string]string)

	// Create project directory structure
	projectName := sanitizeProjectName(req.ManagerRole)
	basePath := fmt.Sprintf("project/%s", projectName)

	for _, wr := range req.WorkerRoles {
		role := wr.Role

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

		// Worker thinks about task - creates TASK.md
		taskMD, err := s.createTaskMD(ctx, llmClient, req.TaskMd, role, basePath)
		if err != nil {
			log.Printf("AI error (think): %v", err)
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		// Generate code
		files, err := s.generateCode(ctx, llmClient, taskMD, role, req.ManagerRole, basePath)
		if err != nil {
			log.Printf("AI error (generate): %v", err)
			worker.Status = "error"
			database.Db.Save(worker)
			continue
		}

		// Collect files
		for path, content := range files {
			allFiles[path] = content
		}

		worker.SolutionMD = taskMD
		worker.Files = marshalJSON(files)
		worker.Status = "done"
		worker.Success = true
		database.Db.Save(worker)

		log.Printf("Worker %s completed: %d files", role, len(files))
	}

	// Create ZIP archive
	zipData, err := createZipArchive(allFiles)
	if err != nil {
		log.Printf("Error creating ZIP: %v", err)
		zipData = []byte{}
	}

	log.Printf("Created ZIP archive: %d bytes", len(zipData))

	return &workerpb.AssignWorkersResponse{
		TaskId:   req.TaskId,
		Status:   "success",
		Solution: zipData,
	}, nil
}

func (s *WorkerService) createTaskMD(ctx context.Context, llmClient llm.Client, taskMD, role, basePath string) (string, error) {
	prompt := fmt.Sprintf(`You are a %s developer. Task:

%s

Create TASK.md file with detailed task breakdown for your role. Include:
1. Files to create
2. Functionality to implement
3. Dependencies

Return ONLY the content of TASK.md file.`, role, taskMD)

	resp, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	log.Printf("Worker created TASK.md (%s)", role)
	return resp, nil
}

func (s *WorkerService) generateCode(ctx context.Context, llmClient llm.Client, taskMD, role, managerRole, basePath string) (map[string]string, error) {
	prompt := fmt.Sprintf(`You are a %s developer. Task description:

%s

Generate READY-TO-USE code. Return ONLY JSON:
{
  "%s/%s/main.go": "package main...",
  "%s/%s/README.md": "# Project...",
  "%s/%s/TASK.md": "Task description..."
}

Include TASK.md with task breakdown.`,
		role, taskMD,
		basePath, role,
		basePath, role,
		basePath, role)

	resp, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var files map[string]string
	if err := json.Unmarshal([]byte(resp), &files); err != nil {
		return nil, err
	}

	log.Printf("Generated %d files", len(files))
	return files, nil
}

// Helper functions

func marshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func parseTokens(s string) []string {
	var tokens []string
	json.Unmarshal([]byte(s), &tokens)
	return tokens
}

func sanitizeProjectName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func createZipArchive(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for path, content := range files {
		f, err := w.Create(path)
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
