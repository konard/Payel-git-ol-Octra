// Package ratelimit provides in-memory rate limiting using token bucket algorithm
package ratelimit

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

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
