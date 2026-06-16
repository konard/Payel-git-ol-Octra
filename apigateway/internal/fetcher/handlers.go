package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"apigateway/internal/fetcher/grpc/boss/bosspb"
	"apigateway/pkg/requests"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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
					"is_history": true,
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

		su := buildStreamUpdate(req.TaskID, update.Status, update.Message, update.Progress, update.Timestamp, update.Data, "")

		data, err := json.Marshal(su)
		if err != nil {
			log.Printf("❌ Failed to marshal update: %v", err)
			conn.WriteJSON(gin.H{"type": "error", "message": "Failed to marshal update"})
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("❌ Failed to write to WebSocket: %v", err)
			return
		}

		if redisClient != nil && redisClient.IsEnabled() {
			if err := redisClient.StoreStreamUpdate(context.Background(), req.TaskID, data, su.Type, su.Progress, su.Message); err != nil {
				log.Printf("❌ Redis StoreStreamUpdate error: %v", err)
			}
		}

		if update.Status == "success" || update.Status == "error" {
			wsHub.Broadcast(req.TaskID, su)
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

// wsWriteJSON writes JSON to WebSocket connection
func wsWriteJSON(conn *websocket.Conn, taskID string, data gin.H) {
	if err := conn.WriteJSON(data); err != nil {
		log.Printf("❌ Failed to write to WebSocket: %v", err)
	}
}

// wsWriteJSONWithRedis writes to WebSocket and stores update in Redis using
// marshal-once + pipeline. It converts gin.H to streamUpdateData, marshals
// once, writes raw JSON to WebSocket, then pipelines all Redis ops.
func wsWriteJSONWithRedis(conn *websocket.Conn, taskID string, update gin.H) {
	su := streamUpdateData{
		Type:    toString(update["type"]),
		TaskID:  taskID,
		Message: toString(update["message"]),
	}
	if p, ok := update["progress"].(int32); ok {
		su.Progress = p
	} else if p, ok := update["progress"].(int); ok {
		su.Progress = int32(p)
	}
	if ts, ok := update["timestamp"].(int64); ok {
		su.Timestamp = ts
	} else {
		su.Timestamp = time.Now().Unix()
	}
	if d, ok := update["data"]; ok {
		su.Data = d
	}
	if s, ok := update["sender"].(string); ok {
		su.Sender = s
	}

	data, err := json.Marshal(su)
	if err != nil {
		log.Printf("❌ Failed to marshal update: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("❌ Failed to write to WebSocket: %v", err)
		return
	}

	if redisClient == nil || !redisClient.IsEnabled() {
		return
	}

	if err := redisClient.StoreStreamUpdate(context.Background(), taskID, data, su.Type, su.Progress, su.Message); err != nil {
		log.Printf("❌ Redis StoreStreamUpdate error: %v", err)
	}
}
