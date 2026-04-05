package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

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
var wsConnsMu sync.Mutex
var wsConns = make(map[*azurewebsockets.Conn]bool)

func init() {
	var err error
	bossHost := "boss:50051"
	bossClient, err = boss.NewClient(bossHost)
	if err != nil {
		log.Printf("Warning: failed to connect to boss service: %v", err)
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

	a.Get("/task/status", func(c *azure.Context) {
		taskID := c.GetQueryParam("task_id")
		if taskID == "" {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "task_id is required",
			})
			return
		}

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
	ws.WriteJSON(azure.M{
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

		ws.WriteJSON(wsUpdate)

		// Check if finished
		if update.Status == "success" || update.Status == "error" {
			// Broadcast to other clients if needed
			wsHub.Broadcast(taskReq.UserID, wsUpdate)
			return
		}
	}
}
