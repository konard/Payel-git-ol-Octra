package main

import (
	"log"
	"os"

	"boss/internal/fetcher/grpc"
	"boss/internal/service"
	"boss/pkg/database"
)

func main() {
	// Инициализируем БД
	database.InitDb()

	// Получаем API ключ из env
	apiKey := os.Getenv("AZURE_API_KEY")
	if apiKey == "" {
		log.Fatal("AZURE_API_KEY environment variable is required")
	}

	// Создаём сервис босса
	bossService := service.NewBossService(apiKey)

	// Запускаем gRPC сервер на порту 50051
	log.Println("Запуск boss gRPC сервера на порту 50051...")
	if err := grpc.Start("50051", bossService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
