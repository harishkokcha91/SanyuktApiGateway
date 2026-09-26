package middleware

import (
	"fmt"
	"time"

	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerPeriod int
	Period            time.Duration
	KeyPrefix         string
}

// defaultRateLimiter creates a rate limiter with the given config
func defaultRateLimiter(config RateLimitConfig) *limiter.Limiter {
	rate := limiter.Rate{
		Period: config.Period,
		Limit:  int64(config.RequestsPerPeriod),
	}
	store := memory.NewStore()
	return limiter.New(store, rate)
}

// GlobalRateLimiter returns middleware for global rate limiting
func GlobalRateLimiter() gin.HandlerFunc {
	config := RateLimitConfig{
		RequestsPerPeriod: 100,
		Period:            time.Minute,
		KeyPrefix:         "global",
	}
	l := defaultRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyPrefix + ":" + c.ClientIP()
		context, err := l.Get(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", time.Unix(context.Reset, 0).Format(time.RFC3339))

		if context.Reached {
			c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// AuthRateLimiter returns stricter rate limiting for auth endpoints
func AuthRateLimiter() gin.HandlerFunc {
	config := RateLimitConfig{
		RequestsPerPeriod: 5,
		Period:            time.Minute,
		KeyPrefix:         "auth",
	}
	l := defaultRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyPrefix + ":" + c.ClientIP()
		context, err := l.Get(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", "5")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", time.Unix(context.Reset, 0).Format(time.RFC3339))

		if context.Reached {
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many attempts. Please try again later."})
			return
		}
		c.Next()
	}
}

// SecurityHeaders returns secure headers middleware
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	config := secure.Config{
		// Prevent MIME type sniffing
		ContentTypeNosniff: true,
		// Prevent clickjacking
		FrameDeny: true,
		// XSS protection (legacy but harmless)
		BrowserXssFilter: true,
		// Content Security Policy
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'",
		// Referrer policy
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}

	// Only enable HSTS in production with HTTPS
	if isProduction {
		config.STSSeconds = 31536000 // 1 year
		config.STSIncludeSubdomains = true
		config.STSPreload = true
		config.SSLRedirect = true
		config.SSLProxyHeaders = map[string]string{"X-Forwarded-Proto": "https"}
	}

	return secure.New(config)
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// SecurityCORSMiddleware returns CORS middleware with explicit origins
func SecurityCORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range config.AllowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
			// Only allow wildcard if explicitly configured (dev only)
			c.Header("Access-Control-Allow-Origin", "*")
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", joinMethods(config.AllowedMethods))
		c.Header("Access-Control-Allow-Headers", joinHeaders(config.AllowedHeaders))
		if len(config.ExposedHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinHeaders(config.ExposedHeaders))
		}
		c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func joinMethods(methods []string) string {
	result := ""
	for i, m := range methods {
		if i > 0 {
			result += ", "
		}
		result += m
	}
	return result
}

func joinHeaders(headers []string) string {
	result := ""
	for i, h := range headers {
		if i > 0 {
			result += ", "
		}
		result += h
	}
	return result
}

// DefaultCORSConfig returns a development-friendly CORS config
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",  // React default
			"http://localhost:5173",  // Vite default
			"http://localhost:4200",  // Angular default
			"http://localhost:8080",  // Vue default
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Content-Length", "X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}
}

// ProductionCORSConfig returns a production CORS config (update origins before deploy)
func ProductionCORSConfig(frontendDomain string) CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			frontendDomain, // e.g., "https://app.yourdomain.com"
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Content-Length", "X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}
}