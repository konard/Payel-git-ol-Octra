package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	Id           uuid.UUID        `gorm:"type:uuid;primary_key" json:"id"`
	UserId       uuid.UUID      `gorm:"type:uuid;index" json:"user_id"`
	Title        string       `gorm:"type:varchar(255)" json:"title"`
	LastMessage  string       `gorm:"type:text" json:"last_message"`
	Workflow    string       `gorm:"type:text" json:"workflow,omitempty"`
	Provider    string       `gorm:"type:varchar(100)" json:"provider,omitempty"`
	Model       string       `gorm:"type:varchar(100)" json:"model,omitempty"`
	Messages    []ChatMessage `gorm:"foreignKey:ChatId" json:"messages,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type ChatMessage struct {
	Id         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ChatId     uuid.UUID `gorm:"type:uuid;index" json:"chat_id"`
	Role       string  `gorm:"type:varchar(20);not null" json:"role"` // user or boss
	Content    string  `gorm:"type:text;not null" json:"content"`
	Provider   string  `gorm:"type:varchar(100)" json:"provider,omitempty"`
	Model      string  `gorm:"type:varchar(100)" json:"model,omitempty"`
	ApiKeyId   string  `gorm:"type:varchar(100)" json:"api_key_id,omitempty"` // ID only, not the key itself
	CreatedAt  time.Time `json:"created_at"`
}