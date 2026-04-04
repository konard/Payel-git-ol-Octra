package main

import (
	"boss/internal/service/boss"
	"log"

	"boss/internal/fetcher/grpc"
	"boss/pkg/database"
)

func main() {
	database.InitDb()

	bossService := boss.NewBossService()

	log.Println("Запуск boss gRPC сервера на порту 50051...")
	if err := grpc.Start("50051", bossService); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
