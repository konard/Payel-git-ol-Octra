package main

import (
	"log"
	"worker/internal/service/worker"

	"worker/internal/fetcher/grpc"
	"worker/pkg/database"
)

func main() {
	// Инициализируем БД
	database.InitDb()

	// Создаём сервис воркеров
	workerService := worker.NewWorkerService()

	// Запускаем gRPC сервер на порту 50053
	log.Println("Запуск worker gRPC сервера на порту 50053...")
	if err := grpc.Start("50053", workerService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
