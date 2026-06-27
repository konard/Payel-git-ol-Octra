package accaunt

import (
	"errors"

	"github.com/gin-gonic/gin"

	"user/internal/core/services"
	"user/pkg/requests"
)

// statusForRegisterError maps a RegisterUser error to the appropriate HTTP
// status code: 409 Conflict when the account already exists, 500 otherwise.
func statusForRegisterError(err error) int {
	if errors.Is(err, services.ErrUserAlreadyExists) {
		return 409
	}
	return 500
}

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
			c.JSON(statusForRegisterError(err), gin.H{"status": "error", "error": err.Error()})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{"status": "ok", "data": result})
	})
}
