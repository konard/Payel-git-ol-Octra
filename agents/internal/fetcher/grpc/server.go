package grpc

import (
	"context"
	"log"
	"net"

	"agents/internal/service"
	"agents/pb"
	"agents/pkg/models"

	"google.golang.org/grpc"
)

// Server wraps the gRPC server and agent service
type Server struct {
	pb.UnimplementedAgentServiceServer
	service *service.AgentService
}

// NewServer creates a new gRPC server
func NewServer(s *service.AgentService) *Server {
	return &Server{
		service: s,
	}
}

// Generate implements the Generate RPC
func (s *Server) Generate(ctx context.Context, req *pb.GenerateRequest) (*pb.GenerateResponse, error) {
	// Convert gRPC request to internal format
	tokens := make(map[string]interface{})
	for k, v := range req.Tokens {
		tokens[k] = v
	}

	// Call agent service
	resp, err := s.service.Generate(ctx, &models.AgentRequest{
		Provider:    req.Provider,
		Model:       req.Model,
		Prompt:      req.Prompt,
		Tokens:      tokens,
		MaxTokens:   int(req.MaxTokens),
		Temperature: req.Temperature,
	})

	if err != nil {
		log.Printf("❌ Generation error: %v", err)
		return &pb.GenerateResponse{
			Provider:  req.Provider,
			Model:     req.Model,
			Error:     err.Error(),
			ErrorCode: "GENERATION_ERROR",
		}, nil
	}

	return &pb.GenerateResponse{
		Provider:  resp.Provider,
		Model:     resp.Model,
		Content:   resp.Content,
		Error:     resp.Error,
		ErrorCode: resp.ErrorCode,
	}, nil
}

// Start starts the gRPC server
func Start(port string, s *service.AgentService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)

	pb.RegisterAgentServiceServer(grpcServer, NewServer(s))

	log.Printf("✅ Agents gRPC server started on port %s", port)

	return grpcServer.Serve(lis)
}
