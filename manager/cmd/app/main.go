package main

import (
	"context"
	"fmt"
	"log"
	"manager/internal/core/services/create"
	"manager/internal/fetcher/grpc"
	"manager/internal/fetcher/grpc/manager/managerspb"
	"manager/pkg/database"
)

// ManagerHandler — реализация обработчика запросов
type ManagerHandler struct {
	taskManager create.TaskManager
}

// NewManagerHandler создаёт обработчик
func NewManagerHandler() *ManagerHandler {
	return &ManagerHandler{
		taskManager: create.NewTaskManager(),
	}
}

// CreateManagers создаёт команду менеджеров для задачи
func (h *ManagerHandler) CreateManagers(ctx context.Context, req *managerspb.CreateManagersRequest) (*managerspb.CreateManagersResponse, error) {
	log.Printf("Создание менеджеров для задачи %s", req.TaskId)

	// TODO: Тут будет логика создания менеджеров через ИИ агентов
	// Пока создаём заглушки

	managers := make([]*managerspb.ManagerInfo, 0, len(req.Roles))
	for i, role := range req.Roles {
		manager := &managerspb.ManagerInfo{
			Id:      fmt.Sprintf("manager-%d", i),
			Role:    role,
			AgentId: fmt.Sprintf("agent-%d", i),
			Status:  "active",
		}
		managers = append(managers, manager)
	}

	return &managerspb.CreateManagersResponse{
		TaskId:          req.TaskId,
		Status:          "success",
		Message:         "Managers created successfully",
		CreatedManagers: int32(len(managers)),
		Managers:        managers,
	}, nil
}

// GetTaskStatus возвращает статус задачи
func (h *ManagerHandler) GetTaskStatus(ctx context.Context, req *managerspb.TaskStatusRequest) (*managerspb.TaskStatusResponse, error) {
	// TODO: Получить статус из БД
	return &managerspb.TaskStatusResponse{
		TaskId:        req.TaskId,
		Status:        "processing",
		ManagersCount: 0,
		WorkersCount:  0,
	}, nil
}

func main() {
	// Инициализируем БД
	database.InitDb()

	// Создаём обработчик
	handler := NewManagerHandler()

	// Запускаем gRPC сервер на порту 50052
	log.Println("Запуск gRPC сервера manager на порту 50052...")
	if err := grpc.Start("50052", handler); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
