package main

import (
	"log"
	"os"

	"worker/internal/fetcher/grpc"
	"worker/internal/service"
	"worker/pkg/database"
)

func main() {
	// Инициализируем БД
	database.InitDb()

	// Получаем API ключ из env
	apiKey := os.Getenv("AZURE_API_KEY")
	if apiKey == "" {
		log.Fatal("AZURE_API_KEY environment variable is required")
	}

	// Создаём сервис воркеров
	workerService := service.NewWorkerService(apiKey)

	// Запускаем gRPC сервер на порту 50053
	log.Println("Запуск worker gRPC сервера на порту 50053...")
	if err := grpc.Start("50053", workerService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
