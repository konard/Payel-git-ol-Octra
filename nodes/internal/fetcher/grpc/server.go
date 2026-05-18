package grpc

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"nodes/internal/fetcher/grpc/bosspb"
	"nodes/internal/redis"
	"nodes/internal/service/rules/boss"
)

// Server — реализация bosspb.BossServiceServer.
// Это единственный внешний gRPC-эндпоинт nodes (помимо клиента к agents).
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
		UserId:        req.UserId,
		Username:      req.Username,
		Title:         req.Title,
		Description:   req.Description,
		Tokens:        req.Tokens,
		Meta:          req.Meta,
		UseAiPlanning: req.UseAiPlanning,
	}

	if err := s.boss.ExecuteTask(ctx, bossReq, progress); err != nil {
		log.Printf("ExecuteTask error: %v", err)
		return err
	}
	sender.flush()
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
