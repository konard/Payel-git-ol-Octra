package grpc

import (
	"boss/internal/service/boss"
	"context"
	"log"
	"net"

	"boss/internal/fetcher/grpc/bosspb"

	"google.golang.org/grpc"
)

type Server struct {
	bosspb.UnimplementedBossServiceServer
	service *boss.BossService
}

func NewServer(s *boss.BossService) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) CreateTask(ctx context.Context, req *bosspb.CreateTaskRequest) (*bosspb.BossDecision, error) {
	return s.service.CreateTask(ctx, req)
}

func (s *Server) CreateTaskStream(req *bosspb.CreateTaskRequest, stream bosspb.BossService_CreateTaskStreamServer) error {
	return s.service.CreateTaskStream(req, stream)
}

func (s *Server) ResumeTaskStream(req *bosspb.ResumeTaskStreamRequest, stream bosspb.BossService_CreateTaskStreamServer) error {
	return s.service.ResumeTaskStream(req, stream)
}

func (s *Server) GetTaskStatus(ctx context.Context, req *bosspb.TaskStatusRequest) (*bosspb.TaskStatusResponse, error) {
	return s.service.GetTaskStatus(ctx, req)
}

func (s *Server) StopTask(ctx context.Context, req *bosspb.StopTaskRequest) (*bosspb.TaskStatusResponse, error) {
	return s.service.StopTask(ctx, req)
}

func Start(port string, s *boss.BossService) error {
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
