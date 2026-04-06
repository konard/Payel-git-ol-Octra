package main

import (
	"auth/internal/core/services"
	"auth/pkg/database"
	"auth/pkg/requests"
	"os"

	"github.com/Payel-git-ol/azure"
)

func main() {
	database.InitDb()
	a := azure.Defoult

	// Health check
	a.Get("/health", func(c *azure.Context) {
		c.Json(azure.M{"status": "ok"})
	})

	// POST /register - Register new user
	a.Post("/register", func(c *azure.Context) {
		var req requests.UserRegisterRequest
		if err := c.BindJSON(&req); err != nil {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Password == "" {
			c.JsonStatus(400, azure.M{
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
			c.JsonStatus(status, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.SetCookie("refresh_token", result["refresh_token"].(string))

		c.Json(azure.M{
			"status": "ok",
			"data":   result,
		})
	})

	// POST /login - Login user
	a.Post("/login", func(c *azure.Context) {
		var req requests.UserLoginRequest
		if err := c.BindJSON(&req); err != nil {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		result, err := services.LoginUser(req)
		if err != nil {
			c.JsonStatus(401, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.SetCookie("refresh_token", result["refresh_token"].(string))

		c.Json(azure.M{
			"status": "ok",
			"data":   result,
		})
	})

	// POST /refresh - Refresh access token
	a.Post("/refresh", func(c *azure.Context) {
		var req requests.RefreshTokenRequest
		if err := c.BindJSON(&req); err != nil {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		// Try to get refresh token from cookie if not in body
		if req.RefreshToken == "" {
			token, ok := c.GetCookie("refresh_token")
			if ok {
				req.RefreshToken = token
			}
		}

		if req.RefreshToken == "" {
			c.JsonStatus(400, azure.M{
				"status": "error",
				"error":  "Refresh token is required",
			})
			return
		}

		result, err := services.RefreshTokens(req)
		if err != nil {
			c.JsonStatus(401, azure.M{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.SetCookie("refresh_token", result["refresh_token"].(string))

		c.Json(azure.M{
			"status": "ok",
			"data":   result,
		})
	})

	// GET /me - Get current user info
	a.Get("/me", func(c *azure.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			tokenVal, ok := c.GetCookie("access_token")
			if ok {
				token = tokenVal
			}
		}

		if token == "" {
			c.JsonStatus(401, azure.M{
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
			c.JsonStatus(401, azure.M{
				"status": "error",
				"error":  "Invalid or expired token: " + err.Error(),
			})
			return
		}

		c.Json(azure.M{
			"status": "ok",
			"data":   user,
		})
	})

	// POST /logout - Logout user
	a.Post("/logout", func(c *azure.Context) {
		c.SetCookie("refresh_token", "")
		c.SetCookie("access_token", "")

		c.Json(azure.M{
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
	a.Run(":" + port)
}
