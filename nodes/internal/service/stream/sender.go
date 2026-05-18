package stream

import (
	"context"
	"log"
	"time"

	"nodes/internal/fetcher/grpc/bosspb"
	"nodes/internal/redis"
)

// Sender — обёртка над gRPC-стримом, отправляет апдейты во внешнее API
// и параллельно сохраняет их в Redis для возобновления стрима после reconnect.
type Sender struct {
	stream      bosspb.BossService_CreateTaskStreamServer
	taskID      string
	userID      string
	redisClient *redis.Client
}

// NewSender — конструктор
func NewSender(
	stream bosspb.BossService_CreateTaskStreamServer,
	taskID, userID string,
	redisClient *redis.Client,
) *Sender {
	return &Sender{stream: stream, taskID: taskID, userID: userID, redisClient: redisClient}
}

// Send — отправляет апдейт и сохраняет в Redis
func (s *Sender) Send(update *bosspb.TaskUpdate) error {
	select {
	case <-s.stream.Context().Done():
		return s.stream.Context().Err()
	default:
	}
	if err := s.stream.Send(update); err != nil {
		log.Printf("Failed to send stream update: %v", err)
		return err
	}
	s.persistToRedis(update)
	return nil
}

func (s *Sender) persistToRedis(update *bosspb.TaskUpdate) {
	if s.redisClient == nil || !s.redisClient.IsEnabled() {
		return
	}
	ctx := context.Background()
	var dataInterface any
	if update.Data != nil {
		dataInterface = update.Data
	}
	state := redis.StreamState{
		TaskID:   s.taskID,
		UserID:   s.userID,
		Status:   update.Status,
		Progress: update.Progress,
		Message:  update.Message,
	}
	if err := s.redisClient.UpdateStreamState(ctx, s.taskID, state); err != nil {
		log.Printf("Failed to update Redis stream state: %v", err)
	}
	streamUpdate := redis.StreamUpdate{
		TaskID:    s.taskID,
		Status:    update.Status,
		Progress:  update.Progress,
		Message:   update.Message,
		Data:      dataInterface,
		Timestamp: update.Timestamp,
	}
	if err := s.redisClient.AddStreamUpdate(ctx, s.taskID, streamUpdate); err != nil {
		log.Printf("Failed to add Redis stream update: %v", err)
	}
}

// SendInitial — стартовый апдейт со статусом processing
func (s *Sender) SendInitial(message string, progress int32) error {
	return s.Send(&bosspb.TaskUpdate{
		TaskId:    s.taskID,
		Message:   message,
		Progress:  progress,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})
}

// SendSuccess — финальный апдейт со статусом success
func (s *Sender) SendSuccess(message string, progress int32, data map[string]string) error {
	return s.Send(&bosspb.TaskUpdate{
		TaskId:    s.taskID,
		Message:   message,
		Progress:  progress,
		Status:    "success",
		Timestamp: time.Now().Unix(),
		Data:      data,
	})
}

// SendError — апдейт с ошибкой
func (s *Sender) SendError(message string) error {
	return s.Send(&bosspb.TaskUpdate{
		TaskId:    s.taskID,
		Message:   message,
		Status:    "error",
		Timestamp: time.Now().Unix(),
	})
}
