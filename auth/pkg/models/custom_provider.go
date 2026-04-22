package models

import (
	"time"

	"github.com/google/uuid"
)

type CustomProvider struct {
	ID             uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	Name           string    `gorm:"column:name;not null"`
	BaseURL        string    `gorm:"column:base_url;not null"`
	APIKey         string    `gorm:"column:api_key"`
	RequiresApiKey bool      `gorm:"column:requires_api_key;default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `gorm:"index"`
}

type CustomModel struct {
	ID         uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	Name       string     `gorm:"column:name;not null"`
	ProviderID *uuid.UUID `gorm:"column:provider_id;type:uuid"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`
}
