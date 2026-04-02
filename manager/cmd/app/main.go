package main

import (
	"log"

	"manager/internal/fetcher/grpc"
	"manager/internal/service"
	"manager/pkg/database"
)

func main() {
	// Инициализируем БД
	database.InitDb()

	// Создаём сервис менеджеров
	managerService := service.NewManagerService()

	// Запускаем gRPC сервер на порту 50052
	log.Println("Запуск manager gRPC сервера на порту 50052...")
	if err := grpc.Start("50052", managerService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
