package database

import (
	"github.com/Payel-git-ol/azure/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"manager/pkg/database/models"
)

var Db *gorm.DB

func InitDb() {
	dns := env.MustGet("DB_DNS", "")

	var err error
	Db, err = gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Авто-миграция всех таблиц
	migrations := []interface{}{
		&models.Manager{},
	}

	for _, model := range migrations {
		if err := Db.AutoMigrate(model); err != nil {
			log.Printf("Failed to auto migrate %T: %v", model, err)
		}
	}

	log.Println("✅ Manager database initialized")
}
