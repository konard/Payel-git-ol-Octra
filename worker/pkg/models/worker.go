package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Worker — рабочий назначенный на задачу
type Worker struct {
	gorm.Model
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index"`
	ManagerID uuid.UUID `gorm:"type:uuid;not null;index"`

	Role    string `gorm:"not null"` // "frontend-dev", "backend-dev", "tester", etc.
	AgentID string // ID ИИ агента
	Status  string `gorm:"default:'thinking'"` // thinking, coding, testing, done, error

	// Задача от менеджера
	TaskMD     string `gorm:"type:text"` // TASK.md - что нужно сделать
	SolutionMD string `gorm:"type:text"` // SOLUTION.md - как решил

	// Результат
	Files    string `gorm:"type:text;default:'{}'"` // map[string]string {path: content}
	Success  bool   `gorm:"default:false"`
	Feedback string `gorm:"type:text"` // Если нужно исправить
	Approved bool   `gorm:"default:false"`
}

// WorkerSolution — файл решения
type WorkerSolution struct {
	gorm.Model
	ID       uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Worker   Worker    `gorm:"foreignKey:WorkerID"`

	FilePath string `gorm:"not null"`
	Content  string `gorm:"type:text"`
	Language string // "go", "typescript", "python", etc.
}
