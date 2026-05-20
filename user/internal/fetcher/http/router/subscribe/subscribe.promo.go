package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user/internal/core/services"
	"user/internal/fetcher/http/oauth/middleware"
)

func registerPromoRoutes(r *gin.Engine) {
	r.POST("/subscribe/promo", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req struct {
			UserID string `json:"user_id"`
			Code   string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
			return
		}

		if req.UserID == "" || req.Code == "" {
			c.JSON(400, gin.H{"status": "error", "error": "User ID and promo code are required"})
			return
		}

		err := services.ActivatePromoCode(userID, req.Code)
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	})
}