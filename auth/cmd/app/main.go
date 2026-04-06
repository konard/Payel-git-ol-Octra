package main

import (
	"auth/internal/core/services"
	"auth/pkg/database"
	"auth/pkg/requests"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDb()
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// POST /register - Register new user
	r.POST("/register", func(c *gin.Context) {
		var req requests.UserRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Password == "" {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Username, email and password are required",
			})
			return
		}

		result, err := services.RegisterUser(req)
		if err != nil {
			status := 500
			if err.Error() == "UNIQUE constraint failed" || err.Error() == "duplicate key value violates unique constraint" {
				status = 409
			}
			c.JSON(status, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   result,
		})
	})

	// POST /login - Login user
	r.POST("/login", func(c *gin.Context) {
		var req requests.UserLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		result, err := services.LoginUser(req)
		if err != nil {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   result,
		})
	})

	// POST /refresh - Refresh access token
	r.POST("/refresh", func(c *gin.Context) {
		var req requests.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		// Try to get refresh token from cookie if not in body
		if req.RefreshToken == "" {
			token, err := c.Cookie("refresh_token")
			if err == nil {
				req.RefreshToken = token
			}
		}

		if req.RefreshToken == "" {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Refresh token is required",
			})
			return
		}

		result, err := services.RefreshTokens(req)
		if err != nil {
			c.JSON(401, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.SetSameSite(2)
		c.SetCookie("refresh_token", result["refresh_token"].(string), 604800, "/", "", false, true)

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   result,
		})
	})

	// GET /me - Get current user info
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

		// Remove "Bearer " prefix if present
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

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   user,
		})
	})

	// POST /logout - Logout user
	r.POST("/logout", func(c *gin.Context) {
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.SetCookie("access_token", "", -1, "/", "", false, true)

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Logged out successfully",
		})
	})

	// Get port from env or default
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "3112"
	}

	println("Auth service starting on port :" + port)
	r.Run(":" + port)
}
