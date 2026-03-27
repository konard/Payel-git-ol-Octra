package create

import (
	"context"
	"fmt"
	"manager/pkg/database"
	"manager/pkg/database/models"

	"github.com/google/uuid"
)

type TaskManager interface {
	CreateManager(ctx context.Context, taskID uuid.UUID, role, agentID string) (*models.Manager, error)
	GetManagersByTask(ctx context.Context, taskID uuid.UUID) ([]models.Manager, error)
	UpdateManagerStatus(ctx context.Context, managerID uuid.UUID, status string) error
}

type TaskManagerImpl struct{}

func NewTaskManager() TaskManager {
	return &TaskManagerImpl{}
}

// CreateManager создаёт менеджера в БД
func (t *TaskManagerImpl) CreateManager(ctx context.Context, taskID uuid.UUID, role, agentID string) (*models.Manager, error) {
	manager := &models.Manager{
		ID:      uuid.New(),
		TaskID:  taskID,
		Role:    role,
		AgentID: agentID,
		Status:  "active",
	}

	if err := database.Db.Create(manager).Error; err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	return manager, nil
}

// GetManagersByTask получает всех менеджеров задачи
func (t *TaskManagerImpl) GetManagersByTask(ctx context.Context, taskID uuid.UUID) ([]models.Manager, error) {
	var managers []models.Manager
	if err := database.Db.Where("task_id = ?", taskID).Find(&managers).Error; err != nil {
		return nil, fmt.Errorf("failed to get managers: %w", err)
	}
	return managers, nil
}

// UpdateManagerStatus обновляет статус менеджера
func (t *TaskManagerImpl) UpdateManagerStatus(ctx context.Context, managerID uuid.UUID, status string) error {
	if err := database.Db.Model(&models.Manager{}).Where("id = ?", managerID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update manager status: %w", err)
	}
	return nil
}
