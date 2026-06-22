package accaunt

import (
	"github.com/gin-gonic/gin"

	"user/internal/core/services"
	"user/pkg/requests"
)

func registerLogin(r *gin.Engine) {
	r.POST("/login", func(c *gin.Context) {
		var req requests.UserLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body: " + err.Error()})
			return
		}
		result, err := services.LoginUser(req)
		if err != nil {
			c.JSON(401, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{"status": "ok", "data": result})
	})
}
