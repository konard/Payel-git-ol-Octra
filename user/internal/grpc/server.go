package grpc

import (
	"context"
	"fmt"

	"user/internal/cache"
	"user/internal/core/services"
	"user/internal/grpc/userfilespb"
)

type UserFilesServer struct {
	userfilespb.UnimplementedUserFilesServiceServer
	orchestrator *OrchestratorClient
	cache        *cache.NixCache
}

func NewUserFilesServer(orch *OrchestratorClient, c *cache.NixCache) *UserFilesServer {
	return &UserFilesServer{
		orchestrator: orch,
		cache:        c,
	}
}

func (s *UserFilesServer) GetProjectFiles(ctx context.Context, req *userfilespb.GetProjectFilesRequest) (*userfilespb.GetProjectFilesResponse, error) {
	if req.ChatId == "" {
		return nil, fmt.Errorf("chat_id is required")
	}

	chat, err := services.GetChatByID(req.ChatId)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	if chat.NixStorePath == "" {
		return &userfilespb.GetProjectFilesResponse{
			ChatId: req.ChatId,
			Files:  []*userfilespb.CodeFileEntry{},
		}, nil
	}

	taskID := chat.TaskId

	if cached, ok := s.cache.Get(taskID); ok {
		entries := make([]*userfilespb.CodeFileEntry, 0, len(cached))
		for _, f := range cached {
			entries = append(entries, &userfilespb.CodeFileEntry{
				Path:     f.Path,
				Content:  f.Content,
				Language: f.Language,
				Encoding: f.Encoding,
			})
		}
		return &userfilespb.GetProjectFilesResponse{
			ChatId:     req.ChatId,
			TaskId:     taskID,
			Files:      entries,
			TotalFiles: int32(len(entries)),
		}, nil
	}

	resp, err := s.orchestrator.RestoreProjectFiles(ctx, chat.NixStorePath, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to restore files: %w", err)
	}

	cacheFiles := make([]cache.CodeFileEntry, 0, len(resp.Files))
	for _, f := range resp.Files {
		cacheFiles = append(cacheFiles, cache.CodeFileEntry{
			Path:     f.Path,
			Content:  f.Content,
			Language: f.Language,
			Encoding: f.Encoding,
		})
	}
	s.cache.Set(taskID, cacheFiles)

	entries := make([]*userfilespb.CodeFileEntry, 0, len(resp.Files))
	for _, f := range resp.Files {
		entries = append(entries, &userfilespb.CodeFileEntry{
			Path:     f.Path,
			Content:  f.Content,
			Language: f.Language,
			Encoding: f.Encoding,
		})
	}

	return &userfilespb.GetProjectFilesResponse{
		ChatId:     req.ChatId,
		TaskId:     taskID,
		Files:      entries,
		TotalFiles: resp.TotalFiles,
	}, nil
}
