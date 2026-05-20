package accaunt

import (
	"github.com/gin-gonic/gin"

	"user/internal/core/services"
	"user/pkg/requests"
)

func registerRefresh(r *gin.Engine) {
	r.POST("/refresh", func(c *gin.Context) {
		var req requests.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body: " + err.Error()})
			return
		}

		if req.RefreshToken == "" {
			token, err := c.Cookie("refresh_token")
			if err == nil {
				req.RefreshToken = token
			}
		}

		if req.RefreshToken == "" {
			c.JSON(400, gin.H{"status": "error", "error": "Refresh token is required"})
			return
		}

		result, err := services.RefreshTokens(req)
		if err != nil {
			c.JSON(401, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{"status": "ok", "data": result})
	})
}
