package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ApiGateway/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log/slog"
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
	// Initialize structured logger
	isProduction := os.Getenv("ENV") == "production"
	logger := middleware.NewLogger("apigateway", isProduction)
	slog.SetDefault(logger)

	r := gin.New()

	// Add middleware
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.SecurityHeaders(isProduction))
	r.Use(middleware.CORSMiddleware())

	// Health check endpoints
	r.GET("/healthz", healthCheck)
	r.GET("/readyz", readyCheck)

	// Read service URLs from environment variables
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	profileServiceURL := os.Getenv("PROFILE_SERVICE_URL")
	brilliantStudentURL := os.Getenv("BRILLIANT_STUDENT_URL")
	imageUploadURL := os.Getenv("IMAGE_UPLOAD_URL")
	eventsServiceURL := os.Getenv("EVENTS_SERVICE_URL")
	businessServiceURL := os.Getenv("BUSINESS_SERVICE_URL")

	// Log service URLs
	logger.Info("Service URLs configured",
		"auth", authServiceURL,
		"user", userServiceURL,
		"profile", profileServiceURL,
		"brilliant_student", brilliantStudentURL,
		"image_upload", imageUploadURL,
		"events", eventsServiceURL,
		"business", businessServiceURL,
	)

	// Routes
	r.Any("/auth/*rest", reverseProxy(authServiceURL))
	r.Any("/user/*rest", reverseProxy(userServiceURL))
	r.Any("/profile/*rest", reverseProxy(profileServiceURL))
	r.Any("/brilliantstudent/*rest", reverseProxy(brilliantStudentURL))
	r.Any("/image/*rest", reverseProxy(imageUploadURL))
	r.Any("/namdevevents/*rest", reverseProxy(eventsServiceURL))
	r.Any("/namdevbusinesses/*rest", reverseProxy(businessServiceURL))

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Run server in goroutine
	go func() {
		logger.Info("API Gateway running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give in-flight requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

// healthCheck checks if the gateway is healthy
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// readyCheck checks if the gateway is ready to serve traffic
func readyCheck(c *gin.Context) {
	// For gateway, ready = healthy (no DB to check)
	healthCheck(c)
}