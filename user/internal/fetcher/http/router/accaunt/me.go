package accaunt

import (
	"github.com/gin-gonic/gin"

	"user/internal/core/services"
)

func registerMe(r *gin.Engine) {
	r.GET("/me", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			tokenVal, err := c.Cookie("access_token")
			if err == nil {
				token = tokenVal
			}
		}
		if token == "" {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  "Authorization token is required",
			})
			return
		}
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		user, err := services.GetMe(token)
		if err != nil {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  "Invalid or expired token: " + err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{"status": "ok", "data": user})
	})
}
