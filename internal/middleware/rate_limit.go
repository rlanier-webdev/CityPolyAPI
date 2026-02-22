package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

var (
	limiters = make(map[string]*ipLimiter)
	limiterMu sync.Mutex
)

func init() {
	go cleanupLimiters()
}

func cleanupLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		limiterMu.Lock()
		for ip, entry := range limiters {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(limiters, ip)
			}
		}
		limiterMu.Unlock()
	}
}

// RateLimitMiddleware limits requests per IP address
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiterMu.Lock()
		entry, exists := limiters[ip]
		if !exists {
			entry = &ipLimiter{limiter: rate.NewLimiter(10, 20)}
			limiters[ip] = entry
		}
		entry.lastSeen = time.Now()
		limiterMu.Unlock()

		if !entry.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please slow down.",
			})
			return
		}
		c.Next()
	}
}