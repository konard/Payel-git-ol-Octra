package fetcher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"apigateway/internal/fetcher/grpc/boss/bosspb"
	"apigateway/pkg/requests"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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
		sender := ""
		if update.IsChat {
			messageType = "chat"
			sender = "boss"
		}

		su := buildStreamUpdate(streamID, messageType, update.Message, update.Progress, update.Timestamp, update.Data, sender)

		data, err := json.Marshal(su)
		if err != nil {
			log.Printf("❌ Failed to marshal update: %v", err)
			conn.WriteJSON(gin.H{
				"type":    "error",
				"message": "Failed to marshal update",
			})
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("❌ Failed to write to WebSocket: %v", err)
			return
		}

		if redisClient != nil && redisClient.IsEnabled() {
			if err := redisClient.StoreStreamUpdate(context.Background(), streamID, data, su.Type, su.Progress, su.Message); err != nil {
				log.Printf("❌ Redis StoreStreamUpdate error: %v", err)
			}
		}

		if update.Status == "success" || update.Status == "error" {
			wsHub.Broadcast(streamID, su)
			return
		}
	}
}
