package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Manager — менеджер назначенный на задачу
type Manager struct {
	gorm.Model
	ID       uuid.UUID `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	TaskID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Role     string    `gorm:"not null"` // frontend, backend, deployment, etc.
	AgentID  string    // ID ИИ агента
	Status   string    `gorm:"default:'active'"` // active, completed, removed
	TaskDesc string    `gorm:"type:text"`        // Техническое описание от boss
	CustomPrompt string `gorm:"type:text"`      // Кастомный промт для менеджера

	// Решение менеджера
	WorkerRoles  string `gorm:"type:text"` // []WorkerRole (JSON)
	WorkersCount int32
}

// WorkerRole — роль воркера
type WorkerRole struct {
	Role        string `json:"role"`
	Description string `json:"description"`
	CustomPrompt string `json:"customPrompt,omitempty"`
}
