package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

func init() {
	_ = godotenv.Load(".env") // Load local env file
}

func main() {
	r := gin.Default()

	// Healthcheck route
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API Gateway is running"})
	})

	// CORS middleware configuration
	r.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:50001", "http://127.0.0.1:50001"},
		AllowOrigins:     []string{"*"}, // Adjust frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

	// Start the API Gateway
	log.Println("API Gateway running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
