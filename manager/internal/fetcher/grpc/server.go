package grpc

import (
	"context"
	"manager/internal/service/manager"

	"manager/internal/fetcher/grpc/managerpb"
	"net"

	"google.golang.org/grpc"
)

// Server — gRPC сервер manager сервиса
type Server struct {
	managerpb.UnimplementedManagerServiceServer
	service *manager.ManagerService
}

func NewServer(s *manager.ManagerService) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) AssignManagersAndWait(ctx context.Context, req *managerpb.AssignManagersRequest) (*managerpb.AssignManagersResponse, error) {
	return s.service.AssignManagersAndWait(ctx, req)
}

func (s *Server) AssignManager(ctx context.Context, req *managerpb.AssignManagerRequest) (*managerpb.ManagerResult, error) {
	return s.service.AssignManager(ctx, req)
}

func (s *Server) AssignManagerStream(req *managerpb.AssignManagerRequest, stream managerpb.ManagerService_AssignManagerStreamServer) error {
	return s.service.AssignManagerStream(req, stream)
}

// Start запускает gRPC сервер
func Start(port string, s *manager.ManagerService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	managerpb.RegisterManagerServiceServer(grpcServer, NewServer(s))

	return grpcServer.Serve(lis)
}
