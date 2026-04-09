package ratelimit

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// RateLimiter manages per-endpoint rate limiting
type RateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
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
