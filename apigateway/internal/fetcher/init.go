package fetcher

import (
	"errors"
	"log"
	"net/http"
	"os"

	"apigateway/internal/core/ratelimit"
	"apigateway/internal/core/redis"
	"apigateway/internal/core/services"
	"apigateway/internal/fetcher/grpc/boss"
	"apigateway/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func validateJWT(tokenString string) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("JWT_SECRET not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid claims")
	}

	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)
	if userID == "" {
		return "", "", errors.New("user_id missing in token")
	}
	return userID, username, nil
}

func init() {
	var err error
	bossHost := os.Getenv("BOSS_SERVICE_HOST")
	if bossHost == "" {
		bossHost = "orchestrator:50052"
	}
	bossClient, err = boss.NewClient(bossHost)
	if err != nil {
		log.Printf("Warning: failed to connect to Boss service: %v", err)
	}

	redisClient = redis.NewClient()
	if redisClient.IsEnabled() {
		pubSubManager = redis.NewPubSubManager(redisClient)
		log.Println("[Redis] PubSub manager initialized")
	}

	wsHub = services.NewHub()
	go wsHub.Run()

	rl = ratelimit.New()

	// Initialize database
	db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: failed to connect to database: %v", err)
	} else {
		db.AutoMigrate(&models.Task{})
	}
}

// RegisterRoutes registers all HTTP routes on the gin engine
func RegisterRoutes(r *gin.Engine) {
	// Rate limit middleware wrappers
	rlHealth := rl.GinMiddleware("health")
	rlTaskCreate := rl.GinMiddleware("task_create")
	rlTaskStatus := rl.GinMiddleware("task_status")
	rlTaskReconnect := rl.GinMiddleware("task_reconnect")

	r.GET("/health", rlHealth, healthHandler)
	r.GET("/task/create", rlTaskCreate, handleTaskCreateWS)
	r.GET("/task/reconnect", rlTaskReconnect, handleTaskReconnectWS)
	r.GET("/task/status", rlTaskStatus, handleTaskStatus)
	r.POST("/task/:taskId/stop", rlTaskStatus, handleTaskStop)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
