package accaunt

import (
	"github.com/gin-gonic/gin"

	"user/internal/core/services"
	"user/pkg/requests"
)

func registerRegister(r *gin.Engine) {
	r.POST("/register", func(c *gin.Context) {
		var req requests.UserRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid request body: " + err.Error()})
			return
		}
		if req.Username == "" || req.Email == "" || req.Password == "" {
			c.JSON(400, gin.H{"status": "error", "error": "Username, email and password are required"})
			return
		}
		if len(req.Password) < 6 {
			c.JSON(400, gin.H{"status": "error", "error": "Password must be at least 6 characters long"})
			return
		}

		result, err := services.RegisterUser(req)
		if err != nil {
			status := 500
			if err.Error() == "UNIQUE constraint failed" || err.Error() == "duplicate key value violates unique constraint" {
				status = 409
			}
			c.JSON(status, gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{"status": "ok", "data": result})
	})
}