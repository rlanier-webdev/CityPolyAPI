package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	limiterMu sync.Mutex
)

// RateLimitMiddleware limits requests per IP address
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiterMu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			// Allow 10 requests per second with burst of 20
			limiter = rate.NewLimiter(10, 20)
			limiters[ip] = limiter
		}
		limiterMu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please slow down.",
			})
			return
		}
		c.Next()
	}
}