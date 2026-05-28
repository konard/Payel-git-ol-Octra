package main

import (
	"log"
	"os"

	"user/internal/core/services"
	"user/internal/fetcher/http/router"
	"user/pkg/database"
	"user/pkg/models"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDb()
	database.Db.AutoMigrate(&models.UserRegister{}, &models.Subscription{}, &models.PromoCode{}, &models.Workflow{}, &models.CustomProvider{}, &models.CustomModel{}, &models.Chat{}, &models.ChatMessage{})

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
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
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

	router.RegisterAllRoutes(r)

	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "3112"
	}

	log.Printf("User service starting on port %s", port)
	r.Run(":" + port)
}