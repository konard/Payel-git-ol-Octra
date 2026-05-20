package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user/internal/core/services"
	"user/internal/fetcher/http/oauth/middleware"
	"user/pkg/database"
	"user/pkg/requests"
)

func registerCustomModels(r *gin.Engine) {
	customProviderService := services.NewCustomProviderService(database.Db)

	// GET /custom-models
	r.GET("/custom-models", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		models, err := customProviderService.GetUserCustomModels(userID)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": models})
	})

	// POST /custom-models
	r.POST("/custom-models", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req requests.CreateCustomModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body: " + err.Error()})
			return
		}

		fmt.Printf("CreateCustomModel request: %+v\n", req)

		model, err := customProviderService.CreateCustomModel(userID, req)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(201, gin.H{"status": "success", "data": model})
	})

	// PUT /custom-models/:id
	r.PUT("/custom-models/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		modelID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid model ID"})
			return
		}

		var req requests.UpdateCustomModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		model, err := customProviderService.UpdateCustomModel(userID, modelID, req)
		if err != nil {
			status := 500
			if err.Error() == "custom model not found" {
				status = 404
			}
			c.JSON(status, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": model})
	})

	// DELETE /custom-models/:id
	r.DELETE("/custom-models/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		modelID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid model ID"})
			return
		}

		err = customProviderService.DeleteCustomModel(userID, modelID)
		if err != nil {
			status := 500
			if err.Error() == "custom model not found" {
				status = 404
			}
			c.JSON(status, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "message": "Model deleted"})
	})
}