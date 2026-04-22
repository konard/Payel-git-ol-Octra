package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"apigateway/internal/core/ratelimit"
	"apigateway/internal/core/redis"
	"apigateway/internal/core/services"
	"apigateway/internal/fetcher/grpc/boss"
	"apigateway/internal/fetcher/grpc/boss/bosspb"
	"apigateway/pkg/models"
	"apigateway/pkg/requests"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
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
}

func init() {
	var err error
	bossHost := "boss:50051"
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

	// Provide default values for missing fields
	if taskReq.Username == "" {
		taskReq.Username = "Anonymous User"
	}
	if taskReq.Title == "" {
		taskReq.Title = "Untitled Task"
	}

	if taskReq.UserID == "" {
		taskReq.UserID = uuid.New().String()
	}

	// Send confirmation
	conn.WriteJSON(gin.H{
		"type":      "connected",
		"task_id":   taskReq.UserID,
		"message":   "Connected to task creation service",
		"timestamp": time.Now().Unix(),
	})

	// Process task stream in background
	go processTaskStreamWS(conn, taskReq)

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

// processTaskStreamWS sends task to Boss and streams updates back via WebSocket
func processTaskStreamWS(conn *websocket.Conn, taskReq requests.CreateTaskRequest) {
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
	wsWriteJSON(conn, taskReq.UserID, gin.H{
		"type":      "progress",
		"task_id":   taskReq.UserID,
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

		wsWriteJSONWithRedis(conn, taskReq.UserID, wsUpdate)

		if update.Status == "success" || update.Status == "error" {
			wsHub.Broadcast(taskReq.UserID, wsUpdate)
			return
		}
	}
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
