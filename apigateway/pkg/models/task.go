package models

import (
	"time"

	"gorm.io/gorm"
)

// Task — задача от пользователя
type Task struct {
	gorm.Model
	UserID                string            `gorm:"not null" json:"user_id"`
	Username              string            `gorm:"not null" json:"username"`
	Title                 string            `json:"title"`
	Description           string            `json:"description"`
	TechnicalDescription  string            `json:"technical_description"`
	Tokens                map[string]string `gorm:"type:jsonb" json:"tokens"`
	Meta                  map[string]string `gorm:"type:jsonb" json:"meta"`
	Managers              []Manager         `gorm:"foreignKey:TaskID;references:TaskID" json:"managers"`
	AwaitingClarification bool              `gorm:"default:false" json:"awaiting_clarification"`
	ClarificationQuestion string            `json:"clarification_question"`
	ClarificationResponse string            `json:"clarification_response"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

// Manager — менеджер назначенный на задачу
type Manager struct {
	gorm.Model
	TaskID  string `gorm:"not null;index" json:"task_id"`
	Role    string `gorm:"not null" json:"role"`
	AgentID string `json:"agent_id"`
	Status  string `gorm:"default:'active'" json:"status"`
}
