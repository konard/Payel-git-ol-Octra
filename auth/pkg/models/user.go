package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRegister struct {
	gorm.Model
	Id       uuid.UUID `json:"id" gorm:"primaryKey"`
	Username string    `json:"username" gorm:"unique"`
	Email    string    `json:"email" gorm:"unique"`
	Password string    `json:"password" gorm:"not null"`
}
