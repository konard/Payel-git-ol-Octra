package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"boss/internal/fetcher/grpc/bosspb"

	"google.golang.org/grpc"
)

type TaskHandler interface {
	GenerateProject(ctx context.Context, req *bosspb.TaskRequest) ([]byte, error)
}

type Server struct {
	bosspb.UnimplementedTaskServiceServer
	handler TaskHandler
}

func NewServer(handler TaskHandler) *Server {
	return &Server{
		handler: handler,
	}
}

func (s *Server) Task(ctx context.Context, req *bosspb.TaskRequest) (*bosspb.TaskResponse, error) {
	log.Printf("Получена задача: user=%s, task=%s, title=%s", req.Username, req.Taskname, req.Title)

	// Генерируем проект через хендлер
	zipData, err := s.handler.GenerateProject(ctx, req)
	if err != nil {
		log.Printf("Ошибка генерации: %v", err)
		return &bosspb.TaskResponse{
			Taskname:     req.Taskname,
			Title:        req.Title,
			Description:  req.Description,
			Status:       "error",
			ErrorMessage: err.Error(),
		}, nil
	}

	log.Printf("Задача выполнена: размер архива=%d байт", len(zipData))

	return &bosspb.TaskResponse{
		Taskname:      req.Taskname,
		Title:         req.Title,
		Description:   req.Description,
		Solution:      zipData,
		Status:        "success",
		ArchiveSize:   int64(len(zipData)),
		ArchiveFormat: "zip",
	}, nil
}

func Start(port string, handler TaskHandler) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)

	bosspb.RegisterTaskServiceServer(grpcServer, NewServer(handler))

	log.Printf("gRPC сервер запущен на порту %s", port)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
