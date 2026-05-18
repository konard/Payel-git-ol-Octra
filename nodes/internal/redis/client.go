package redis

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps go-redis client with CrewAI-specific methods
type Client struct {
	rdb     *redis.Client
	enabled bool
}

// NewClient creates and initializes Redis client
func NewClient() *Client {
	redisURL := os.Getenv("REDIS_URL")
	enabled := os.Getenv("REDIS_ENABLED") == "true"

	if !enabled || redisURL == "" {
		log.Println("[Redis] Disabled or no REDIS_URL - running without Redis")
		return &Client{enabled: false}
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("[Redis] Failed to parse REDIS_URL: %v, running without Redis", err)
		return &Client{enabled: false}
	}

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[Redis] Failed to connect: %v, running without Redis", err)
		return &Client{enabled: false}
	}

	log.Printf("[Redis] Connected to %s", redisURL)
	return &Client{
		rdb:     rdb,
		enabled: true,
	}
}

// IsEnabled returns true if Redis is available
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// Close closes Redis connection
func (c *Client) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

// GetRedisClient returns underlying Redis client for advanced operations
func (c *Client) GetRedisClient() *redis.Client {
	return c.rdb
}
