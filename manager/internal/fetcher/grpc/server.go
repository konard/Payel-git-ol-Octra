package grpc

import (
	"context"
	"log"
	"net"

	"manager/internal/fetcher/grpc/managerpb"

	"google.golang.org/grpc"
)

// Server — gRPC сервер manager сервиса
type Server struct {
	managerpb.UnimplementedManagerServiceServer
}

// CreateManagers принимает запрос на создание менеджеров
func (s *Server) CreateManagers(ctx context.Context, req *managerpb.CreateManagersRequest) (*managerpb.CreateManagersResponse, error) {
	log.Printf("Получена задача: %s, менеджеров: %d, роли: %v", req.TaskId, req.QuantityManagers, req.Roles)

	managers := make([]*managerpb.ManagerInfo, 0, len(req.Roles))
	for i, role := range req.Roles {
		managers = append(managers, &managerpb.ManagerInfo{
			Id:      "manager-" + string(rune(i)),
			Role:    role,
			AgentId: "agent-" + string(rune(i)),
			Status:  "active",
		})
	}

	return &managerpb.CreateManagersResponse{
		TaskId:          req.TaskId,
		Status:          "success",
		Message:         "Managers created",
		CreatedManagers: int32(len(managers)),
		Managers:        managers,
	}, nil
}

// GetTaskStatus возвращает статус задачи
func (s *Server) GetTaskStatus(ctx context.Context, req *managerpb.TaskStatusRequest) (*managerpb.TaskStatusResponse, error) {
	return &managerpb.TaskStatusResponse{
		TaskId:        req.TaskId,
		Status:        "processing",
		ManagersCount: 0,
		WorkersCount:  0,
	}, nil
}

// Start запускает gRPC сервер
func Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	managerpb.RegisterManagerServiceServer(server, &Server{})

	log.Printf("Manager gRPC сервер запущен на порту %s", port)

	return server.Serve(lis)
}
