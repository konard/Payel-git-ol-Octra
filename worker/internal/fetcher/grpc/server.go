package grpc

import (
	"context"
	"net"

	"worker/internal/fetcher/grpc/workerpb"
	"worker/internal/service"

	"google.golang.org/grpc"
)

// Server — gRPC сервер worker сервиса
type Server struct {
	workerpb.UnimplementedWorkerServiceServer
	service *service.WorkerService
}

func NewServer(s *service.WorkerService) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) AssignWorkersAndWait(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	return s.service.AssignWorkersAndWait(ctx, req)
}

// Start запускает gRPC сервер
func Start(port string, s *service.WorkerService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	workerpb.RegisterWorkerServiceServer(grpcServer, NewServer(s))

	return grpcServer.Serve(lis)
}
