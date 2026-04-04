package main

import (
	"log"
	"manager/internal/service/manager"

	"manager/internal/fetcher/grpc"
	"manager/pkg/database"
)

func main() {
	database.InitDb()

	managerService := manager.NewManagerService()

	log.Println("Запуск manager gRPC сервера на порту 50052...")
	if err := grpc.Start("50052", managerService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
