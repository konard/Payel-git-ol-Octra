package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user/internal/core/services"
	"user/internal/fetcher/http/oauth/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	// GET /chat/history
	r.GET("/chat/history", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		chats, err := services.GetChatHistory(userID)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": chats})
	})

	// POST /chat/create
	r.POST("/chat/create", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req struct {
			Title string `json:"title"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			req.Title = "New Chat"
		}
		if req.Title == "" {
			req.Title = "New Chat"
		}

		chat, err := services.CreateChat(userID, req.Title)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(201, gin.H{"status": "success", "data": chat})
	})

	// GET /chat/:id
	r.GET("/chat/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		chatID := c.Param("id")

		chat, err := services.GetChat(userID, chatID)
		if err != nil {
			c.JSON(404, gin.H{"status": "error", "error": "Chat not found"})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": chat})
	})

	// POST /chat/:id/messages
	r.POST("/chat/:id/messages", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		chatID := c.Param("id")

		var req struct {
			Role     string `json:"role"`
			Content  string `json:"content"`
			Provider string `json:"provider"`
			Model    string `json:"model"`
			ApiKeyId string `json:"api_key_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		msg, err := services.AddChatMessage(userID, chatID, req.Role, req.Content, req.Provider, req.Model, req.ApiKeyId)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(201, gin.H{"status": "success", "data": msg})
	})

	// PUT /chat/:id/title
	r.PUT("/chat/:id/title", middleware.AuthMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")

		var req struct {
			Title string `json:"title"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		err := services.UpdateChatTitle(chatID, req.Title)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success"})
	})

	// PUT /chat/:id/workflow
	r.PUT("/chat/:id/workflow", middleware.AuthMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")

		var req struct {
			Workflow string `json:"workflow"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		err := services.UpdateChatWorkflow(chatID, req.Workflow)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success"})
	})

	// DELETE /chat/:id
	r.DELETE("/chat/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")

		err := services.DeleteChat(chatID)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success"})
	})
}