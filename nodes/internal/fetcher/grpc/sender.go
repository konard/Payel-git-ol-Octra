package grpc

import (
	"log"
	"time"

	"nodes/internal/fetcher/grpc/bosspb"
	"nodes/internal/redis"
	"nodes/internal/service/rules"
	"nodes/internal/service/stream"
)

// streamSender — мост между rules.ProgressFunc и gRPC-стримом apigateway.
// Использует одну goroutine для всех вызовов stream.Send — это устраняет
// HTTP/2 data-race, который существовал в старом boss-сервисе.
type streamSender struct {
	inner  *stream.Sender
	taskID string
	sendCh chan *bosspb.TaskUpdate
	doneCh chan struct{}
}

// newSender — создаёт и запускает отдельную goroutine-отправитель
func newSender(grpcStream bosspb.BossService_CreateTaskStreamServer, taskID, userID string, redisClient *redis.Client) *streamSender {
	s := &streamSender{
		inner:  stream.NewSender(grpcStream, taskID, userID, redisClient),
		taskID: taskID,
		sendCh: make(chan *bosspb.TaskUpdate, 64),
		doneCh: make(chan struct{}),
	}
	go s.loop(grpcStream)
	return s
}

// loop — последовательно отправляет апдейты в gRPC-стрим
func (s *streamSender) loop(grpcStream bosspb.BossService_CreateTaskStreamServer) {
	defer close(s.doneCh)
	ctx := grpcStream.Context()
	for update := range s.sendCh {
		select {
		case <-ctx.Done():
			log.Printf("stream context cancelled, stopping sender")
			return
		default:
			if err := s.inner.Send(update); err != nil {
				log.Printf("stream send error: %v", err)
				return
			}
		}
	}
}

// asProgressFunc — возвращает callback, который оборачивает ProgressFunc
// и кладёт TaskUpdate в очередь отправки
func (s *streamSender) asProgressFunc() rules.ProgressFunc {
	return func(progress int32, message string, data map[string]string) {
		status := "processing"
		if progress >= 100 {
			status = "success"
		}
		if data != nil && data["status"] == "error" {
			status = "error"
			delete(data, "status")
		}
		update := &bosspb.TaskUpdate{
			TaskId:    s.taskID,
			Message:   message,
			Progress:  progress,
			Status:    status,
			Timestamp: time.Now().Unix(),
			Data:      data,
		}
		select {
		case s.sendCh <- update:
		default:
			log.Printf("stream send channel full, dropping update: %s", message)
		}
	}
}

// flush — закрывает канал отправки и дожидается завершения goroutine
func (s *streamSender) flush() {
	close(s.sendCh)
	<-s.doneCh
}
