package grpc

import (
	"context"
	"log"
	"net"

	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/service"

	"google.golang.org/grpc"
)

// Server — gRPC сервер boss сервиса
type Server struct {
	bosspb.UnimplementedBossServiceServer
	service *service.BossService
}

func NewServer(s *service.BossService) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) CreateTask(ctx context.Context, req *bosspb.CreateTaskRequest) (*bosspb.BossDecision, error) {
	return s.service.CreateTask(ctx, req)
}

func (s *Server) GetTaskStatus(ctx context.Context, req *bosspb.TaskStatusRequest) (*bosspb.TaskStatusResponse, error) {
	return s.service.GetTaskStatus(ctx, req)
}

// Start запускает gRPC сервер
func Start(port string, s *service.BossService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)
	bosspb.RegisterBossServiceServer(grpcServer, NewServer(s))

	log.Printf("Boss gRPC сервер запущен на порту %s", port)

	return grpcServer.Serve(lis)
}
