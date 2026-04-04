package grpc

import (
	"context"
	"net"
	"worker/internal/service/worker"

	"google.golang.org/grpc"
	"worker/internal/fetcher/grpc/workerpb"
)

// Server — gRPC сервер worker сервиса
type Server struct {
	workerpb.UnimplementedWorkerServiceServer
	service *worker.WorkerService
}

func NewServer(s *worker.WorkerService) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) AssignWorkersAndWait(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	return s.service.AssignWorkersAndWait(ctx, req)
}

func (s *Server) ReviewWorker(ctx context.Context, req *workerpb.ReviewRequest) (*workerpb.ReviewResponse, error) {
	return s.service.ReviewWorker(ctx, req)
}

// Start запускает gRPC сервер
func Start(port string, s *worker.WorkerService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	workerpb.RegisterWorkerServiceServer(grpcServer, NewServer(s))

	return grpcServer.Serve(lis)
}
