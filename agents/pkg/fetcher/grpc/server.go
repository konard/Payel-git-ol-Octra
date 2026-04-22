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
	log.Printf("Generate RPC called: provider=%s, model=%s, tokens=%+v", req.Provider, req.Model, req.Tokens)
	// Convert gRPC request to internal format
	tokens := make(map[string]interface{})
	for k, v := range req.Tokens {
		tokens[k] = v
	}
	log.Printf("Converted tokens: %+v", tokens)

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

// GenerateStream implements the streaming Generate RPC
func (s *Server) GenerateStream(req *pb.GenerateRequest, stream pb.AgentService_GenerateStreamServer) error {
	tokens := make(map[string]interface{})
	for k, v := range req.Tokens {
		tokens[k] = v
	}

	ctx := stream.Context()
	resp, err := s.service.Generate(ctx, &models.AgentRequest{
		Provider:    req.Provider,
		Model:       req.Model,
		Prompt:      req.Prompt,
		Tokens:      tokens,
		MaxTokens:   int(req.MaxTokens),
		Temperature: req.Temperature,
	})

	if err != nil {
		log.Printf("❌ Stream generation error: %v", err)
		stream.Send(&pb.GenerateStreamChunk{
			Error: err.Error(),
			Done:  true,
		})
		return nil
	}

	if resp.Error != "" {
		stream.Send(&pb.GenerateStreamChunk{
			Error: resp.Error,
			Done:  true,
		})
		return nil
	}

	// Send content in chunks (for streaming simulation, send full content as one chunk)
	chunkSize := 500
	content := resp.Content
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		stream.Send(&pb.GenerateStreamChunk{
			Content: content[i:end],
			Done:    false,
		})
	}

	// Send final chunk
	stream.Send(&pb.GenerateStreamChunk{
		Done:       true,
		TokensUsed: int32(resp.Tokens),
	})

	return nil
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
