package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Workflow — конфигурация workflow
type Workflow struct {
	UseAIPlanning bool     `json:"use_ai_planning"`
	Architecture  string   `json:"architecture"`
	TechStack     []string `json:"tech_stack"`
}

// Task — задача от пользователя
type Task struct {
	ID                    uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TaskID                string    `gorm:"uniqueIndex;not null" json:"task_id"`
	UserID                string    `gorm:"not null" json:"user_id"`
	Username              string    `gorm:"not null" json:"username"`
	Status                string    `gorm:"default:'pending'" json:"status"`
	Title                 string    `json:"title"`
	Description           string    `json:"description"`
	TechnicalDescription  string    `json:"technical_description"`
	Tokens                string    `gorm:"type:jsonb" json:"tokens"`
	Meta                  string    `gorm:"type:jsonb" json:"meta"`
	Workflow              Workflow  `gorm:"embedded;embeddedPrefix:workflow_" json:"workflow"`
	Managers              []Manager `gorm:"foreignKey:TaskID;references:TaskID" json:"managers"`
	AwaitingClarification bool      `gorm:"default:false" json:"awaiting_clarification"`
	ClarificationQuestion string    `json:"clarification_question"`
	ClarificationResponse string    `json:"clarification_response"`
	ProjectJSON           string    `gorm:"type:text" json:"project_json"`
	Solution              []byte    `gorm:"type:bytea" json:"solution"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// BossDecision — решение босса
type BossDecision struct {
	gorm.Model
	ID     uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
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
	Role         string `json:"role"` // "frontend", "backend", "testing"
	Description  string `json:"description"`
	Priority     int32  `json:"priority"`
	CustomPrompt string `json:"customPrompt,omitempty"`
}

// Manager — менеджер назначенный на задачу
type Manager struct {
	gorm.Model
	ID             uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
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
	Role         string `json:"role"` // "react-dev", "go-dev", "qa-automation"
	Description  string `json:"description"`
	CustomPrompt string `json:"customPrompt,omitempty"`
}

// Worker — рабочий назначенный на задачу
type Worker struct {
	gorm.Model
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
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
	ID       uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Worker   Worker    `gorm:"foreignKey:WorkerID"`
	TaskID   uuid.UUID `gorm:"type:uuid;not null;index"`

	// Файлы решения
	Files    string `gorm:"type:jsonb"` // map[string]string {path: content}
	Success  bool
	Feedback string `gorm:"type:text"` // Если нужно исправить
	Approved bool   `gorm:"default:false"`
}
