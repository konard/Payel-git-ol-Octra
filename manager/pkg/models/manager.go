package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Manager — менеджер назначенный на задачу
type Manager struct {
	gorm.Model
	ID       uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	TaskID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Role     string    `gorm:"not null"` // frontend, backend, deployment, etc.
	AgentID  string    // ID ИИ агента
	Status   string    `gorm:"default:'active'"` // active, completed, removed
	TaskDesc string    `gorm:"type:text"`        // Техническое описание от boss
}
