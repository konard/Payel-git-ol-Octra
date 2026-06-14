package grpc

import (
	"log"
	"time"

	"orchestrator/internal/fetcher/grpc/bosspb"
	"orchestrator/internal/redis"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/stream"
)

// streamSender — мост между rules.ProgressFunc и gRPC-стримом apigateway.
// Использует одну goroutine для всех вызовов stream.Send — это устраняет
// HTTP/2 data-race, который существовал в старом boss-сервисе.
type streamSender struct {
	inner     *stream.Sender
	taskID    string
	sendCh    chan *bosspb.TaskUpdate
	doneCh    chan struct{}
	closedCh  chan struct{} // закрыт после flush/discard
}

// newSender — создаёт и запускает отдельную goroutine-отправитель
func newSender(grpcStream bosspb.BossService_CreateTaskStreamServer, taskID, userID string, redisClient *redis.Client) *streamSender {
	s := &streamSender{
		inner:    stream.NewSender(grpcStream, taskID, userID, redisClient),
		taskID:   taskID,
		sendCh:   make(chan *bosspb.TaskUpdate, 256),
		doneCh:   make(chan struct{}),
		closedCh: make(chan struct{}),
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
			log.Printf("stream context cancelled, draining sender")
			// Продолжаем дрэйнить канал, пока не будет закрыт
			// чтобы asProgressFunc не блокировал вызывающие горутины
			continue
		default:
			if err := s.inner.Send(update); err != nil {
				log.Printf("stream send error: %v", err)
				return
			}
		}
	}
}

// asProgressFunc — возвращает callback, который оборачивает ProgressFunc
// и кладёт TaskUpdate в очередь отправки. Использует неблокирующую отправку
// чтобы не допустить утечку горутин при закрытии stream.
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
		// Неблокирующая отправка: если канал полон или закрыт — дропаем апдейт.
		// Это предотвращает утечку горутин когда stream уже закрыт но ExecuteTask
		// ещё работает (менеджеры/воркеры продолжают вызывать progress).
		select {
		case s.sendCh <- update:
		default:
			log.Printf("sendCh full or closed, dropping progress update: %s", message)
		}
	}
}

// flush — закрывает канал отправки и дожидается завершения goroutine.
// Вызывать ВСЕГДА — и при успехе, и при ошибке.
func (s *streamSender) flush() {
	select {
	case <-s.closedCh:
		return // уже закрыт
	default:
	}
	close(s.closedCh)
	close(s.sendCh)
	<-s.doneCh
}

// discard — закрывает канал без ожидания (для аварийных случаев).
func (s *streamSender) discard() {
	select {
	case <-s.closedCh:
		return
	default:
	}
	close(s.closedCh)
	close(s.sendCh)
}
