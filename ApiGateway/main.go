package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// reverseProxy sets up a reverse proxy for the target service
func reverseProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the target URL
		targetURL, err := url.Parse(target)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
			return
		}

		// Create a reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Modify the request path to match the target service
		c.Request.URL.Path = c.Param("rest")

		// Serve the request using the proxy
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// rateLimiter creates a token-bucket rate limiter
func rateLimiter(requestsPerPeriod int, period time.Duration, keyPrefix string) *limiter.Limiter {
	rate := limiter.Rate{
		Period: period,
		Limit:  int64(requestsPerPeriod),
	}
	store := memory.NewStore()
	return limiter.New(store, rate)
}

// globalRateLimiter returns middleware for global rate limiting (100 req/min per IP)
func globalRateLimiter() gin.HandlerFunc {
	l := rateLimiter(100, time.Minute, "gateway-global")

	return func(c *gin.Context) {
		key := "gateway-global:" + c.ClientIP()
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

// securityHeaders returns secure headers middleware
func securityHeaders(isProduction bool) gin.HandlerFunc {
	config := secure.Config{
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}

	if isProduction {
		config.STSSeconds = 31536000
		config.STSIncludeSubdomains = true
		config.STSPreload = true
		config.SSLRedirect = true
		config.SSLProxyHeaders = map[string]string{"X-Forwarded-Proto": "https"}
	}

	return secure.New(config)
}

// corsMiddleware returns CORS middleware with explicit origins
func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:4200",
		"http://localhost:8080",
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
		c.Header("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func init() {
	_ = godotenv.Load(".env") // Load local env file
}

func main() {
	r := gin.New()

	// Add middleware
	r.Use(globalRateLimiter())
	r.Use(securityHeaders(false)) // false = dev mode
	r.Use(corsMiddleware())

	// Healthcheck route
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API Gateway is running"})
	})

	// Read service URLs from environment variables
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	profileServiceURL := os.Getenv("PROFILE_SERVICE_URL")
	brilliantStudentURL := os.Getenv("BRILLIANT_STUDENT_URL")
	imageUploadURL := os.Getenv("IMAGE_UPLOAD_URL")
	eventsServiceURL := os.Getenv("EVENTS_SERVICE_URL")
	businessServiceURL := os.Getenv("BUSINESS_SERVICE_URL")

	// Routes
	r.Any("/auth/*rest", reverseProxy(authServiceURL))
	r.Any("/user/*rest", reverseProxy(userServiceURL))
	r.Any("/profile/*rest", reverseProxy(profileServiceURL))
	r.Any("/brilliantstudent/*rest", reverseProxy(brilliantStudentURL))
	r.Any("/image/*rest", reverseProxy(imageUploadURL))
	r.Any("/namdevevents/*rest", reverseProxy(eventsServiceURL))
	r.Any("/namdevbusinesses/*rest", reverseProxy(businessServiceURL))

	fmt.Println("AuthService URLs:", authServiceURL)
	fmt.Println("BrilliantStudent URLs:", brilliantStudentURL)
	fmt.Println("BusinessService URLs:", businessServiceURL)
	fmt.Println("ImageUpload URLs:", imageUploadURL)
	fmt.Println("EventsService URLs:", eventsServiceURL)
	fmt.Println("UserService URLs:", userServiceURL)
	fmt.Println("ProfileService URLs:", profileServiceURL)
	fmt.Println("API Gateway is running...")
	log.Println("API Gateway running on http://localhost:8080")

	// TLS Strategy: This service assumes TLS termination at a reverse proxy/load balancer
	// (nginx, Traefik, AWS ALB, Cloudflare, etc.). If running without a proxy,
	// replace r.Run(":8080") with r.RunTLS(":8443", "cert.pem", "key.pem")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}