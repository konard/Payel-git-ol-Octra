package main

import (
	"boss/internal/core/services/create"
	"boss/internal/fetcher/grpc"
	"boss/internal/fetcher/grpc/bosspb"
	"boss/pkg/database"
	"context"
	"log"
)

// ProjectGenerator — генератор проектов с сохранением в БД
type ProjectGenerator struct {
	taskManager create.TaskManager
}

// NewProjectGenerator создаёт генератор с TaskManager
func NewProjectGenerator() *ProjectGenerator {
	return &ProjectGenerator{
		taskManager: create.NewTasker(),
	}
}

// GenerateProject генерирует проект и сохраняет в БД
func (g *ProjectGenerator) GenerateProject(ctx context.Context, req *bosspb.TaskRequest) ([]byte, error) {
	log.Printf("Получена задача: user=%s, task=%s, title=%s", req.Username, req.Taskname, req.Title)

	// Создаём задачу в БД
	task, err := g.taskManager.CreateNewTask(ctx, req)
	if err != nil {
		log.Printf("Ошибка создания задачи: %v", err)
		return nil, err
	}

	log.Printf("Задача создана в БД: ID=%s", task.ID)

	// Обновляем статус на "processing"
	if err := g.taskManager.UpdateTaskStatus(ctx, task.ID, "processing"); err != nil {
		log.Printf("Ошибка обновления статуса: %v", err)
	}

	// TODO: Тут будет логика генерации проекта через YAML файлы
	// Пока возвращаем тестовые данные
	testData := []byte("Project generated for: " + req.Title)

	// Сохраняем решение в БД
	if err := g.taskManager.SetSolution(ctx, task.ID, testData); err != nil {
		log.Printf("Ошибка сохранения решения: %v", err)
	}

	// Обновляем статус на "done"
	if err := g.taskManager.UpdateTaskStatus(ctx, task.ID, "done"); err != nil {
		log.Printf("Ошибка обновления статуса: %v", err)
	}

	log.Printf("Задача выполнена: размер=%d байт", len(testData))

	return testData, nil
}

func main() {
	// Инициализируем БД
	database.InitDb()

	// Создаём генератор проектов
	generator := NewProjectGenerator()

	// Запускаем gRPC сервер на порту 50051
	log.Println("Запуск gRPC сервера на порту 50051...")
	if err := grpc.Start("50051", generator); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
