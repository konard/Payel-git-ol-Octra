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
func (s *Server) CreateManagers(ctx context.Context, req *managerpb.AssignManagersRequest) (*managerpb.AssignManagersResponse, error) {
	log.Printf("Получена задача: %s, роли менеджеров: %v", req.TaskId, req.ManagerRoles)

	managers := make([]*managerpb.ManagerInfo, 0, len(req.ManagerRoles))
	for i, mr := range req.ManagerRoles {
		managers = append(managers, &managerpb.ManagerInfo{
			Id:          "manager-" + string(rune(i+1)),
			Role:        mr.Role,
			Status:      "active",
			WorkerRoles: []*managerpb.WorkerRole{},
		})
	}

	return &managerpb.AssignManagersResponse{
		TaskId:   req.TaskId,
		Status:   "success",
		Managers: managers,
	}, nil
}

// GetTaskStatus возвращает статус задачи
func (s *Server) GetTaskStatus(ctx context.Context, req *managerpb.GetTaskRequest) (*managerpb.GetTaskResponse, error) {
	return &managerpb.GetTaskResponse{
		TaskId:               "unknown",
		TechnicalDescription: "Task processing...",
		ManagerRole:          "unknown",
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
