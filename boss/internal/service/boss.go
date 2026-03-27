package main

import (
	"context"
	"encoding/json"
	"log"

	"boss/internal/fetcher/grpc/bosspb"
	"boss/pkg/database"
	"boss/pkg/models"
	"github.com/google/uuid"
	"github.com/yhwhpe/llm-unified-client"
)

// BossService — сервис босса
type BossService struct {
	bosspb.UnimplementedBossServiceServer
	llm *llm.Client
}

func NewBossService() *BossService {
	// Инициализация LLM клиента
	llmClient, err := llm.NewClient(llm.Config{
		Provider: "azure",
		// API ключ из env
	})
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	return &BossService{
		llm: llmClient,
	}
}

// CreateTask принимает задачу от пользователя и расписывает план
func (s *BossService) CreateTask(ctx context.Context, req *bosspb.CreateTaskRequest) (*bosspb.BossDecision, error) {
	log.Printf("Получена задача от %s: %s", req.Username, req.Title)

	// 1. Сохраняем задачу в БД
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      req.UserId,
		Username:    req.Username,
		Title:       req.Title,
		Description: req.Description,
		Status:      "boss_planning",
	}

	// Сериализуем токены и метаданные
	tokensJSON, _ := json.Marshal(req.Tokens)
	metaJSON, _ := json.Marshal(req.Meta)
	task.Tokens = string(tokensJSON)
	task.Meta = string(metaJSON)

	if err := database.Db.Create(task).Error; err != nil {
		return nil, err
	}

	// 2. Босс думает над задачей через ИИ
	decision, err := s.thinkAboutTask(req)
	if err != nil {
		return &bosspb.BossDecision{
			TaskId:      task.ID.String(),
			Status:      "error",
			ErrorMessage: err.Error(),
		}, nil
	}

	// 3. Сохраняем решение босса
	bossDecision := &models.BossDecision{
		ID:                   uuid.New(),
		TaskID:               task.ID,
		Status:               "planning",
		ManagersCount:        decision.ManagersCount,
		TechnicalDescription: decision.TechnicalDescription,
		ArchitectureNotes:    decision.ArchitectureNotes,
	}

	// Сериализуем роли менеджеров и стек
	rolesJSON, _ := json.Marshal(decision.ManagerRoles)
	stackJSON, _ := json.Marshal(decision.TechStack)
	bossDecision.ManagerRoles = string(rolesJSON)
	bossDecision.TechStack = string(stackJSON)

	if err := database.Db.Create(bossDecision).Error; err != nil {
		return nil, err
	}

	// 4. Обновляем статус задачи
	task.Status = "managers_assigned"
	database.Db.Save(task)

	// 5. Отправляем задачу менеджер сервису
	// TODO: Вызвать managerClient.AssignManagers()

	return &bosspb.BossDecision{
		TaskId:               task.ID.String(),
		Status:               "managers_assigned",
		ManagersCount:        decision.ManagersCount,
		TechnicalDescription: decision.TechnicalDescription,
		TechStack:            decision.TechStack,
		ArchitectureNotes:    decision.ArchitectureNotes,
	}, nil
}

// BossDecision — результат размышлений босса
type BossDecision struct {
	ManagersCount        int32
	ManagerRoles         []models.ManagerRole
	TechnicalDescription string
	TechStack            []string
	ArchitectureNotes    string
}

// thinkAboutTask — босс думает над задачей через ИИ
func (s *BossService) thinkAboutTask(req *bosspb.CreateTaskRequest) (*BossDecision, error) {
	prompt := `Ты технический директор (CTO) компании. Тебе дали задачу:

Название: ` + req.Title + `
Описание: ` + req.Description + `

Твоя задача:
1. Определить сколько менеджеров нужно для этой задачи
2. Расписать роли менеджеров (frontend, backend, testing, deployment, devops)
3. Описать техническое задание для менеджеров
4. Указать стек технологий
5. Описать архитектуру решения

Ответь в формате JSON:
{
  "managers_count": 3,
  "manager_roles": [
    {"role": "frontend", "description": "Отвечает за frontend часть", "priority": 1},
    {"role": "backend", "description": "Отвечает за backend API", "priority": 2},
    {"role": "testing", "description": "Тестирование и QA", "priority": 3}
  ],
  "technical_description": "Подробное техническое описание...",
  "tech_stack": ["Go", "React", "PostgreSQL", "Docker"],
  "architecture_notes": "Заметки об архитектуре..."
}`

	// Вызов к ИИ
	resp, err := s.llm.Chat(context.Background(), llm.Message{
		Role:    "user",
		Content: prompt,
	})
	if err != nil {
		return nil, err
	}

	// Парсим ответ ИИ
	var decision BossDecision
	if err := json.Unmarshal([]byte(resp.Content), &decision); err != nil {
		return nil, err
	}

	log.Printf("Босс принял решение: %+v", decision)

	return &decision, nil
}

// GetTaskStatus возвращает статус задачи
func (s *BossService) GetTaskStatus(ctx context.Context, req *bosspb.TaskStatusRequest) (*bosspb.TaskStatusResponse, error) {
	var task models.Task
	if err := database.Db.First(&task, "id = ?", req.TaskId).Error; err != nil {
		return nil, err
	}

	return &bosspb.TaskStatusResponse{
		TaskId:  task.ID.String(),
		Status:  task.Status,
		Progress: "50%", // TODO: Считать прогресс
	}, nil
}
