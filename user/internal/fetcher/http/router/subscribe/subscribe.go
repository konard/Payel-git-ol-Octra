package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user/internal/core/services"
	"user/internal/fetcher/http/oauth/middleware"
	"user/pkg/requests"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/plans", func(c *gin.Context) {
		plans := services.GetSubscriptionPlans()
		c.JSON(200, gin.H{"status": "success", "data": plans})
	})

	r.POST("/subscribe", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req requests.SubscribeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		if req.Plan == "" {
			c.JSON(400, gin.H{"status": "error", "error": "Plan is required"})
			return
		}

		err := services.SubscribeUser(userID, req.Plan)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	})

	registerPromoRoutes(r)
}