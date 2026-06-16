package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"orchestrator/internal/fetcher/grpc/bosspb"
	"orchestrator/internal/memory"
	"orchestrator/internal/redis"
	"orchestrator/internal/service/rules/boss"
)

// Server — реализация bosspb.BossServiceServer.
// Это единственный внешний gRPC-эндпоинт orchestrator (помимо клиента к agents).
// Он переводит protobuf-структуры в внутренние rules.* и вызывает boss.Service.
type Server struct {
	bosspb.UnimplementedBossServiceServer
	boss        *boss.Service
	redisClient *redis.Client
}

// NewServer — создаёт gRPC-сервер для apigateway
func NewServer(b *boss.Service, r *redis.Client) *Server {
	return &Server{boss: b, redisClient: r}
}

// CreateTaskStream — основной streaming-эндпоинт от apigateway.
// Запускает boss.ExecuteTask и отправляет апдейты прогресса в gRPC-stream.
func (s *Server) CreateTaskStream(req *bosspb.CreateTaskRequest, stream bosspb.BossService_CreateTaskStreamServer) error {
	ctx := stream.Context()
	taskID := generateTaskID()
	log.Printf("CreateTaskStream from %s (user_id=%s): %s (task_id=%s)", req.Username, req.UserId, req.Title, taskID)

	sender := newSender(stream, taskID, req.UserId, s.redisClient)
	progress := sender.asProgressFunc()

	bossReq := &boss.CreateTaskRequest{
		UserId:                 req.UserId,
		Username:               req.Username,
		Title:                  req.Title,
		Description:            req.Description,
		Tokens:                 req.Tokens,
		Meta:                   req.Meta,
		UseAiPlanning:          req.UseAiPlanning,
		PredefinedArchitecture: req.PredefinedArchitecture,
		PredefinedTechStack:    req.PredefinedTechStack,
	}
	for _, manager := range req.PredefinedManagers {
		if manager == nil {
			continue
		}
		workflowManager := boss.ManagerWorkflow{
			Role:        manager.Role,
			Description: manager.Description,
			Priority:    manager.Priority,
		}
		for _, worker := range manager.Workers {
			if worker == nil {
				continue
			}
			workflowManager.Workers = append(workflowManager.Workers, boss.WorkerWorkflow{
				Role:        worker.Role,
				Description: worker.Description,
			})
		}
		bossReq.PredefinedManagers = append(bossReq.PredefinedManagers, workflowManager)
	}

	err := s.boss.ExecuteTask(ctx, bossReq, progress)
	sender.flush() // всегда вызываем flush, и при успехе, и при ошибке

	// Задача завершена: все промежуточные буферы (результаты воркеров,
	// прочитанные файлы, вывод команд) стали мусором. Возвращаем память ОС,
	// чтобы RSS опускался к холостому уровню, а не копился между задачами
	// (issue #89).
	memory.ReleaseToOS()

	if err != nil {
		log.Printf("ExecuteTask error: %v", err)
		return err
	}
	return nil
}

// ResumeTaskStream — переподключение к запущенной задаче.
// Простейшая реализация: проигрывает историю из Redis и выходит.
func (s *Server) ResumeTaskStream(req *bosspb.ResumeTaskStreamRequest, stream bosspb.BossService_CreateTaskStreamServer) error {
	if s.redisClient == nil || !s.redisClient.IsEnabled() {
		return stream.Send(&bosspb.TaskUpdate{
			TaskId:    req.TaskId,
			Message:   "No persisted stream available",
			Status:    "error",
			Timestamp: time.Now().Unix(),
		})
	}
	updates, err := s.redisClient.GetStreamUpdates(stream.Context(), req.TaskId)
	if err != nil {
		return err
	}
	for _, u := range updates {
		up := &bosspb.TaskUpdate{
			TaskId:    u.TaskID,
			Message:   u.Message,
			Progress:  u.Progress,
			Status:    u.Status,
			Timestamp: u.Timestamp,
		}
		if m, ok := u.Data.(map[string]any); ok {
			up.Data = map[string]string{}
			for k, v := range m {
				if s, ok := v.(string); ok {
					up.Data[k] = s
				}
			}
		}
		if err := stream.Send(up); err != nil {
			return err
		}
	}
	return nil
}

// StopTask — помечает задачу как cancelled
func (s *Server) StopTask(ctx context.Context, req *bosspb.StopTaskRequest) (*bosspb.TaskStatusResponse, error) {
	task, err := s.boss.StopTask(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &bosspb.TaskStatusResponse{TaskId: task.ID.String(), Status: task.Status, Progress: "0%"}, nil
}

// GetTaskStatus — возвращает текущий статус задачи
func (s *Server) GetTaskStatus(ctx context.Context, req *bosspb.TaskStatusRequest) (*bosspb.TaskStatusResponse, error) {
	task, err := s.boss.GetTaskStatus(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &bosspb.TaskStatusResponse{TaskId: task.ID.String(), Status: task.Status, Progress: "50%"}, nil
}

// RestoreProjectFiles — восстанавливает проект из Nix store и возвращает файлы
func (s *Server) RestoreProjectFiles(ctx context.Context, req *bosspb.RestoreProjectFilesRequest) (*bosspb.RestoreProjectFilesResponse, error) {
	if req.NixStorePath == "" {
		return nil, fmt.Errorf("nix_store_path is required")
	}

	// Create temp dir for restore
	tmpDir, err := os.MkdirTemp("", "octra-restore-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Restore from Nix store
	if err := s.boss.RestoreProjectFromNix(req.NixStorePath, tmpDir); err != nil {
		return nil, fmt.Errorf("failed to restore from nix store: %w", err)
	}

	// Read files
	codeFiles, err := boss.ReadProjectFiles(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read project files: %w", err)
	}

	// Convert to proto
	entries := make([]*bosspb.CodeFileEntry, 0, len(codeFiles))
	for _, f := range codeFiles {
		entries = append(entries, &bosspb.CodeFileEntry{
			Path:     f.Path,
			Content:  f.Content,
			Language: f.Language,
			Encoding: f.Encoding,
		})
	}

	return &bosspb.RestoreProjectFilesResponse{
		TaskId:     req.TaskId,
		Files:      entries,
		TotalFiles: int32(len(entries)),
	}, nil
}

// CreateTask — legacy unary эндпоинт. Сейчас просто возвращает not-implemented.
// Apigateway использует CreateTaskStream.
func (s *Server) CreateTask(ctx context.Context, req *bosspb.CreateTaskRequest) (*bosspb.BossDecision, error) {
	return &bosspb.BossDecision{Status: "use_stream", ErrorMessage: "use CreateTaskStream"}, nil
}

// generateTaskID — простой идентификатор для логов; реальный UUID создаётся в boss.ExecuteTask
var taskCounter struct {
	sync.Mutex
	n int64
}

func generateTaskID() string {
	taskCounter.Lock()
	defer taskCounter.Unlock()
	taskCounter.n++
	return time.Now().Format("20060102-150405") + "-" + strconv.FormatInt(taskCounter.n, 10)
}

