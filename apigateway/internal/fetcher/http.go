package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"apigateway/internal/core/redis"
	"apigateway/internal/core/services"
	"apigateway/internal/fetcher/grpc/boss"
	"apigateway/internal/fetcher/grpc/boss/bosspb"
	"apigateway/pkg/requests"

	"github.com/Payel-git-ol/azure"
	"github.com/Payel-git-ol/azure/azurewebsockets"
	"github.com/google/uuid"
)

var bossClient *boss.Client
var wsHub *services.Hub
var redisClient *redis.Client
var pubSubManager *redis.PubSubManager
var wsConnsMu sync.Mutex
var wsConns = make(map[*azurewebsockets.Conn]bool)

func init() {
	var err error
	bossHost := "boss:50051"
	bossClient, err = boss.NewClient(bossHost)
	if err != nil {
		log.Printf("Warning: failed to connect to boss service: %v", err)
	}

	// Initialize Redis client
	redisClient = redis.NewClient()
	if redisClient.IsEnabled() {
		pubSubManager = redis.NewPubSubManager(redisClient)
		log.Println("[Redis] PubSub manager initialized")
	}

	// Initialize WebSocket hub
	wsHub = services.NewHub()
	go wsHub.Run()
}

func HttpManager(a *azure.Azure) {
	a.Use(azure.Recovery())
	a.Use(azure.Logger())

	a.Get("/health", func(c *azure.Context) {
		c.JsonStatus(200, azure.M{
			"status": "OK",
		})
	})

	// WebSocket endpoint for task creation
	a.Get("/task/create", azurewebsockets.HandlerFunc(handleTaskCreateWebSocket))

	// WebSocket endpoint for reconnect to existing task
	a.Get("/task/reconnect", azurewebsockets.HandlerFunc(handleTaskReconnectWebSocket))

	a.Get("/task/status", func(c *azure.Context) {
		taskID := c.GetQueryParam("task_id")
		if taskID == "" {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "task_id is required",
			})
			return
		}

		// Try to get status from Redis first
		if redisClient != nil && redisClient.IsEnabled() {
			state, err := redisClient.GetStreamState(context.Background(), taskID)
			if err != nil {
				log.Printf("Warning: failed to get stream state from Redis: %v", err)
			} else if state != nil {
				c.JsonStatus(200, azure.M{
					"status":   "success",
					"task_id":  state.TaskID,
					"progress": state.Progress,
					"message":  state.Message,
					"source":   "redis",
				})
				return
			}
		}

		// Fallback to Boss service
		if bossClient == nil {
			c.JsonStatus(503, azure.M{
				"status": "error",
				"error":  "boss service unavailable",
			})
			return
		}

		resp, err := bossClient.GetTaskStatus(context.Background(), taskID)
		if err != nil {
			c.JsonStatus(500, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JsonStatus(200, azure.M{
			"status":   "success",
			"task_id":  resp.TaskId,
			"progress": resp.Progress,
		})
	})
}

// handleTaskCreateWebSocket handles WebSocket connections for task creation
func handleTaskCreateWebSocket(ws *azurewebsockets.Conn, opcode int, data []byte) {
	switch opcode {
	case azurewebsockets.OpText:
		// Ignore empty or whitespace-only frames (initial connection)
		if len(data) == 0 || len(bytes.TrimSpace(data)) == 0 {
			return
		}

		// Parse initial task request
		var taskReq requests.CreateTaskRequest
		if err := json.Unmarshal(data, &taskReq); err != nil {
			log.Printf("❌ Failed to parse task request: %v", err)
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "Invalid JSON: " + err.Error(),
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Validate request
		if taskReq.Username == "" || taskReq.Title == "" {
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "Missing required fields: username, title",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Generate task ID if not provided
		if taskReq.UserID == "" {
			taskReq.UserID = uuid.New().String()
		}

		// Send initial connection confirmation
		ws.WriteJSON(azure.M{
			"type":      "connected",
			"task_id":   taskReq.UserID,
			"message":   "Connected to task creation service",
			"timestamp": time.Now().Unix(),
		})

		// Register this connection
		wsConnsMu.Lock()
		wsConns[ws] = true
		wsConnsMu.Unlock()

		// Process task stream in background
		go processTaskStreamAzureWS(ws, taskReq)

	case azurewebsockets.OpClose:
		log.Printf("WebSocket closed")
		wsConnsMu.Lock()
		delete(wsConns, ws)
		wsConnsMu.Unlock()
		ws.Close()

	case azurewebsockets.OpPing:
		// Respond with Pong
		ws.WriteMessage(azurewebsockets.OpPong, nil)
	}
}

// processTaskStreamAzureWS sends task to Boss and streams updates back via WebSocket
func processTaskStreamAzureWS(ws *azurewebsockets.Conn, taskReq requests.CreateTaskRequest) {
	// Create gRPC request
	grpcReq := &bosspb.CreateTaskRequest{
		UserId:      taskReq.UserID,
		Username:    taskReq.Username,
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Tokens:      taskReq.Tokens,
		Meta:        taskReq.Meta,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Send initial progress
	wsWriteJSONWithRedis(ws, taskReq.UserID, azure.M{
		"type":      "progress",
		"task_id":   taskReq.UserID,
		"message":   "Connecting to Boss service...",
		"progress":  5,
		"timestamp": time.Now().Unix(),
	})

	// Call Boss streaming RPC
	stream, err := bossClient.CreateTaskStream(ctx, grpcReq)
	if err != nil {
		log.Printf("❌ Error calling CreateTaskStream: %v", err)
		ws.WriteJSON(azure.M{
			"type":      "error",
			"message":   "Failed to connect to Boss service: " + err.Error(),
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Read messages from stream and send to WebSocket client
	for {
		update, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("❌ Stream error: %v", err)
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "Connection error: " + err.Error(),
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Convert gRPC update to WebSocket format
		wsUpdate := azure.M{
			"type":      update.Status,
			"task_id":   update.TaskId,
			"message":   update.Message,
			"progress":  update.Progress,
			"timestamp": update.Timestamp,
		}
		if update.Data != nil {
			wsUpdate["data"] = update.Data
		}

		// Send to WebSocket and store in Redis
		wsWriteJSONWithRedis(ws, taskReq.UserID, wsUpdate)

		// Check if finished
		if update.Status == "success" || update.Status == "error" {
			// Broadcast to other clients if needed
			wsHub.Broadcast(taskReq.UserID, wsUpdate)
			return
		}
	}
}

// wsWriteJSONWithRedis writes to WebSocket and stores update in Redis
func wsWriteJSONWithRedis(ws *azurewebsockets.Conn, taskID string, update azure.M) {
	// Write to WebSocket
	if err := ws.WriteJSON(update); err != nil {
		log.Printf("❌ Failed to write to WebSocket: %v", err)
		return
	}

	// Store in Redis if available
	if redisClient != nil && redisClient.IsEnabled() {
		ctx := context.Background()

		// Update stream state
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
			UserID:   "", // Will be set on first call
			Status:   status,
			Progress: progress,
			Message:  message,
		}

		if err := redisClient.UpdateStreamState(ctx, taskID, state); err != nil {
			log.Printf("❌ Failed to update Redis stream state: %v", err)
		}

		// Add to updates list
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

		// Publish to PubSub for other Apigateway instances
		if pubSubManager != nil {
			if err := pubSubManager.Publish(ctx, taskID, streamUpdate); err != nil {
				log.Printf("❌ Failed to publish to Redis PubSub: %v", err)
			}
		}
	}
}

// handleTaskReconnectWebSocket handles WebSocket reconnection to an existing task
func handleTaskReconnectWebSocket(ws *azurewebsockets.Conn, opcode int, data []byte) {
	switch opcode {
	case azurewebsockets.OpText:
		// Parse reconnect request
		var req struct {
			TaskID string `json:"task_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			log.Printf("❌ Failed to parse reconnect request: %v", err)
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "Invalid JSON: " + err.Error(),
				"timestamp": time.Now().Unix(),
			})
			return
		}

		if req.TaskID == "" {
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "task_id is required",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		log.Printf("🔄 Reconnecting to task %s", req.TaskID)

		// Try to restore from Redis
		if redisClient != nil && redisClient.IsEnabled() {
			ctx := context.Background()

			// Get current state
			state, err := redisClient.GetStreamState(ctx, req.TaskID)
			if err != nil {
				log.Printf("❌ Failed to get stream state: %v", err)
				ws.WriteJSON(azure.M{
					"type":      "error",
					"message":   "Failed to restore task state: " + err.Error(),
					"timestamp": time.Now().Unix(),
				})
				return
			}

			if state == nil {
				ws.WriteJSON(azure.M{
					"type":      "error",
					"message":   "Task not found or expired",
					"timestamp": time.Now().Unix(),
				})
				return
			}

			// Send current state
			ws.WriteJSON(azure.M{
				"type":      "reconnected",
				"task_id":   state.TaskID,
				"progress":  state.Progress,
				"message":   state.Message,
				"status":    state.Status,
				"timestamp": time.Now().Unix(),
			})

			// Send recent updates
			updates, err := redisClient.GetStreamUpdates(ctx, req.TaskID)
			if err != nil {
				log.Printf("❌ Failed to get stream updates: %v", err)
			} else {
				for _, update := range updates {
					wsUpdate := azure.M{
						"type":      update.Status,
						"task_id":   update.TaskID,
						"message":   update.Message,
						"progress":  update.Progress,
						"timestamp": update.Timestamp,
					}
					if update.Data != nil {
						wsUpdate["data"] = update.Data
					}
					ws.WriteJSON(wsUpdate)
				}
				log.Printf("📜 Sent %d historical updates", len(updates))
			}

			// Subscribe to future updates via PubSub
			if pubSubManager != nil {
				subCh, err := pubSubManager.Subscribe(ctx, req.TaskID)
				if err != nil {
					log.Printf("❌ Failed to subscribe to task updates: %v", err)
				} else if subCh != nil {
					// Forward PubSub updates to WebSocket
					go func() {
						for update := range subCh {
							wsUpdate := azure.M{
								"type":      update.Status,
								"task_id":   update.TaskID,
								"message":   update.Message,
								"progress":  update.Progress,
								"timestamp": update.Timestamp,
							}
							if update.Data != nil {
								wsUpdate["data"] = update.Data
							}
							ws.WriteJSON(wsUpdate)

							// Stop if task completed
							if update.Status == "success" || update.Status == "error" {
								return
							}
						}
					}()
				}
			}

			// If task already completed, close connection
			if state.Status == "success" || state.Status == "error" {
				log.Printf("✅ Task %s already completed with status: %s", req.TaskID, state.Status)
				return
			}

			log.Printf("✅ Reconnected to task %s (progress: %d%%)", req.TaskID, state.Progress)
			return
		}

		// Redis not available - fallback to Boss service
		if bossClient == nil {
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "Task state not available (Redis disabled)",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		// Query Boss for task status
		resp, err := bossClient.GetTaskStatus(context.Background(), req.TaskID)
		if err != nil {
			ws.WriteJSON(azure.M{
				"type":      "error",
				"message":   "Failed to get task status: " + err.Error(),
				"timestamp": time.Now().Unix(),
			})
			return
		}

		ws.WriteJSON(azure.M{
			"type":      "reconnected",
			"task_id":   resp.TaskId,
			"progress":  resp.Progress,
			"message":   "Restored from Boss service",
			"timestamp": time.Now().Unix(),
		})

	case azurewebsockets.OpClose:
		log.Printf("WebSocket closed (reconnect)")
		ws.Close()

	case azurewebsockets.OpPing:
		ws.WriteMessage(azurewebsockets.OpPong, nil)
	}
}
