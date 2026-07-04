package grpcserver

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"backend/internal/model"
	pb "backend/proto/chat"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Authenticate(ctx context.Context, token string) (*model.User, error)
}

type DashboardEnvRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.DashboardEnvironment, error)
}

type CanvasNodeRepository interface {
	ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]model.CanvasNode, error)
}

type ChatService interface {
	ChatWithEnvironmentStream(ctx context.Context, user *model.User, envID uuid.UUID, prompt string, w io.Writer) error
}

type Server struct {
	pb.UnimplementedChatServiceServer
	auth       AuthService
	envRepo    DashboardEnvRepository
	nodeRepo   CanvasNodeRepository
	chat       ChatService
}

func New(auth AuthService, envRepo DashboardEnvRepository, nodeRepo CanvasNodeRepository, chat ChatService) *Server {
	return &Server{
		auth:     auth,
		envRepo:  envRepo,
		nodeRepo: nodeRepo,
		chat:     chat,
	}
}

func (s *Server) RegisterWith(reg grpc.ServiceRegistrar) {
	pb.RegisterChatServiceServer(reg, s)
}

func (s *Server) Chat(stream pb.ChatService_ChatServer) error {
	ctx := stream.Context()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	apiKeys := md.Get("x-api-key")
	if len(apiKeys) == 0 {
		apiKeys = md.Get("authorization")
	}
	if len(apiKeys) == 0 {
		return status.Error(codes.Unauthenticated, "missing api key")
	}

	user, err := s.auth.Authenticate(ctx, apiKeys[0])
	if err != nil {
		return status.Error(codes.Unauthenticated, "invalid api key")
	}

	envIDs := md.Get("x-environment-id")
	if len(envIDs) == 0 {
		return status.Error(codes.InvalidArgument, "missing environment id")
	}

	envID, err := uuid.Parse(envIDs[0])
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid environment id")
	}

	env, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return status.Error(codes.NotFound, "environment not found")
	}
	if env.UserID != user.ID {
		return status.Error(codes.PermissionDenied, "not your environment")
	}

	nodes, err := s.nodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return status.Error(codes.Internal, "failed to read canvas nodes")
	}
	if !hasAdapterNode(nodes, "grpc") {
		return status.Error(codes.PermissionDenied, "gRPC adapter not enabled: add an Adapter node with protocol 'gRPC' on the canvas")
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if req.Prompt == "" {
			stream.Send(&pb.ChatResponse{
				Payload: &pb.ChatResponse_Error{Error: "prompt is required"},
				Done:    true,
			})
			continue
		}

		err = s.chat.ChatWithEnvironmentStream(ctx, user, envID, req.Prompt, &grpcStreamWriter{stream: stream})
		if err != nil {
			log.Printf("grpc chat error: %v", err)
			stream.Send(&pb.ChatResponse{
				Payload: &pb.ChatResponse_Error{Error: err.Error()},
				Done:    true,
			})
		} else {
			stream.Send(&pb.ChatResponse{Done: true})
		}
	}
}

// hasAdapterNode checks if any canvas node has kind "adapter" with the given protocol.
func hasAdapterNode(nodes []model.CanvasNode, protocol string) bool {
	for _, n := range nodes {
		if n.Kind != "adapter" {
			continue
		}
		var meta map[string]*string
		if n.Meta != "" {
			json.Unmarshal([]byte(n.Meta), &meta)
		}
		if meta != nil {
			if v := strPtrVal(meta["protocol"]); v == protocol {
				return true
			}
		}
	}
	return false
}

func strPtrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type grpcStreamWriter struct {
	stream pb.ChatService_ChatServer
}

func (w *grpcStreamWriter) Write(p []byte) (int, error) {
	err := w.stream.Send(&pb.ChatResponse{
		Payload: &pb.ChatResponse_Token{Token: string(p)},
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
