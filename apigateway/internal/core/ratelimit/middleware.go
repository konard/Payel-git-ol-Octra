// Package ratelimit provides in-memory rate limiting using token bucket algorithm
package ratelimit

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket implements token bucket algorithm
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func newTokenBucket(maxTokens int, windowSeconds int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(maxTokens),
		maxTokens:  float64(maxTokens),
		refillRate: float64(maxTokens) / float64(windowSeconds),
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

// RateLimiter manages per-endpoint rate limiting
type RateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
}

// New creates a new rate limiter with config from env
func New() *RateLimiter {
	rl := &RateLimiter{buckets: make(map[string]*tokenBucket)}
	rl.addBucket("task_create", parseLimit("RATE_LIMIT_TASK_CREATE", 10))
	rl.addBucket("task_status", parseLimit("RATE_LIMIT_TASK_STATUS", 60))
	rl.addBucket("task_reconnect", parseLimit("RATE_LIMIT_TASK_RECONNECT", 30))
	rl.addBucket("health", parseLimit("RATE_LIMIT_HEALTH", 120))

	var info []string
	for name, b := range rl.buckets {
		windowSec := int(b.maxTokens / b.refillRate)
		info = append(info, name+"="+strconv.Itoa(int(b.maxTokens))+"/"+strconv.Itoa(windowSec)+"s")
	}
	log.Printf("[RateLimit] Enabled: %s", strings.Join(info, ", "))

	return rl
}

type limit struct {
	max    int
	window int
}

func parseLimit(envVar string, defaultMax int) limit {
	val := os.Getenv(envVar)
	if val == "" {
		return limit{max: defaultMax, window: 60}
	}
	parts := strings.Split(val, "/")
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return limit{max: defaultMax, window: 60}
	}
	window := 60
	if len(parts) > 1 {
		w, err := strconv.Atoi(parts[1])
		if err == nil && w > 0 {
			window = w
		}
	}
	return limit{max: n, window: window}
}

func (rl *RateLimiter) addBucket(name string, l limit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.buckets[name] = newTokenBucket(l.max, l.window)
}

// GinMiddleware returns a gin-compatible middleware for the given bucket name
func (rl *RateLimiter) GinMiddleware(bucketName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rl.mu.RLock()
		bucket := rl.buckets[bucketName]
		rl.mu.RUnlock()

		if bucket == nil || bucket.allow() {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"status":  "error",
			"error":   "rate limit exceeded",
			"message": "Too many requests. Please try again later.",
		})
	}
}
