package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task — задача от пользователя
type Task struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      string    `gorm:"not null;index"`
	Username    string
	Title       string `gorm:"not null"`
	Description string `gorm:"type:text"`
	Tokens      string `gorm:"type:jsonb"` // []string
	Meta        string `gorm:"type:jsonb"` // map[string]string

	// Статусы
	Status string `gorm:"default:'pending'"` // pending, boss_planning, managers_assigned, workers_assigned, processing, reviewing, done, error

	// Результаты
	Solution      []byte `gorm:"type:bytea"` // ZIP архив
	ResultArchive string // Путь к архиву /tasks/{task_id}/solution.zip
}

// BossDecision — решение босса
type BossDecision struct {
	gorm.Model
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Task   Task      `gorm:"foreignKey:TaskID"`
	Status string    `gorm:"default:'planning'"`

	// Босс расписывает архитектуру
	ManagersCount        int32
	ManagerRoles         string `gorm:"type:jsonb"` // []ManagerRole
	TechnicalDescription string `gorm:"type:text"`
	TechStack            string `gorm:"type:jsonb"` // []string
	ArchitectureNotes    string `gorm:"type:text"`
}

// ManagerRole — роль менеджера
type ManagerRole struct {
	Role        string `json:"role"` // "frontend", "backend", "testing"
	Description string `json:"description"`
	Priority    int32  `json:"priority"`
}

// Manager — менеджер назначенный на задачу
type Manager struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Task           Task      `gorm:"foreignKey:TaskID"`
	BossDecisionID uuid.UUID `gorm:"type:uuid;index"`

	Role    string `gorm:"not null"` // "frontend", "backend", "testing"
	AgentID string // ID ИИ агента
	Status  string `gorm:"default:'active'"` // active, completed, error

	// Решение менеджера
	WorkerRoles  string `gorm:"type:jsonb"` // []WorkerRole
	WorkersCount int32
}

// WorkerRole — роль воркера
type WorkerRole struct {
	Role        string `json:"role"` // "react-dev", "go-dev", "qa-automation"
	Description string `json:"description"`
}

// Worker — рабочий назначенный на задачу
type Worker struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index"`
	ManagerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Manager   Manager   `gorm:"foreignKey:ManagerID"`

	Role       string `gorm:"not null"`
	AgentID    string
	Status     string `gorm:"default:'thinking'"` // thinking, coding, testing, done, error
	TaskMD     string `gorm:"type:text"`          // TASK.md
	SolutionMD string `gorm:"type:text"`          // SOLUTION.md
}

// WorkerSolution — результат работы воркера
type WorkerSolution struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Worker   Worker    `gorm:"foreignKey:WorkerID"`
	TaskID   uuid.UUID `gorm:"type:uuid;not null;index"`

	// Файлы решения
	Files    string `gorm:"type:jsonb"` // map[string]string {path: content}
	Success  bool
	Feedback string `gorm:"type:text"` // Если нужно исправить
	Approved bool   `gorm:"default:false"`
}
