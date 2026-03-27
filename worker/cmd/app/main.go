package main

import (
	"log"

	"worker/internal/fetcher/grpc"
	"worker/internal/service"
	"worker/pkg/database"
)

func main() {
	// Инициализируем БД
	database.InitDb()

	// Создаём сервис воркеров
	workerService := service.NewWorkerService()

	// Запускаем gRPC сервер на порту 50053
	log.Println("Запуск worker gRPC сервера на порту 50053...")
	if err := grpc.Start("50053", workerService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
