package boss

import (
	"context"
	"fmt"
	"os"

	"nodes/internal/redis"
	"nodes/internal/service/agents"
	"nodes/internal/service/github"
	"nodes/internal/service/rules"
	"nodes/internal/service/rules/manager"
	"nodes/pkg/database"
	"nodes/pkg/models"

	"gorm.io/gorm"
)

// Service — узел "boss". Раньше был самостоятельным gRPC-сервисом,
// теперь это просто компонент внутри nodes, который напрямую вызывает
// manager.Service вместо gRPC-клиента.
type Service struct {
	agentsClient *agents.Client
	manager      *manager.Service
	redisClient  *redis.Client
	githubClient *github.Client
	db           *gorm.DB
}

// NewService — создаёт босса, привязанного к локальному менеджеру
func NewService(agentsClient *agents.Client, mgr *manager.Service, redisClient *redis.Client) *Service {
	var gh *github.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		gh = github.NewClient(token, os.Getenv("GIT_USER_NAME"), os.Getenv("GIT_USER_EMAIL"))
	}
	return &Service{
		agentsClient: agentsClient,
		manager:      mgr,
		redisClient:  redisClient,
		githubClient: gh,
		db:           database.Db,
	}
}

// GetTaskStatus — простой статус задачи (внешний gRPC-метод)
func (s *Service) GetTaskStatus(ctx context.Context, taskID string) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return &task, nil
}

// StopTask — помечает задачу как cancelled
func (s *Service) StopTask(ctx context.Context, taskID string) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	task.Status = "cancelled"
	if err := s.db.Save(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTaskRequest — структура входной задачи (раньше bosspb.CreateTaskRequest)
type CreateTaskRequest struct {
	UserId        string
	Username      string
	Title         string
	Description   string
	Tokens        map[string]string
	Meta          map[string]string
	UseAiPlanning bool
	Grade         int // model-predicted complexity 1-10

	// Для доработок существующего проекта
	ExistingRepoUrl string // если есть — продолжаем работу над этим репозиторием
	IsRefinement    bool   // если true — не восстанавливаем проект, а продолжаем
}

// DecisionResult — итог boss-планирования
type DecisionResult struct {
	// TaskType — вид результата: "code" | "research" | "document" | "presentation".
	// json-тег нужен, потому что AI возвращает snake_case ключ.
	TaskType             string `json:"task_type"`
	ManagersCount        int32
	ManagerRoles         []models.ManagerRole
	TechnicalDescription string
	TechStack            []string
	ArchitectureNotes    string
}

// progressBridge — переходник между rules.ProgressFunc и progressCallback менеджеров
func (s *Service) bridge(progress rules.ProgressFunc) func(role string, p int, msg string) {
	if progress == nil {
		return func(string, int, string) {}
	}
	return func(role string, p int, msg string) {
		progress(int32(p), msg, map[string]string{"current_role": role})
	}
}
