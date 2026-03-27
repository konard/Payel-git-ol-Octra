package main

import (
	"log"

	"boss/internal/fetcher/grpc"
	"boss/internal/service"
	"boss/pkg/database"
)

func main() {
	// Инициализируем БД
	database.InitDb()

	// Создаём сервис босса
	bossService := service.NewBossService()

	// Запускаем gRPC сервер на порту 50051
	log.Println("Запуск boss gRPC сервера на порту 50051...")
	if err := grpc.Start("50051", bossService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
