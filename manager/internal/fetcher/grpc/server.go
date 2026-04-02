package grpc

import (
	"context"

	"manager/internal/fetcher/grpc/managerpb"
	"manager/internal/service"
	"net"

	"google.golang.org/grpc"
)

// Server — gRPC сервер manager сервиса
type Server struct {
	managerpb.UnimplementedManagerServiceServer
	service *service.ManagerService
}

func NewServer(s *service.ManagerService) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) AssignManagersAndWait(ctx context.Context, req *managerpb.AssignManagersRequest) (*managerpb.AssignManagersResponse, error) {
	return s.service.AssignManagersAndWait(ctx, req)
}

// Start запускает gRPC сервер
func Start(port string, s *service.ManagerService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	managerpb.RegisterManagerServiceServer(grpcServer, NewServer(s))

	return grpcServer.Serve(lis)
}
