package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager"
	"boss/internal/fetcher/grpc/merger"
	"boss/internal/redis"
	"boss/pkg/database"
	"boss/pkg/models"
	"context"
	"log"
	"os"
)

// BossService — boss service
type BossService struct {
	bosspb.UnimplementedBossServiceServer
	managerClient *manager.Client
	mergerClient  *merger.Client
	redisClient   *redis.Client
}

func NewBossService() *BossService {
	mgrClient, err := manager.NewClient(os.Getenv("MANAGER_SERVICE_HOST"))
	if err != nil {
		log.Printf("Warning: failed to connect to manager service: %v", err)
	}

	mrgClient, err := merger.NewClient(os.Getenv("MERGER_SERVICE_HOST"))
	if err != nil {
		log.Printf("Warning: failed to connect to merger service: %v", err)
	}

	redisClient := redis.NewClient()

	return &BossService{
		managerClient: mgrClient,
		mergerClient:  mrgClient,
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
