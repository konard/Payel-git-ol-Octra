package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task — задача от пользователя
type Task struct {
	ID                    uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID                string    `gorm:"not null;size:36" json:"user_id"`
	Username              string    `gorm:"not null;size:100" json:"username"`
	Status                string    `gorm:"default:'pending';size:50" json:"status"`
	Title                 string    `gorm:"size:500" json:"title"`
	Description           string    `gorm:"type:text" json:"description"`
	TechnicalDescription  string    `gorm:"type:text" json:"technical_description"`
	Tokens                string    `gorm:"type:jsonb" json:"tokens"`
	Meta                  string    `gorm:"type:jsonb" json:"meta"`
	Managers              []Manager `gorm:"foreignKey:TaskID;references:ID" json:"managers"`
	AwaitingClarification bool      `gorm:"default:false" json:"awaiting_clarification"`
	ClarificationQuestion string    `gorm:"size:2000" json:"clarification_question"`
	ClarificationResponse string    `gorm:"size:5000" json:"clarification_response"`
	ProjectJSON           string    `gorm:"type:text" json:"project_json"`
	GitData               string    `gorm:"type:text" json:"git_data"`
	Solution              []byte    `gorm:"type:bytea" json:"solution"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// BossDecision — решение босса по архитектуре задачи
type BossDecision struct {
	gorm.Model
	ID                   uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID               uuid.UUID `gorm:"type:uuid;not null;index"`
	Task                 Task      `gorm:"foreignKey:TaskID"`
	Status               string    `gorm:"default:'planning'"`
	TaskType             string    `gorm:"size:32;default:'code'"`
	ManagersCount        int32
	ManagerRoles         string `gorm:"type:jsonb"`
	TechnicalDescription string `gorm:"type:text"`
	TechStack            string `gorm:"type:jsonb"`
	ArchitectureNotes    string `gorm:"type:text"`
}

// ManagerRole — роль менеджера
type ManagerRole struct {
	Role         string `json:"role"`
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
	Role           string    `gorm:"not null"`
	AgentID        string
	Status         string `gorm:"default:'active'"`
	TaskDesc       string `gorm:"type:text"`
	CustomPrompt   string `gorm:"type:text"`
	WorkerRoles    string
	WorkersCount   int32
}

// WorkerRole — роль воркера
type WorkerRole struct {
	Role         string `json:"role"`
	Description  string `json:"description"`
	CustomPrompt string `json:"customPrompt,omitempty"`
}

// Worker — рабочий назначенный менеджером
type Worker struct {
	gorm.Model
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ManagerID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Manager    Manager   `gorm:"foreignKey:ManagerID"`
	Role       string    `gorm:"not null"`
	AgentID    string
	Status     string `gorm:"default:'thinking'"`
	TaskMD     string `gorm:"type:text"`
	SolutionMD string `gorm:"type:text"`
	Files      string `gorm:"type:text;default:'{}'"`
	Success    bool   `gorm:"default:false"`
	Feedback   string `gorm:"type:text"`
	Approved   bool   `gorm:"default:false"`
}

// WorkerSolution — файл решения воркера
type WorkerSolution struct {
	gorm.Model
	ID       uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Worker   Worker    `gorm:"foreignKey:WorkerID"`
	TaskID   uuid.UUID `gorm:"type:uuid;not null;index"`
	FilePath string    `gorm:"not null"`
	Content  string    `gorm:"type:text"`
	Language string
	Files    string
	Success  bool
	Feedback string `gorm:"type:text"`
	Approved bool   `gorm:"default:false"`
}
