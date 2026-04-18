package database

import (
	"boss/pkg/models"
	"log"

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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Авто-миграция всех таблиц
	migrations := []interface{}{
		&models.Task{},
		&models.BossDecision{},
		&models.Manager{},
		&models.Worker{},
		&models.WorkerSolution{},
	}

	for _, model := range migrations {
		if err := Db.AutoMigrate(model); err != nil {
			log.Printf("Failed to auto migrate %T: %v", model, err)
		}
	}

	log.Println("✅ Database initialized")
}

func GetDB() *gorm.DB {
	return Db
}
