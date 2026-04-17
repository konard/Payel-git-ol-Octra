package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager"
	"boss/internal/redis"
	"boss/pkg/database"
	"boss/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// BossService — boss service
type BossService struct {
	bosspb.UnimplementedBossServiceServer
	managerClient *manager.Client
	redisClient   *redis.Client
}

func NewBossService() *BossService {
	mgrClient, err := manager.NewClient(os.Getenv("MANAGER_SERVICE_HOST"))
	if err != nil {
		log.Printf("Warning: failed to connect to manager service: %v", err)
	}

	redisClient := redis.NewClient()

	return &BossService{
		managerClient: mgrClient,
		redisClient:   redisClient,
	}
}

// BossDecisionResult — result of boss thinking
type BossDecisionResult struct {
	ManagersCount        int32                `json:"managers_count"`
	ManagerRoles         []models.ManagerRole `json:"manager_roles"`
	TechnicalDescription string               `json:"technical_description"`
	TechStack            []string             `json:"tech_stack"`
	ArchitectureNotes    string               `json:"architecture_notes"`
	// Predefined worker roles per manager (from user workflow)
	ManagerWorkerRoles map[string][]models.WorkerRole `json:"manager_worker_roles,omitempty"`
}

// GetTaskStatus returns task status
func (s *BossService) GetTaskStatus(ctx context.Context, req *bosspb.TaskStatusRequest) (*bosspb.TaskStatusResponse, error) {
	var task models.Task
	if err := database.Db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		return nil, err
	}

	return &bosspb.TaskStatusResponse{
		TaskId:   task.ID.String(),
		Status:   task.Status,
		Progress: "50%",
	}, nil
}

// StopTask stops a running task
func (s *BossService) StopTask(ctx context.Context, req *bosspb.StopTaskRequest) (*bosspb.TaskStatusResponse, error) {
	var task models.Task
	if err := database.Db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		return nil, err
	}

	// Update task status to cancelled
	task.Status = "cancelled"
	if err := database.Db.Save(&task).Error; err != nil {
		return nil, err
	}

	// TODO: Implement actual cancellation logic (context cancellation, goroutine cleanup)

	return &bosspb.TaskStatusResponse{
		TaskId:   task.ID.String(),
		Status:   "cancelled",
		Progress: "0%",
	}, nil
}

// restoreProject restores project from Redis JSON
func (s *BossService) restoreProject(taskID string) (string, error) {
	key := fmt.Sprintf("project:%s", taskID)
	data, err := s.redisClient.GetRedisClient().Get(context.Background(), key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get project from Redis: %w", err)
	}

	var project struct {
		NameProject string `json:"name_project"`
		Project     struct {
			Dir   map[string]string `json:"dir"`
			Files map[string]string `json:"files"`
		} `json:"project"`
	}
	if err := json.Unmarshal([]byte(data), &project); err != nil {
		return "", fmt.Errorf("failed to unmarshal project JSON: %w", err)
	}

	projectPath := fmt.Sprintf("/tmp/projects/%s", taskID)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create project dir: %w", err)
	}

	for relPath, content := range project.Project.Files {
		fullPath := filepath.Join(projectPath, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			continue
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			continue
		}
	}

	// For dir, create empty files or dirs if needed
	for relPath, _ := range project.Project.Dir {
		fullPath := filepath.Join(projectPath, relPath)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			continue
		}
	}

	return projectPath, nil
}
