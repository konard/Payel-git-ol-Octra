package app

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user/internal/core/services"
	"user/internal/fetcher/http/oauth/middleware"
	"user/pkg/database"
	"user/pkg/requests"
)

func registerCustomProviders(r *gin.Engine) {
	customProviderService := services.NewCustomProviderService(database.Db)

	r.GET("/custom-providers", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		providers, err := customProviderService.GetUserCustomProviders(userID)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": providers})
	})

	r.POST("/custom-providers", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req requests.CreateCustomProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		provider, err := customProviderService.CreateCustomProvider(userID, req)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(201, gin.H{"status": "success", "data": provider})
	})

	r.PUT("/custom-providers/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		providerID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid provider ID"})
			return
		}

		var req requests.UpdateCustomProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		provider, err := customProviderService.UpdateCustomProvider(userID, providerID, req)
		if err != nil {
			status := 500
			if err.Error() == "custom provider not found" {
				status = 404
			}
			c.JSON(status, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": provider})
	})

	r.DELETE("/custom-providers/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		providerID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid provider ID"})
			return
		}

		err = customProviderService.DeleteCustomProvider(userID, providerID)
		if err != nil {
			status := 500
			if err.Error() == "custom provider not found" {
				status = 404
			}
			c.JSON(status, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "message": "Provider deleted"})
	})
}