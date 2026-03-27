package main

import (
	"log"
	"manager/internal/fetcher/grpc"
)

func main() {
	// Запускаем gRPC сервер на порту 50052
	if err := grpc.Start("50052"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
