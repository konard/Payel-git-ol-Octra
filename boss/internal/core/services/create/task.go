package create

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/pkg/database"
	"boss/pkg/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type TaskManager interface {
	CreateNewTask(ctx context.Context, req *bosspb.TaskRequest) (*models.Task, error)
	GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) error
	AddManager(ctx context.Context, taskID uuid.UUID, role, agentID string) error
	AddWorker(ctx context.Context, taskID uuid.UUID, role, agentID string) error
	SetSolution(ctx context.Context, taskID uuid.UUID, solution []byte) error
}

type Tasker struct {
	db *database.Database
}

func NewTasker() *Tasker {
	return &Tasker{}
}

// CreateNewTask создаёт новую задачу в БД
func (t *Tasker) CreateNewTask(ctx context.Context, req *bosspb.TaskRequest) (*models.Task, error) {
	// Сериализуем токены и метаданные в JSON
	tokensJSON, err := json.Marshal(req.Tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tokens: %w", err)
	}

	metaJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	task := &models.Task{
		ID:          uuid.New(),
		Username:    req.Username,
		TaskName:    req.Taskname,
		Title:       req.Title,
		Description: req.Description,
		Tokens:      string(tokensJSON),
		Meta:        string(metaJSON),
		Status:      "pending",
	}

	if err := database.Db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetTask получает задачу по ID
func (t *Tasker) GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := database.Db.Preload("Managers").Preload("Workers").First(&task, "id = ?", taskID).Error; err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// UpdateTaskStatus обновляет статус задачи
func (t *Tasker) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) error {
	if err := database.Db.Model(&models.Task{}).Where("id = ?", taskID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// AddManager добавляет менеджера к задаче
func (t *Tasker) AddManager(ctx context.Context, taskID uuid.UUID, role, agentID string) error {
	manager := &models.Manager{
		ID:      uuid.New(),
		TaskID:  taskID,
		Role:    role,
		AgentID: agentID,
		Status:  "active",
	}

	if err := database.Db.Create(manager).Error; err != nil {
		return fmt.Errorf("failed to add manager: %w", err)
	}
	return nil
}

// AddWorker добавляет рабочего к задаче
func (t *Tasker) AddWorker(ctx context.Context, taskID uuid.UUID, role, agentID string) error {
	worker := &models.Worker{
		ID:      uuid.New(),
		TaskID:  taskID,
		Role:    role,
		AgentID: agentID,
		Status:  "active",
	}

	if err := database.Db.Create(worker).Error; err != nil {
		return fmt.Errorf("failed to add worker: %w", err)
	}
	return nil
}

// SetSolution устанавливает решение (ZIP архив) для задачи
func (t *Tasker) SetSolution(ctx context.Context, taskID uuid.UUID, solution []byte) error {
	if err := database.Db.Model(&models.Task{}).Where("id = ?", taskID).Update("solution", solution).Error; err != nil {
		return fmt.Errorf("failed to set solution: %w", err)
	}
	return nil
}
