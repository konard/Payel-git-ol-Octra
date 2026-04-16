package boss

import (
	"boss/internal/fetcher/grpc/bosspb"
	"boss/internal/redis"
	"context"
	"log"
	"time"
)

// streamSender wraps a gRPC stream sender with Redis persistence
type streamSender struct {
	stream      bosspb.BossService_CreateTaskStreamServer
	taskID      string
	userID      string
	redisClient *redis.Client
}

// newStreamSender creates a new stream sender
func newStreamSender(
	stream bosspb.BossService_CreateTaskStreamServer,
	taskID string,
	userID string,
	redisClient *redis.Client,
) *streamSender {
	return &streamSender{
		stream:      stream,
		taskID:      taskID,
		userID:      userID,
		redisClient: redisClient,
	}
}

// send sends a task update and persists it in Redis
func (s *streamSender) send(update *bosspb.TaskUpdate) error {
	// Check if context is cancelled
	select {
	case <-s.stream.Context().Done():
		return s.stream.Context().Err()
	default:
	}

	// Send via gRPC stream
	if err := s.stream.Send(update); err != nil {
		log.Printf("❌ Failed to send stream update: %v", err)
		return err
	}

	// Persist in Redis
	s.persistToRedis(update)

	return nil
}

// persistToRedis saves the update to Redis asynchronously
func (s *streamSender) persistToRedis(update *bosspb.TaskUpdate) {
	if s.redisClient == nil || !s.redisClient.IsEnabled() {
		return
	}

	ctx := context.Background()

	// Convert data map to interface for Redis
	var dataInterface any
	if update.Data != nil {
		dataInterface = update.Data
	}

	// Update stream state
	state := redis.StreamState{
		TaskID:   s.taskID,
		UserID:   s.userID,
		Status:   update.Status,
		Progress: update.Progress,
		Message:  update.Message,
	}

	if err := s.redisClient.UpdateStreamState(ctx, s.taskID, state); err != nil {
		log.Printf("❌ Failed to update Redis stream state: %v", err)
	}

	// Add stream update
	streamUpdate := redis.StreamUpdate{
		TaskID:    s.taskID,
		Status:    update.Status,
		Progress:  update.Progress,
		Message:   update.Message,
		Data:      dataInterface,
		Timestamp: update.Timestamp,
	}

	if err := s.redisClient.AddStreamUpdate(ctx, s.taskID, streamUpdate); err != nil {
		log.Printf("❌ Failed to add Redis stream update: %v", err)
	}
}

// sendInitial sends the initial update with custom message and progress
func (s *streamSender) sendInitial(message string, progress int32) error {
	return s.send(&bosspb.TaskUpdate{
		TaskId:    s.taskID,
		Message:   message,
		Progress:  progress,
		Status:    "processing",
		Timestamp: time.Now().Unix(),
	})
}

// sendSuccess sends a success update
func (s *streamSender) sendSuccess(message string, progress int32, data map[string]string) error {
	return s.send(&bosspb.TaskUpdate{
		TaskId:    s.taskID,
		Message:   message,
		Progress:  progress,
		Status:    "success",
		Timestamp: time.Now().Unix(),
		Data:      data,
	})
}

// sendError sends an error update
func (s *streamSender) sendError(message string) error {
	return s.send(&bosspb.TaskUpdate{
		TaskId:    s.taskID,
		Message:   message,
		Status:    "error",
		Timestamp: time.Now().Unix(),
	})
}
