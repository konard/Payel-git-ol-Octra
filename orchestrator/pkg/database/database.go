package database

import (
	"log"

	"orchestrator/pkg/models"

	"github.com/Payel-git-ol/azure/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Db — глобальное соединение к базе данных
var Db *gorm.DB

// InitDb — открывает соединение и делает авто-миграцию таблиц
func InitDb() {
	dns := env.MustGet("DB_DNS", "")

	var err error
	Db, err = gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	migrations := []interface{}{
		&models.Task{},
		&models.BossDecision{},
		&models.Manager{},
		&models.Worker{},
		&models.WorkerSolution{},
		&models.ContextEntry{},
	}
	for _, model := range migrations {
		if err := Db.AutoMigrate(model); err != nil {
			log.Printf("Failed to auto migrate %T: %v", model, err)
		}
	}

	log.Println("Database initialized")
}

// GetDB — возвращает текущее соединение
func GetDB() *gorm.DB {
	return Db
}

