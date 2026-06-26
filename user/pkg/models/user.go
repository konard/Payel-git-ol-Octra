package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRegister struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time `gorm:"index"`
	Username        string     `gorm:"unique"`
	Email           string     `gorm:"unique"`
	Password        string     `gorm:"default:null"`
	BalanceCredits  int64      `gorm:"default:100" json:"balance_credits"`
	SubscriptionEnd *int64     `json:"subscription_end"`
}
