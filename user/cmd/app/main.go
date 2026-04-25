package main

import (
	"fmt"
	"os"
	"strconv"
	"user/internal/core/services"
	"user/pkg/database"
	"user/pkg/models"
	"user/pkg/requests"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
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
	database.Db.AutoMigrate(&models.UserRegister{}, &models.Subscription{}, &models.PromoCode{}, &models.Workflow{}, &models.CustomProvider{}, &models.CustomModel{}, &models.Chat{}, &models.ChatMessage{})
	services.InitRedis()
	services.InitDefaultPromoCodes()

	// Initialize services
	customProviderService := services.NewCustomProviderService(database.Db)

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
		c.JSON(200, gin.H{"status": "ok", "message": "Logged out successfully"})
	})

	// Custom Models routes
	// GET /custom-models - Get user's custom models
	r.GET("/custom-models", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		models, err := customProviderService.GetUserCustomModels(userID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"data":   models,
		})
	})

	// POST /custom-models - Create a new custom model
	r.POST("/custom-models", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req requests.CreateCustomModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		// Debug log
		fmt.Printf("CreateCustomModel request: %+v\n", req)

		model, err := customProviderService.CreateCustomModel(userID, req)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"status": "success",
			"data":   model,
		})
	})

	// PUT /custom-models/:id - Update a custom model
	r.PUT("/custom-models/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		modelID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid model ID",
			})
			return
		}

		var req requests.UpdateCustomModelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body",
			})
			return
		}

		model, err := customProviderService.UpdateCustomModel(userID, modelID, req)
		if err != nil {
			status := 500
			if err.Error() == "custom model not found" {
				status = 404
			}
			c.JSON(status, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"data":   model,
		})
	})

	// DELETE /custom-models/:id - Delete a custom model
	r.DELETE("/custom-models/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		modelID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid model ID",
			})
			return
		}

		err = customProviderService.DeleteCustomModel(userID, modelID)
		if err != nil {
			status := 500
			if err.Error() == "custom model not found" {
				status = 404
			}
			c.JSON(status, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
		})
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
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

	// ==================== CHAT HISTORY ENDPOINTS ====================

	// GET /chat/history - Get user chat history
	r.GET("/chat/history", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		chats, err := services.GetChatHistory(userID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   chats,
		})
	})

	// POST /chat/create - Create new chat
	r.POST("/chat/create", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req struct {
			Title string `json:"title"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			req.Title = "New Chat"
		}

		chat, err := services.CreateChat(userID, req.Title)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   chat,
		})
	})

	// GET /chat/:id - Get chat with messages
	r.GET("/chat/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		chatID := c.Param("id")

		chat, err := services.GetChat(userID, chatID)
		if err != nil {
			c.JSON(404, gin.H{
				"status": "error",
				"error":  "Chat not found",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   chat,
		})
	})

	// POST /chat/:id/messages - Add message to chat
	r.POST("/chat/:id/messages", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		chatID := c.Param("id")

		var req struct {
			Role     string `json:"role"`
			Content string `json:"content"`
			Provider string `json:"provider,omitempty"`
			Model    string `json:"model,omitempty"`
			ApiKeyId string `json:"api_key_id,omitempty"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body",
			})
			return
		}

		msg, err := services.AddChatMessage(userID, chatID, req.Role, req.Content, req.Provider, req.Model, req.ApiKeyId)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   msg,
		})
	})

	// PUT /chat/:id/title - Update chat title
	r.PUT("/chat/:id/title", AuthMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")

		var req struct {
			Title string `json:"title"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body",
			})
			return
		}

		err := services.UpdateChatTitle(chatID, req.Title)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// PUT /chat/:id/workflow - Update chat workflow
	r.PUT("/chat/:id/workflow", AuthMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")

		var req struct {
			Workflow string `json:"workflow"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body",
			})
			return
		}

		err := services.UpdateChatWorkflow(chatID, req.Workflow)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// DELETE /chat/:id - Delete chat
	r.DELETE("/chat/:id", AuthMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")

		err := services.DeleteChat(chatID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
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

	// ==================== CUSTOM PROVIDERS ENDPOINTS ====================

	// GET /custom-providers - Get user's custom providers
	r.GET("/custom-providers", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		providers, err := customProviderService.GetUserCustomProviders(userID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   providers,
		})
	})

	// POST /custom-providers - Create a new custom provider
	r.POST("/custom-providers", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req requests.CreateCustomProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		provider, err := customProviderService.CreateCustomProvider(userID, req)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"status": "ok",
			"data":   provider,
		})
	})

	// PUT /custom-providers/:id - Update a custom provider
	r.PUT("/custom-providers/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		providerID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid provider ID",
			})
			return
		}

		var req requests.UpdateCustomProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		provider, err := customProviderService.UpdateCustomProvider(userID, providerID, req)
		if err != nil {
			status := 500
			if err.Error() == "custom provider not found" {
				status = 404
			}
			c.JSON(status, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data":   provider,
		})
	})

	// DELETE /custom-providers/:id - Delete a custom provider
	r.DELETE("/custom-providers/:id", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		providerID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid provider ID",
			})
			return
		}

		err = customProviderService.DeleteCustomProvider(userID, providerID)
		if err != nil {
			status := 500
			if err.Error() == "custom provider not found" {
				status = 404
			}
			c.JSON(status, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Custom provider deleted successfully",
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

	// POST /payments/create - Create payment session
	r.POST("/payments/create", AuthMiddleware(), func(c *gin.Context) {
		var req requests.CreatePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
			})
			return
		}

		userID := c.MustGet("userID").(uuid.UUID)

		// Получаем план подписки
		planConfig, err := services.GetPlanByID(req.PlanID)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid plan ID",
			})
			return
		}

		// Создаем платежную сессию через YooKassa
		yookassaService := services.NewYooKassaService()
		price, _ := strconv.ParseInt(planConfig.Price, 10, 64)

		payment, err := yookassaService.CreatePayment(
			price,
			"RUB",
			fmt.Sprintf("Подписка %s для пользователя %s", planConfig.Name, userID.String()),
			req.ReturnURL,
			userID.String(),
			req.PlanID,
		)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  "Failed to create payment: " + err.Error(),
			})
			return
		}

		// Сохраняем информацию о платеже в базе данных (можно добавить модель Payment)
		// ...

var confirmationURL string
	if redirect, ok := payment.Confirmation.(*yoopayment.Redirect); ok {
		confirmationURL = redirect.ConfirmationURL
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"payment_id":       payment.ID,
			"confirmation_url": confirmationURL,
			"amount":           payment.Amount,
		},
	})
	})

	// POST /payments/webhook - YooKassa webhook
	r.POST("/payments/webhook", func(c *gin.Context) {
		var notification map[string]interface{}
		if err := c.ShouldBindJSON(&notification); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid webhook data"})
			return
		}

		// Обработка webhook от YooKassa
		// Здесь нужно проверить подпись и обработать платеж
		event := notification["event"].(string)
		if event == "payment.succeeded" {
			paymentObj := notification["object"].(map[string]interface{})
			_ = paymentObj["id"].(string)
			metadata := paymentObj["metadata"].(map[string]interface{})

			// Извлекаем user_id из metadata
			userIDStr, ok := metadata["user_id"].(string)
			if !ok {
				c.JSON(400, gin.H{"status": "error", "error": "Missing user_id in metadata"})
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "error": "Invalid user_id"})
				return
			}

			planID, ok := metadata["plan_id"].(string)
			if !ok {
				c.JSON(400, gin.H{"status": "error", "error": "Missing plan_id in metadata"})
				return
			}

			// Активируем подписку
			err = services.SubscribeUser(userID, planID)
			if err != nil {
				fmt.Printf("Failed to activate subscription for user %s: %v\n", userID, err)
				c.JSON(500, gin.H{"status": "error", "error": "Failed to activate subscription"})
				return
			}

			fmt.Printf("Subscription activated for user %s with plan %s\n", userID, planID)
		}

		c.JSON(200, gin.H{"status": "ok"})
	})

	// POST /payments/simulate-success - Simulate successful payment (for testing)
	r.POST("/payments/simulate-success", AuthMiddleware(), func(c *gin.Context) {
		var req requests.SimulatePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  "Invalid request body: " + err.Error(),
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

		// Активируем подписку для тестирования
		err = services.SubscribeUser(userID, req.PlanID)
		if err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Subscription activated successfully (simulated)",
		})
	})

	// Get port from env or default
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "3112"
	}

	println("User service starting on port :" + port)
	r.Run(":" + port)
}
