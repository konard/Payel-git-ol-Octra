package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"user/internal/core/services"
)

// AuthMiddleware extracts user info from JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  "Authorization token is required",
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		t, err := services.ValidateAccessToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		claims := t.Claims.(jwt.MapClaims)
		userIDStr, _ := claims["user_id"].(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)

		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()
	}
}
