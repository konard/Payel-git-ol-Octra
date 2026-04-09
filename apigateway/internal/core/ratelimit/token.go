package ratelimit

import (
	"sync"
	"time"
)

// tokenBucket implements token bucket algorithm
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}
