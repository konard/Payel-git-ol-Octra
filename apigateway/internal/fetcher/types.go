package fetcher

import (
	"net/http"
	"regexp"

	"apigateway/internal/core/ratelimit"
	"apigateway/internal/core/redis"
	"apigateway/internal/core/services"
	"apigateway/internal/fetcher/grpc/boss"
	"apigateway/pkg/requests"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var bossClient *boss.Client
var wsHub *services.Hub
var redisClient *redis.Client
var pubSubManager *redis.PubSubManager

var rl *ratelimit.RateLimiter
var db *gorm.DB

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

// githubTaskURLPattern распознаёт конкретные GitHub issue/pull request ссылки.
// Раньше вставленная ссылка не запускала workflow — чат просто отвечал «Задача
// создана» и ничего не происходило (issue #44). Теперь такая ссылка сразу
// запускает пайплайн создания pull request.
var githubTaskURLPattern = regexp.MustCompile(`(?i)\b(?:https?://)?(?:www\.)?github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/(?:issues|pull|pulls)/[0-9]+`)

type chatWSMessage struct {
	Type        string                      `json:"type"`
	Message     string                      `json:"message"`
	Sender      string                      `json:"sender"`
	TaskPayload *requests.CreateTaskRequest `json:"taskPayload"`
}

// streamUpdateData is the canonical JSON payload for streaming updates to the
// frontend and Redis. It replaces ad-hoc gin.H maps so we marshal once and
// reuse the bytes for WebSocket + Redis state + history + PubSub.
type streamUpdateData struct {
	Type      string `json:"type"`
	TaskID    string `json:"task_id"`
	Message   string `json:"message"`
	Progress  int32  `json:"progress"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data,omitempty"`
	Sender    string `json:"sender,omitempty"`
	IsHistory bool   `json:"is_history,omitempty"`
}

// newStreamID returns a unique identifier for a single task/chat WebSocket
// stream. It is used as the Redis key (STREAM:<id>) and PubSub channel so that
// concurrent streams belonging to the same user (e.g. two browser tabs) never
// share state. The user id is kept as a prefix only for debugging/ownership
// readability; uniqueness comes from the appended UUID.
func newStreamID(userID string) string {
	return userID + ":" + uuid.NewString()
}

func buildStreamUpdate(taskID, msgType, message string, progress int32, timestamp int64, data any, sender string) streamUpdateData {
	return streamUpdateData{
		Type:      msgType,
		TaskID:    taskID,
		Message:   message,
		Progress:  progress,
		Timestamp: timestamp,
		Data:      data,
		Sender:    sender,
	}
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}
