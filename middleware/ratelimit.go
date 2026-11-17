package middleware

import (
	"github.com/gin-gonic/gin"
	// TODO(5): "golang.org/x/time/rate" (token bucket) and "net/http" for the 429 response.
)

// RateLimiter limits requests per client.
//
// TODO(5): Create a rate.NewLimiter(1, 5) -> 1 request/second, burst of 5 (make it configurable).
// TODO(5): Decide scope: one global limiter (simplest) or one per client IP (map + mutex, c.ClientIP()).
// TODO(5): If !limiter.Allow(), c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"}).
// TODO(5): (Optional) Evict stale per-IP limiters so the map doesn't grow forever.
// TODO(7): Test by firing >5 quick requests (e.g. a curl loop) and confirm 429s, then recovery.
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // pass-through until implemented
	}
}
