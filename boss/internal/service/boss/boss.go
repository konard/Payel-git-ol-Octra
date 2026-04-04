package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/fetcher/grpc/manager"
	"boss/internal/fetcher/grpc/merger"
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

	return &BossService{
		managerClient: mgrClient,
		mergerClient:  mrgClient,
	}
}

// BossDecisionResult — result of boss thinking
type BossDecisionResult struct {
	ManagersCount        int32                `json:"managers_count"`
	ManagerRoles         []models.ManagerRole `json:"manager_roles"`
	TechnicalDescription string               `json:"technical_description"`
	TechStack            []string             `json:"tech_stack"`
	ArchitectureNotes    string               `json:"architecture_notes"`
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
