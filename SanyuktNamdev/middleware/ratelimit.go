package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	cleanup  time.Duration
	rate     int
	interval time.Duration
}

type clientBucket struct {
	tokens     int
	lastRefill time.Time
}

func NewRateLimiter(requestsPerInterval int, interval time.Duration) *rateLimiter {
	rl := &rateLimiter{
		clients:  make(map[string]*clientBucket),
		cleanup:  5 * time.Minute,
		rate:     requestsPerInterval,
		interval: interval,
	}
	go rl.cleanupStaleClients()
	return rl
}

func (rl *rateLimiter) cleanupStaleClients() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, bucket := range rl.clients {
			if now.Sub(bucket.lastRefill) > rl.cleanup {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.clients[ip]
	if !exists {
		bucket = &clientBucket{
			tokens:     rl.rate - 1,
			lastRefill: time.Now(),
		}
		rl.clients[ip] = bucket
		return true
	}

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= rl.interval {
		bucket.tokens = rl.rate - 1
		bucket.lastRefill = now
		return true
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func RateLimit(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

// GlobalRateLimiter returns a global rate limiter middleware
// Default: 100 requests per minute per IP
func GlobalRateLimiter() gin.HandlerFunc {
	rl := NewRateLimiter(100, time.Minute)
	return RateLimit(rl)
}

// AuthRateLimiter returns a stricter rate limiter for auth endpoints
// Default: 10 requests per minute per IP
func AuthRateLimiter() gin.HandlerFunc {
	rl := NewRateLimiter(10, time.Minute)
	return RateLimit(rl)
}