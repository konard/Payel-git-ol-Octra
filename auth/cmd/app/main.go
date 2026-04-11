package main

import (
	"auth/internal/core/services"
	"auth/pkg/database"
	"auth/pkg/models"
	"auth/pkg/requests"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func main() {
	database.InitDb()
	database.Db.AutoMigrate(&models.UserRegister{}, &models.Subscription{}, &models.PromoCode{}, &models.Workflow{})
	services.InitDefaultPromoCodes()
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost",
			"http://127.0.0.1:5173",
			"http://127.0.0.1",
		}

		// Проверяем, есть ли origin в списке разрешённых
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Если origin не указан или не разрешён — ставим localhost по умолчанию для dev
		if c.GetHeader("Access-Control-Allow-Origin") == "" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

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

	// GET /plans - Get subscription plans
	r.GET("/plans", func(c *gin.Context) {
		plans := services.GetSubscriptionPlans()
		c.JSON(200, gin.H{
			"status": "ok",
			"data":   plans,
		})
	})

	// POST /subscribe - Subscribe user to a plan
	r.POST("/subscribe", func(c *gin.Context) {
		var req requests.SubscribeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		if req.UserID == "" || req.Plan == "" {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "User ID and plan are required",
			})
			return
		}

		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid user ID format",
			})
			return
		}

		err = services.SubscribeUser(userID, req.Plan)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Subscription activated successfully",
		})
	})

	// POST /subscribe/promo - Activate promo code
	r.POST("/subscribe/promo", func(c *gin.Context) {
		var req requests.PromoCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		if req.UserID == "" || req.Code == "" {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "User ID and promo code are required",
			})
			return
		}

		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid user ID format",
			})
			return
		}

		err = services.ActivatePromoCode(userID, req.Code)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Promo code activated successfully",
		})
	})

	// GET /subscription/status - Get user subscription status
	r.GET("/subscription/status", func(c *gin.Context) {
		userIDParam := c.Query("user_id")
		if userIDParam == "" {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "user_id is required",
			})
			return
		}

		userID, err := uuid.Parse(userIDParam)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid user_id",
			})
			return
		}

		info, err := services.GetUserSubscriptionInfo(userID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   info,
		})
	})

	// ==================== WORKFLOW LIBRARY ENDPOINTS ====================

	// POST /workflows - Create a new workflow
	r.POST("/workflows", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		username := c.MustGet("username").(string)

		var req requests.CreateWorkflowRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		workflow, err := services.CreateWorkflow(userID, username, req)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"status": "ok",
			"data":   workflow,
		})
	})

	// GET /workflows/library - Get public workflows (with optional filters)
	r.GET("/workflows/library", func(c *gin.Context) {
		category := c.Query("category")
		tag := c.Query("tag")
		limit := 50 // default limit

		workflows, err := services.GetPublicWorkflows(category, tag, limit, 0)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   workflows,
		})
	})

	// GET /workflows/categories - Get all workflow categories
	r.GET("/workflows/categories", func(c *gin.Context) {
		categories, err := services.GetWorkflowCategories()
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   categories,
		})
	})

	// GET /workflows/my - Get current user's workflows
	r.GET("/workflows/my", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		workflows, err := services.GetUserWorkflows(userID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   workflows,
		})
	})

	// GET /workflows/:id - Get a specific workflow
	r.GET("/workflows/:id", func(c *gin.Context) {
		workflowID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid workflow ID",
			})
			return
		}

		// Try to get user ID from context if authenticated
		var userID uuid.UUID
		if val, exists := c.Get("userID"); exists {
			userID = val.(uuid.UUID)
		} else {
			// For public workflows, use a nil UUID
			userID = uuid.Nil
		}

		workflow, err := services.GetWorkflowByID(workflowID, userID)
		if err != nil {
			c.JSON(404, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   workflow,
		})
	})

	// POST /workflows/:id/download - Increment download counter
	r.POST("/workflows/:id/download", func(c *gin.Context) {
		workflowID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid workflow ID",
			})
			return
		}

		err = services.DownloadWorkflow(workflowID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Download registered",
		})
	})

	// PUT /workflows/:id - Update a workflow
	r.PUT("/workflows/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		workflowID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid workflow ID",
			})
			return
		}

		var req requests.UpdateWorkflowRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		workflow, err := services.UpdateWorkflow(workflowID, userID, req)
		if err != nil {
			c.JSON(404, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   workflow,
		})
	})

	// DELETE /workflows/:id - Delete a workflow
	r.DELETE("/workflows/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		workflowID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid workflow ID",
			})
			return
		}

		err = services.DeleteWorkflow(workflowID, userID)
		if err != nil {
			c.JSON(404, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Workflow deleted successfully",
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
