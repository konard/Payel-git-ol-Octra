package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"user/internal/cache"
	"user/internal/core/services"
	orchgrpc "user/internal/grpc"
	"user/internal/fetcher/http/router"
	"user/pkg/database"
	"user/pkg/models"

	"github.com/gin-gonic/gin"
)

// orchClient и nixCache доступны HTTP-обработчикам через пакетные переменные.
var (
	orchClient *orchgrpc.OrchestratorClient
	nixCache   *cache.NixCache
)

func main() {
	database.InitDb()
	database.Db.AutoMigrate(&models.UserRegister{}, &models.Subscription{}, &models.PromoCode{}, &models.Workflow{}, &models.CustomProvider{}, &models.CustomModel{}, &models.Chat{}, &models.ChatMessage{})

	services.InitDefaultPromoCodes()

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost",
			"http://127.0.0.1:5173",
			"http://127.0.0.1",
		}
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		if c.GetHeader("Access-Control-Allow-Origin") == "" {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize orchestrator gRPC client
	orchAddr := os.Getenv("ORCHESTRATOR_GRPC_ADDR")
	if orchAddr == "" {
		orchAddr = "localhost:50051"
	}
	var err error
	orchClient, err = orchgrpc.NewOrchestratorClient(orchAddr)
	if err != nil {
		log.Printf("Warning: failed to connect to orchestrator: %v", err)
	} else {
		defer orchClient.Close()
		nixCache = cache.NewNixCache(500*1024*1024, 10)
	}

	// HTTP endpoints (including /chat/:id/files for frontend)
	router.RegisterAllRoutes(r)
	registerFileRoutes(r)

	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "3112"
	}

	log.Printf("User service starting on port %s", port)
	r.Run(":" + port)
}

// registerFileRoutes — HTTP endpoint для получения файлов проекта из Nix store.
// Фронтенд вызывает этот endpoint вместо gRPC (браузер не поддерживает
// стандартный gRPC без Envoy proxy).
func registerFileRoutes(r *gin.Engine) {
	r.GET("/chat/:id/files", func(c *gin.Context) {
		chatID := c.Param("id")

		if orchClient == nil || nixCache == nil {
			c.JSON(503, gin.H{"status": "error", "error": "Orchestrator not connected"})
			return
		}

		// 1. Get chat from DB
		chat, err := services.GetChatByID(chatID)
		if err != nil {
			c.JSON(404, gin.H{"status": "error", "error": "Chat not found"})
			return
		}

		if chat.NixStorePath == "" {
			c.JSON(200, gin.H{
				"status": "success",
				"data": gin.H{
					"chat_id":     chatID,
					"task_id":     "",
					"files":       []interface{}{},
					"total_files": 0,
				},
			})
			return
		}

		taskID := chat.TaskId

		// 2. Check cache
		if cached, ok := nixCache.Get(taskID); ok {
			c.JSON(200, gin.H{
				"status": "success",
				"data": gin.H{
					"chat_id":     chatID,
					"task_id":     taskID,
					"files":       cached,
					"total_files": len(cached),
				},
			})
			return
		}

		// 3. Fetch from orchestrator via gRPC
		resp, err := orchClient.RestoreProjectFiles(context.Background(), chat.NixStorePath, taskID)
		if err != nil {
			log.Printf("Failed to restore files from nix store: %v", err)
			c.JSON(500, gin.H{"status": "error", "error": fmt.Sprintf("Failed to restore files: %v", err)})
			return
		}

		// 4. Cache result
		cacheFiles := make([]cache.CodeFileEntry, 0, len(resp.Files))
		type fileEntry struct {
			Path     string `json:"path"`
			Content  string `json:"content"`
			Language string `json:"language"`
			Encoding string `json:"encoding,omitempty"`
		}
		entries := make([]fileEntry, 0, len(resp.Files))
		for _, f := range resp.Files {
			cacheFiles = append(cacheFiles, cache.CodeFileEntry{
				Path:     f.Path,
				Content:  f.Content,
				Language: f.Language,
				Encoding: f.Encoding,
			})
			entries = append(entries, fileEntry{
				Path:     f.Path,
				Content:  f.Content,
				Language: f.Language,
				Encoding: f.Encoding,
			})
		}
		nixCache.Set(taskID, cacheFiles)

		// 5. Return
		c.JSON(200, gin.H{
			"status": "success",
			"data": gin.H{
				"chat_id":     chatID,
				"task_id":     taskID,
				"files":       entries,
				"total_files": len(entries),
			},
		})
	})
}
