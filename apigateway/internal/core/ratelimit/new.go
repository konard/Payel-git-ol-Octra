package ratelimit

import (
	"log"
	"strconv"
	"strings"
	"time"
)

func newTokenBucket(maxTokens int, windowSeconds int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(maxTokens),
		maxTokens:  float64(maxTokens),
		refillRate: float64(maxTokens) / float64(windowSeconds),
		lastRefill: time.Now(),
	}
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
