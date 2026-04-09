package main

import (
	"log"
	"os"

	"apigateway/internal/fetcher"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Register all routes
	fetcher.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3111"
	}

	log.Printf("🚀 Apigateway starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
