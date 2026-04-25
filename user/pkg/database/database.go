package database

import (
	"user/pkg/models"

	"github.com/Payel-git-ol/azure/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDb() {
	dns := env.MustGet("DB_DNS", "")

	var err error
	Db, err = gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Migrate all tables
	Db.AutoMigrate(&models.UserRegister{}, &models.Subscription{}, &models.PromoCode{}, &models.Chat{}, &models.ChatMessage{})
}
