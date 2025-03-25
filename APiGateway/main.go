package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

func main() {
	r := gin.Default()

	// Healthcheck route
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API Gateway is running"})
	})

	// CORS middleware configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Adjust frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Routes to services
	r.Any("/auth/*rest", reverseProxy("http://localhost:8080"))
	r.Any("/user/*rest", reverseProxy("http://localhost:8083"))             // User service
	r.Any("/profile/*rest", reverseProxy("http://localhost:8082"))          // Profile service
	r.Any("/brilliantstudent/*rest", reverseProxy("http://localhost:8085")) // Profile service
	r.Any("/image/*rest", reverseProxy("http://localhost:8086"))            //Image upload service
	r.Any("/namdevevents/*rest", reverseProxy("http://localhost:8087"))     //Events service
	r.Any("/namdevbusinesses/*rest", reverseProxy("http://localhost:8088")) //Business service

	// Start the API Gateway
	log.Println("API Gateway running on http://localhost:8084")
	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
