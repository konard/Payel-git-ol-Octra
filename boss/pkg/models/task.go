package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task — задача на создание проекта
type Task struct {
	gorm.Model
	ID          uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Username    string    `gorm:"not null"`
	TaskName    string    `gorm:"not null"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	Tokens      string    `gorm:"type:jsonb"`
	Meta        string    `gorm:"type:jsonb"`
	Status      string    `gorm:"default:'pending'"`

	Managers []Manager `gorm:"foreignKey:TaskID"`
	Workers  []Worker  `gorm:"foreignKey:TaskID"`

	Solution []byte `gorm:"type:bytea"`
}

type Manager struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	TaskID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Role    string    `gorm:"not null"`
	AgentID string
	Status  string `gorm:"default:'active'"`
}

type Worker struct {
	gorm.Model
	ID      uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	TaskID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Role    string    `gorm:"not null"`
	AgentID string
	Status  string `gorm:"default:'active'"`
}
