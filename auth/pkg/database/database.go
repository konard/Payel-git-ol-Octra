package database

import (
	"auth/pkg/models"
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

	Db.AutoMigrate(&models.UserRegister{})
}
