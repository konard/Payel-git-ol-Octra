package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"manager/internal/fetcher/grpc/managerpb"
	"net"
)

// ManagerHandler — интерфейс для обработки запросов на создание менеджеров
type ManagerHandler interface {
	CreateManagers(ctx context.Context, req *managerpb.CreateManagersRequest) (*managerpb.CreateManagersResponse, error)
	GetTaskStatus(ctx context.Context, req *managerpb.TaskStatusRequest) (*managerpb.TaskStatusResponse, error)
}

type Server struct {
	managerpb.UnimplementedManagerServiceServer
	handler ManagerHandler
}

func NewServer(handler ManagerHandler) *Server {
	return &Server{
		handler: handler,
	}
}

func (s *Server) CreateManagers(ctx context.Context, req *managerpb.CreateManagersRequest) (*managerpb.CreateManagersResponse, error) {
	log.Printf("Получен запрос на создание менеджеров: task=%s, username=%s, quantity=%d", req.TaskId, req.Username, req.QuantityManagers)
	log.Printf("Роли: %v", req.Roles)

	resp, err := s.handler.CreateManagers(ctx, req)
	if err != nil {
		log.Printf("Ошибка создания менеджеров: %v", err)
		return &managerpb.CreateManagersResponse{
			TaskId:  req.TaskId,
			Status:  "error",
			Message: err.Error(),
		}, nil
	}

	return resp, nil
}

func (s *Server) GetTaskStatus(ctx context.Context, req *managerpb.TaskStatusRequest) (*managerpb.TaskStatusResponse, error) {
	log.Printf("Запрос статуса задачи: task=%s", req.TaskId)

	resp, err := s.handler.GetTaskStatus(ctx, req)
	if err != nil {
		log.Printf("Ошибка получения статуса: %v", err)
		return &managerpb.TaskStatusResponse{
			TaskId:       req.TaskId,
			Status:       "error",
			ErrorMessage: err.Error(),
		}, nil
	}

	return resp, nil
}

func Start(port string, handler ManagerHandler) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)

	managerpb.RegisterManagerServiceServer(grpcServer, NewServer(handler))

	log.Printf("gRPC сервер manager запущен на порту %s", port)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
