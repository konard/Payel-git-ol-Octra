package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"worker/internal/fetcher/grpc/workerpb"
	"worker/internal/service"
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

func (s *Server) AssignWorkers(ctx context.Context, req *workerpb.AssignWorkersRequest) (*workerpb.AssignWorkersResponse, error) {
	return s.service.AssignWorkers(ctx, req)
}

func (s *Server) SubmitResult(ctx context.Context, req *workerpb.WorkerResult) (*workerpb.ReviewResponse, error) {
	return s.service.SubmitResult(ctx, req)
}

// Start запускает gRPC сервер
func Start(port string, s *service.WorkerService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)
	workerpb.RegisterWorkerServiceServer(grpcServer, NewServer(s))

	log.Printf("Worker gRPC сервер запущен на порту %s", port)

	return grpcServer.Serve(lis)
}
