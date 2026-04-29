package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager"
	"boss/internal/redis"
	"boss/internal/service/git"
	"boss/internal/service/github"
	"boss/pkg/database"
	"boss/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

// BossService — boss service
type BossService struct {
	bosspb.UnimplementedBossServiceServer
	managerClient *manager.Client
	redisClient   *redis.Client
	db            *gorm.DB
	githubClient  *github.Client
}

func NewBossService() *BossService {
	mgrClient, err := manager.NewClient(os.Getenv("MANAGER_SERVICE_HOST"))
	if err != nil {
		log.Printf("Warning: failed to connect to manager service: %v", err)
	}

	redisClient := redis.NewClient()
	db := database.GetDB()

	// Initialize GitHub client if token is available
	var githubClient *github.Client
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		githubClient = github.NewClient(
			githubToken,
			os.Getenv("GIT_USER_NAME"),
			os.Getenv("GIT_USER_EMAIL"),
		)
		log.Printf("GitHub client initialized")
	}

	return &BossService{
		managerClient: mgrClient,
		redisClient:   redisClient,
		db:            db,
		githubClient:  githubClient,
	}
}

func (s *BossService) getGitHubClient() *github.Client {
	return s.githubClient
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
	if err := s.db.First(&task, "id = ?", req.TaskId).Error; err != nil {
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
	if err := s.db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		return nil, err
	}

	// Update task status to cancelled
	task.Status = "cancelled"
	if err := s.db.Save(&task).Error; err != nil {
		return nil, err
	}

	// TODO: Implement actual cancellation logic (context cancellation, goroutine cleanup)

	return &bosspb.TaskStatusResponse{
		TaskId:   task.ID.String(),
		Status:   "cancelled",
		Progress: "0%",
	}, nil
}

// restoreProject restores project from DB or Redis JSON
func (s *BossService) restoreProject(taskID string) (string, error) {
	var data string
	var task models.Task

	// First, try DB for faster lookup (one-to-one relation)
	if err := s.db.First(&task, "id = ?", taskID).Error; err == nil && task.ProjectJSON != "" {
		data = task.ProjectJSON
	} else {
		// Fallback to Redis
		key := fmt.Sprintf("project:%s", taskID)
		var err error
		data, err = s.redisClient.GetRedisClient().Get(context.Background(), key).Result()
		if err != nil {
			return "", fmt.Errorf("failed to get project from DB/Redis: %w", err)
		}
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

	// Determine projects directory
	projectsDir := os.Getenv("PROJECTS_DIR")
	if projectsDir == "" {
		projectsDir = "/workspace/projects"
	}

	// Create project in projects directory
	projectPath := filepath.Join(projectsDir, taskID)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create project dir: %w", err)
	}

	// Log the project path for debugging
	log.Printf("📁 Restoring project at: %s", projectPath)

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

	// Restore Git data if available
	if err := s.db.First(&task, "id = ?", taskID).Error; err == nil && task.GitData != "" {
		log.Printf("📦 Restoring Git data...")
		if err := git.RestoreGitDataFromJSON(projectPath, task.GitData); err != nil {
			log.Printf("Warning: failed to restore git data: %v", err)
		} else {
			// Make initial commit after restore
			if err := git.Add(projectPath); err != nil {
				log.Printf("Warning: failed to git add: %v", err)
			} else if err := git.Commit(projectPath, "Restored from history"); err != nil {
				log.Printf("Warning: failed to git commit: %v", err)
			} else {
				log.Printf("✅ Git data restored and committed")
			}
		}
	}

	// Save project JSON to DB for one-to-one relation and faster lookup
	var dbTask models.Task
	if err := database.Db.First(&dbTask, "id = ?", taskID).Error; err == nil {
		dbTask.ProjectJSON = data
		database.Db.Save(&dbTask)
	}

	return projectPath, nil
}
