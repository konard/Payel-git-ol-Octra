package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"apigateway/internal/core/ratelimit"
	"apigateway/internal/core/redis"
	"apigateway/internal/core/services"
	"apigateway/internal/fetcher/grpc/boss"
	"apigateway/internal/fetcher/grpc/boss/bosspb"
	"apigateway/pkg/models"
	"apigateway/pkg/requests"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newStreamID returns a unique identifier for a single task/chat WebSocket
// stream. It is used as the Redis key (STREAM:<id>) and PubSub channel so that
// concurrent streams belonging to the same user (e.g. two browser tabs) never
// share state. The user id is kept as a prefix only for debugging/ownership
// readability; uniqueness comes from the appended UUID.
func newStreamID(userID string) string {
	return userID + ":" + uuid.NewString()
}

var bossClient *boss.Client
var wsHub *services.Hub
var redisClient *redis.Client
var pubSubManager *redis.PubSubManager

func validateJWT(tokenString string) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("JWT_SECRET not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid claims")
	}

	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)
	if userID == "" {
		return "", "", errors.New("user_id missing in token")
	}
	return userID, username, nil
}

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

type chatWSMessage struct {
	Type        string                      `json:"type"`
	Message     string                      `json:"message"`
	Sender      string                      `json:"sender"`
	TaskPayload *requests.CreateTaskRequest `json:"taskPayload"`
}

// PingWriter periodically sends pings to keep connection alive
func PingWriter(conn *websocket.Conn, done <-chan struct{}) {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				log.Printf("❌ Ping write error: %v", err)
				return
			}
		case <-done:
			return
		}
	}
}

func init() {
	var err error
	bossHost := os.Getenv("BOSS_SERVICE_HOST")
	if bossHost == "" {
		bossHost = "orchestrator:50052"
	}
	bossClient, err = boss.NewClient(bossHost)
	if err != nil {
		log.Printf("Warning: failed to connect to Boss service: %v", err)
	}

	redisClient = redis.NewClient()
	if redisClient.IsEnabled() {
		pubSubManager = redis.NewPubSubManager(redisClient)
		log.Println("[Redis] PubSub manager initialized")
	}

	wsHub = services.NewHub()
	go wsHub.Run()

	rl = ratelimit.New()

	// Initialize database
	db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: failed to connect to database: %v", err)
	} else {
		db.AutoMigrate(&models.Task{})
	}
}

// RegisterRoutes registers all HTTP routes on the gin engine
func RegisterRoutes(r *gin.Engine) {
	// Rate limit middleware wrappers
	rlHealth := rl.GinMiddleware("health")
	rlTaskCreate := rl.GinMiddleware("task_create")
	rlTaskStatus := rl.GinMiddleware("task_status")
	rlTaskReconnect := rl.GinMiddleware("task_reconnect")

	r.GET("/health", rlHealth, healthHandler)
	r.GET("/task/create", rlTaskCreate, handleTaskCreateWS)
	r.GET("/task/reconnect", rlTaskReconnect, handleTaskReconnectWS)
	r.GET("/task/status", rlTaskStatus, handleTaskStatus)
	r.POST("/task/:taskId/stop", rlTaskStatus, handleTaskStop)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// handleTaskCreateWS upgrades HTTP to WebSocket and processes task creation
func handleTaskCreateWS(c *gin.Context) {
	// Check JWT from cookie or Authorization header
	token, _ := c.Cookie("access_token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	userID, username, err := validateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	// Read initial message
	_, data, err := conn.ReadMessage()
	if err != nil {
		log.Printf("❌ Failed to read initial WebSocket message: %v", err)
		conn.Close()
		return
	}

	// Ignore empty/whitespace frames
	if len(bytes.TrimSpace(data)) == 0 {
		conn.Close()
		return
	}

	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Type == "chat" {
		// Unique per-connection stream id keeps each tab/chat isolated in Redis.
		streamID := newStreamID(userID)
		conn.WriteJSON(gin.H{
			"type":      "connected",
			"task_id":   streamID,
			"message":   "Connected to boss chat",
			"timestamp": time.Now().Unix(),
		})
		handleBossChatWS(conn, streamID, userID, username, data)
		return
	}

	// Parse task request
	var taskReq requests.CreateTaskRequest
	if err := json.Unmarshal(data, &taskReq); err != nil {
		log.Printf("❌ Failed to parse task request: %v", err)
		conn.WriteJSON(gin.H{
			"type":    "error",
			"message": "Invalid JSON: " + err.Error(),
		})
		conn.Close()
		return
	}

	// Force authenticated user data (ignore client values)
	taskReq.UserID = userID
	taskReq.Username = username

	if taskReq.Title == "" {
		taskReq.Title = "Untitled Task"
	}

	log.Printf("✅ Authenticated user: %s (%s)", taskReq.Username, taskReq.UserID)

	// Unique per-connection stream id keeps each tab/task isolated in Redis so
	// that history from one tab never leaks into another tab on reconnect.
	streamID := newStreamID(userID)

	// Send confirmation
	conn.WriteJSON(gin.H{
		"type":      "connected",
		"task_id":   streamID,
		"message":   "Connected to task creation service",
		"timestamp": time.Now().Unix(),
	})

	// Process task stream in background
	go processTaskStreamWS(conn, taskReq, streamID)

	// Keep connection alive with periodic pings
	done := make(chan struct{})
	go PingWriter(conn, done)
	defer close(done)

	// Keep reading to handle close/ping and messages
	go func() {
		defer conn.Close()
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if msgType == websocket.CloseMessage {
				return
			}
			if msgType == websocket.PingMessage {
				conn.WriteMessage(websocket.PongMessage, nil)
			}
			if msgType == websocket.TextMessage {
				// Handle ping messages from frontend
				message := string(data)
				if message == `{"type":"ping"}` {
					conn.WriteJSON(gin.H{"type": "pong"})
				}
			}
		}
	}()
}

func handleBossChatWS(conn *websocket.Conn, streamID, userID, username string, initial []byte) {
	defer conn.Close()

	done := make(chan struct{})
	go PingWriter(conn, done)
	defer close(done)

	handleFrame := func(data []byte) bool {
		if len(bytes.TrimSpace(data)) == 0 {
			return false
		}

		var msg chatWSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Invalid chat JSON: " + err.Error(),
			})
			return false
		}

		if msg.Type == "ping" {
			conn.WriteJSON(gin.H{"type": "pong"})
			return false
		}
		if msg.Type != "chat" {
			return false
		}

		message := strings.TrimSpace(msg.Message)
		if message == "" {
			return false
		}

		if shouldLaunchWorkflowFromChat(message) {
			if msg.TaskPayload == nil {
				writeBossChatMessage(conn, streamID, "I can start a workflow from chat, but I need your model settings first.", true)
				return false
			}

			taskReq := *msg.TaskPayload
			taskReq.UserID = userID
			taskReq.Username = username
			if taskReq.Title == "" {
				taskReq.Title = message
			}
			if taskReq.Description == "" {
				taskReq.Description = message
			}
			if taskReq.Meta == nil {
				taskReq.Meta = map[string]interface{}{}
			}

			writeBossChatMessage(conn, streamID, "I'll start the workflow for this request.", false)
			processTaskStreamWS(conn, taskReq, streamID)
			return true
		}

		writeBossChatMessage(conn, streamID, buildBossChatReply(message), false)
		return false
	}

	if handleFrame(initial) {
		return
	}

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		switch msgType {
		case websocket.CloseMessage:
			return
		case websocket.PingMessage:
			conn.WriteMessage(websocket.PongMessage, nil)
		case websocket.TextMessage:
			if handleFrame(data) {
				return
			}
		}
	}
}

func writeBossChatMessage(conn *websocket.Conn, taskID, message string, clarification bool) {
	conn.WriteJSON(gin.H{
		"type":             "chat",
		"task_id":          taskID,
		"sender":           "boss",
		"message":          message,
		"is_clarification": clarification,
		"timestamp":        time.Now().Unix(),
	})
}

// githubTaskURLPattern распознаёт конкретные GitHub issue/pull request ссылки.
// Раньше вставленная ссылка не запускала workflow — чат просто отвечал «Задача
// создана» и ничего не происходило (issue #44). Теперь такая ссылка сразу
// запускает пайплайн создания pull request.
var githubTaskURLPattern = regexp.MustCompile(`(?i)\b(?:https?://)?(?:www\.)?github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/(?:issues|pull|pulls)/[0-9]+`)

// containsGitHubTaskURL сообщает, есть ли в сообщении ссылка на конкретный
// GitHub issue или pull request, которую нужно превратить в задачу.
func containsGitHubTaskURL(message string) bool {
	return githubTaskURLPattern.MatchString(message)
}

func shouldLaunchWorkflowFromChat(message string) bool {
	// Вставленная ссылка на GitHub issue/PR — это всегда задача, даже если в
	// сообщении нет других ключевых слов.
	if containsGitHubTaskURL(message) {
		return true
	}

	words := normalizedWords(message)
	if shouldLaunchSearchWorkflowFromChat(words) {
		return true
	}

	triggers := []string{
		"build", "building", "create", "creating", "develop", "fix", "generate", "generating",
		"implement", "implementing", "launch", "make", "refactor", "run", "scaffold", "start", "write",
		"введи", "добавь", "запусти", "исправь", "напиши", "подготовь", "разработай",
		"сгенерируй", "сделай", "создай", "создать", "собери",
	}
	targets := []string{
		"app", "application", "api", "backend", "bug", "code", "component", "feature", "frontend", "proxy",
		"integration", "page", "project", "service", "site", "tool", "webapp", "website", "workflow",
		"бот", "доклад", "документ", "код", "презентацией", "презентации", "презентацию",
		"презентация", "приложение", "проект", "прокси", "сайт", "сервис", "слайд", "слайдами",
		"слайды", "функцию",
	}
	return hasAnyWord(words, triggers) && hasAnyWord(words, targets)
}

func shouldLaunchSearchWorkflowFromChat(words map[string]bool) bool {
	searchTriggers := []string{
		"find", "google", "lookup", "research", "search",
		"найди", "погугли", "поиск", "поищи",
	}
	if hasAnyWord(words, searchTriggers) || (words["look"] && words["up"]) {
		return true
	}

	searchTargets := []string{
		"configure", "docs", "documentation", "install", "latest", "links", "news", "reference", "setup",
		"документация", "документацию", "настроить", "новости", "ссылки", "установить",
	}
	if hasAnyWord(words, []string{"how", "как"}) && hasAnyWord(words, searchTargets) {
		return true
	}

	if len(words) > 1 && hasAnyWord(words, searchTargets) {
		return true
	}

	return false
}

// isRussian reports whether the message contains Cyrillic letters. The chat
// answers in the same language the user wrote in, so a greeting like «привет»
// is met with a Russian reply instead of the previous English-only canned text
// (issue #70: the chat ignored casual/Russian messages and felt dead).
func isRussian(message string) bool {
	for _, r := range message {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

// buildBossChatReply produces a conversational reply for messages that are not
// workflow/search requests. The chat must behave like a normal assistant — a
// plain «привет» or "hello" should get a friendly answer rather than silence or
// a generic English line (issue #70). Replies are returned in the user's
// language (Russian when the message contains Cyrillic, English otherwise).
func buildBossChatReply(message string) string {
	ru := isRussian(message)
	words := normalizedWords(message)

	// Greetings — hello / hi / привет / здравствуйте / добрый день …
	if hasAnyWord(words, []string{
		"hello", "hi", "hey", "yo", "hiya", "howdy",
		"привет", "приветик", "прив", "здравствуй", "здравствуйте", "здарова",
		"хай", "ку", "салют", "добрый", "доброе",
	}) {
		if ru {
			return "Привет! Я Octra. Можем просто пообщаться, а можно описать, что нужно собрать или найти, — и я возьмусь за задачу."
		}
		return "Hi! I'm Octra. We can just chat, or you can describe what to build or look up and I'll get to work."
	}

	// How are you — как дела / как ты / how are you …
	if hasAnyWord(words, []string{"дела", "поживаешь"}) ||
		(words["how"] && words["you"] && (words["are"] || words["doing"])) {
		if ru {
			return "У меня всё отлично, спасибо! Чем помочь — пообщаться, поискать что-то в интернете или собрать проект?"
		}
		return "I'm doing great, thanks! Want to chat, research something, or have me build a project?"
	}

	// Thanks — thanks / thank you / спасибо / благодарю …
	if hasAnyWord(words, []string{"thanks", "thank", "thx", "спасибо", "благодарю", "спс", "пасиб"}) {
		if ru {
			return "Всегда пожалуйста! Если понадобится что-то ещё — просто напишите."
		}
		return "You're welcome! Let me know if there's anything else."
	}

	// Farewell — bye / goodbye / пока / до свидания …
	if hasAnyWord(words, []string{"bye", "goodbye", "cya", "пока", "свидания", "увидимся", "прощай"}) {
		if ru {
			return "Пока! Возвращайтесь, когда понадобится помощь."
		}
		return "Bye! Come back whenever you need a hand."
	}

	// Identity / capabilities / help — who are you / что ты умеешь / помоги …
	if hasAnyWord(words, []string{"help", "помощь", "помоги", "умеешь", "можешь"}) ||
		(words["who"] && words["you"]) || (words["what"] && (words["can"] || words["do"]) && words["you"]) ||
		(words["кто"] && words["ты"]) || (words["что"] && words["ты"]) {
		if ru {
			return "Я Octra — фабрика ИИ-агентов. Могу просто общаться, искать информацию в интернете или собрать проект целиком. Например: «создай php сервер» или «найди документацию по httpx»."
		}
		return "I'm Octra, an AI agent factory. I can chat, research the web, or build a whole project for you. Try: \"create a php server\" or \"find the httpx docs\"."
	}

	if ru {
		return "Понял вас. Опишите задачу — и я возьмусь за работу, либо продолжим общение."
	}
	return "Got it. Describe a task and I'll get to work, or we can keep chatting."
}

func normalizedWords(message string) map[string]bool {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, message)

	words := make(map[string]bool)
	for _, word := range strings.Fields(cleaned) {
		words[word] = true
	}
	return words
}

func hasAnyWord(words map[string]bool, candidates []string) bool {
	for _, candidate := range candidates {
		if words[candidate] {
			return true
		}
	}
	return false
}

// processTaskStreamWS sends task to Boss and streams updates back via WebSocket.
// streamID is the unique per-connection identifier used for Redis stream state,
// the updates history list and the PubSub channel, so concurrent streams from
// the same user stay isolated.
func processTaskStreamWS(conn *websocket.Conn, taskReq requests.CreateTaskRequest, streamID string) {
	defer conn.Close()

	// Convert Meta from map[string]interface{} to map[string]string
	meta := make(map[string]string)
	for k, v := range taskReq.Meta {
		if s, ok := v.(string); ok {
			meta[k] = s
		} else {
			// Convert to JSON string for complex values
			if jsonBytes, err := json.Marshal(v); err == nil {
				meta[k] = string(jsonBytes)
			}
		}
	}

	grpcReq := &bosspb.CreateTaskRequest{
		UserId:      taskReq.UserID,
		Username:    taskReq.Username,
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Tokens:      taskReq.Tokens,
		Meta:        meta,
	}

	// Add predefined workflow if provided
	if taskReq.Workflow != nil {
		wf := taskReq.Workflow
		grpcReq.UseAiPlanning = wf.UseAIPlanning
		grpcReq.PredefinedArchitecture = wf.Architecture
		grpcReq.PredefinedTechStack = wf.TechStack

		for _, mgr := range wf.Managers {
			protoMgr := &bosspb.ManagerConfig{
				Role:        mgr.Role,
				Description: mgr.Description,
				Priority:    mgr.Priority,
			}
			for _, w := range mgr.Workers {
				protoMgr.Workers = append(protoMgr.Workers, &bosspb.WorkerRole{
					Role:        w.Role,
					Description: w.Description,
				})
			}
			grpcReq.PredefinedManagers = append(grpcReq.PredefinedManagers, protoMgr)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Initial progress
	wsWriteJSON(conn, streamID, gin.H{
		"type":      "progress",
		"task_id":   streamID,
		"message":   "Connecting to Boss service...",
		"progress":  5,
		"timestamp": time.Now().Unix(),
	})

	stream, err := bossClient.CreateTaskStream(ctx, grpcReq)
	if err != nil {
		log.Printf("❌ Error calling CreateTaskStream: %v", err)
		conn.WriteJSON(gin.H{
			"type":    "error",
			"message": "Failed to connect to Boss service: " + err.Error(),
		})
		return
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("❌ Stream error: %v", err)
			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Connection error: " + err.Error(),
			})
			return
		}

		// Determine message type
		messageType := update.Status
		if update.IsChat {
			messageType = "chat"
		}

		wsUpdate := gin.H{
			"type":      messageType,
			"task_id":   update.TaskId,
			"message":   update.Message,
			"progress":  update.Progress,
			"timestamp": update.Timestamp,
		}
		if update.Data != nil {
			wsUpdate["data"] = update.Data
		}

		// Add sender for chat messages
		if update.IsChat {
			wsUpdate["sender"] = "boss"
		}

		wsWriteJSONWithRedis(conn, streamID, wsUpdate)

		if update.Status == "success" || update.Status == "error" {
			wsHub.Broadcast(streamID, wsUpdate)
			// When the workflow finishes successfully the boss reports back to the
			// user in the chat ("отчитаться") so a completed task never ends in
			// silence (issue #70). Errors keep their existing red status banner.
			if update.Status == "success" {
				sendCompletionReport(conn, streamID, taskReq, update.Data)
			}
			return
		}
	}
}

// sendCompletionReport posts a short boss chat message summarizing a finished
// task. It folds in the boss's own answer (chatSummary, used by
// research/document tasks) and a link to the result when one is available, and
// is written in the language of the original request (issue #70).
func sendCompletionReport(conn *websocket.Conn, taskID string, taskReq requests.CreateTaskRequest, data map[string]string) {
	writeBossChatMessage(conn, taskID, buildCompletionReport(taskReq, data), false)
}

// buildCompletionReport composes the boss's completion message in the language
// of the original request, folding in the boss's own answer (chatSummary) and a
// result link when available.
func buildCompletionReport(taskReq requests.CreateTaskRequest, data map[string]string) string {
	ru := isRussian(taskReq.Title + " " + taskReq.Description)

	link := data["pullRequestUrl"]
	if link == "" {
		link = data["repoUrl"]
	}
	if link == "" {
		link = data["zipUrl"]
	}

	var b strings.Builder
	if ru {
		b.WriteString("✅ Готово! Я завершил задачу")
	} else {
		b.WriteString("✅ Done! I've finished the task")
	}
	if title := strings.TrimSpace(taskReq.Title); title != "" {
		b.WriteString(" «" + title + "»")
	}
	b.WriteString(".")

	if summary := strings.TrimSpace(data["chatSummary"]); summary != "" {
		b.WriteString("\n\n" + summary)
	}
	if link != "" {
		if ru {
			b.WriteString("\n\nРезультат: " + link)
		} else {
			b.WriteString("\n\nResult: " + link)
		}
	}

	return b.String()
}

// handleTaskReconnectWS handles WebSocket reconnection to an existing task
func handleTaskReconnectWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	_, data, err := conn.ReadMessage()
	if err != nil {
		return
	}

	var req struct {
		TaskID string `json:"task_id"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		conn.WriteJSON(gin.H{"type": "error", "message": "Invalid JSON"})
		return
	}

	if req.TaskID == "" {
		conn.WriteJSON(gin.H{"type": "error", "message": "task_id is required"})
		return
	}

	log.Printf("🔄 Reconnecting to task %s", req.TaskID)

	if redisClient != nil && redisClient.IsEnabled() {
		ctx := context.Background()

		state, err := redisClient.GetStreamState(ctx, req.TaskID)
		if err != nil {
			conn.WriteJSON(gin.H{"type": "error", "message": "Failed to restore task state: " + err.Error()})
			return
		}

		if state == nil {
			conn.WriteJSON(gin.H{"type": "error", "message": "Task not found or expired"})
			return
		}

		// Send current state
		conn.WriteJSON(gin.H{
			"type":      "reconnected",
			"task_id":   state.TaskID,
			"progress":  state.Progress,
			"message":   state.Message,
			"status":    state.Status,
			"timestamp": time.Now().Unix(),
		})

		// Send historical updates
		updates, err := redisClient.GetStreamUpdates(ctx, req.TaskID)
		if err == nil {
			for _, update := range updates {
				wsUpdate := gin.H{
					"type":       update.Status,
					"task_id":    update.TaskID,
					"message":    update.Message,
					"progress":   update.Progress,
					"timestamp":  update.Timestamp,
					"is_history": true, // Mark as historical to avoid reprocessing
				}
				if update.Data != nil {
					wsUpdate["data"] = update.Data
				}
				conn.WriteJSON(wsUpdate)
			}
			log.Printf("📜 Sent %d historical updates", len(updates))
		}

		// If task is still running, subscribe to live updates via PubSub
		if pubSubManager != nil && (state.Status == "processing" || state.Status == "boss_planning" || state.Status == "managers_assigned") {
			subCh, err := pubSubManager.Subscribe(ctx, req.TaskID)
			if err == nil && subCh != nil {
				go func() {
					for update := range subCh {
						wsUpdate := gin.H{
							"type":      update.Status,
							"task_id":   update.TaskID,
							"message":   update.Message,
							"progress":  update.Progress,
							"timestamp": update.Timestamp,
						}
						if update.Data != nil {
							wsUpdate["data"] = update.Data
						}
						conn.WriteJSON(wsUpdate)
						if update.Status == "success" || update.Status == "error" {
							return
						}
					}
				}()
				log.Printf("✅ Subscribed to live updates for task %s", req.TaskID)
			}
		}

		if state.Status == "success" || state.Status == "error" {
			return
		}

		log.Printf("✅ Reconnected to task %s (progress: %d%%, status: %s)", req.TaskID, state.Progress, state.Status)
		return
	}

	// Fallback to Boss service - use ResumeTaskStream
	if bossClient == nil {
		conn.WriteJSON(gin.H{"type": "error", "message": "Task state not available"})
		return
	}

	log.Printf("🔄 Using Boss service ResumeTaskStream for task %s", req.TaskID)

	grpcReq := &bosspb.ResumeTaskStreamRequest{
		TaskId: req.TaskID,
		UserId: req.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	stream, err := bossClient.ResumeTaskStream(ctx, grpcReq)
	if err != nil {
		log.Printf("❌ Error calling ResumeTaskStream: %v", err)
		conn.WriteJSON(gin.H{
			"type":    "error",
			"message": "Failed to resume task: " + err.Error(),
		})
		return
	}

	// Stream updates from Boss service
	for {
		update, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("❌ Stream error: %v", err)
			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Connection error: " + err.Error(),
			})
			return
		}

		wsUpdate := gin.H{
			"type":      update.Status,
			"task_id":   update.TaskId,
			"message":   update.Message,
			"progress":  update.Progress,
			"timestamp": update.Timestamp,
		}
		if update.Data != nil {
			wsUpdate["data"] = update.Data
		}

		wsWriteJSONWithRedis(conn, req.TaskID, wsUpdate)

		if update.Status == "success" || update.Status == "error" {
			wsHub.Broadcast(req.TaskID, wsUpdate)
			return
		}
	}
}

// handleTaskStatus returns task status via HTTP
func handleTaskStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "task_id is required",
		})
		return
	}

	// Try Redis first
	if redisClient != nil && redisClient.IsEnabled() {
		state, err := redisClient.GetStreamState(context.Background(), taskID)
		if err != nil {
			log.Printf("Warning: failed to get stream state from Redis: %v", err)
		} else if state != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":   "success",
				"task_id":  state.TaskID,
				"progress": state.Progress,
				"message":  state.Message,
				"source":   "redis",
			})
			return
		}
	}

	// Fallback to Boss
	if bossClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "boss service unavailable",
		})
		return
	}

	resp, err := bossClient.GetTaskStatus(context.Background(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"task_id":  resp.TaskId,
		"progress": resp.Progress,
	})
}

// handleTaskStop stops a running task
func handleTaskStop(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "task_id is required",
		})
		return
	}

	if bossClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "boss service unavailable",
		})
		return
	}

	resp, err := bossClient.StopTask(context.Background(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"task_id": resp.TaskId,
		"message": "Task stopped successfully",
	})
}

// wsWriteJSON writes JSON to WebSocket connection
func wsWriteJSON(conn *websocket.Conn, taskID string, data gin.H) {
	if err := conn.WriteJSON(data); err != nil {
		log.Printf("❌ Failed to write to WebSocket: %v", err)
	}
}

// wsWriteJSONWithRedis writes to WebSocket and stores update in Redis
func wsWriteJSONWithRedis(conn *websocket.Conn, taskID string, update gin.H) {
	if err := conn.WriteJSON(update); err != nil {
		log.Printf("❌ Failed to write to WebSocket: %v", err)
		return
	}

	if redisClient != nil && redisClient.IsEnabled() {
		ctx := context.Background()

		status, _ := update["type"].(string)
		message, _ := update["message"].(string)
		progress := int32(0)
		if p, ok := update["progress"].(int32); ok {
			progress = p
		} else if p, ok := update["progress"].(int); ok {
			progress = int32(p)
		}

		state := redis.StreamState{
			TaskID:   taskID,
			UserID:   "",
			Status:   status,
			Progress: progress,
			Message:  message,
		}

		if err := redisClient.UpdateStreamState(ctx, taskID, state); err != nil {
			log.Printf("❌ Failed to update Redis stream state: %v", err)
		}

		streamUpdate := redis.StreamUpdate{
			TaskID:    taskID,
			Status:    status,
			Progress:  progress,
			Message:   message,
			Data:      update["data"],
			Timestamp: time.Now().Unix(),
		}

		if err := redisClient.AddStreamUpdate(ctx, taskID, streamUpdate); err != nil {
			log.Printf("❌ Failed to add Redis stream update: %v", err)
		}

		if pubSubManager != nil {
			if err := pubSubManager.Publish(ctx, taskID, streamUpdate); err != nil {
				log.Printf("❌ Failed to publish to Redis PubSub: %v", err)
			}
		}
	}
}
