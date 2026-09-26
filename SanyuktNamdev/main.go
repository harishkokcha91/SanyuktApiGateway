package main

import (
	"context"
	"SanyuktNamdev/database"
	"SanyuktNamdev/middleware"
	"SanyuktNamdev/routes"
	"SanyuktNamdev/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log/slog"
)

func init() {
	// Load .env file for local development
	_ = godotenv.Load(".env")
}

func main() {
	// Get port from environment variable
	port := os.Getenv("AUTHSERVICE_PORT")
	if port == "" {
		port = "8081" // Default for local run
	}

	// Initialize structured logger
	isProduction := os.Getenv("ENV") == "production"
	logger := middleware.NewLogger("sanyuktnamdev", isProduction)
	slog.SetDefault(logger)

	// Initialize Database
	database.Connect()
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Create Router
	r := gin.New()

	// Add middleware
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.SecurityCORSMiddleware(middleware.DefaultCORSConfig()))
	r.Use(middleware.SecurityHeaders(isProduction))
	r.Use(middleware.BodyLimitMiddleware())
	r.Use(middleware.GlobalRateLimiter())

	// Health check endpoints
	r.GET("/healthz", healthCheck)
	r.GET("/readyz", readyCheck)

	routes.SetupAuthRoutes(r)
	routes.AchievementRoutes(r)
	routes.BusinessRoutes(r)
	routes.EventRoutes(r)
	routes.UserRoutes(r)
	routes.ProfileRoutes(r)
	// Route for image upload (requires auth)
	r.POST("/upload", middleware.AuthMiddleware(), uploadImage)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Run server in goroutine
	go func() {
		logger.Info("Achievement service running on port " + port)
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

// healthCheck checks if the service is healthy (DB connectivity)
func healthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	db := database.WithContext(ctx)
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection failed"})
		return
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// readyCheck checks if the service is ready to serve traffic
func readyCheck(c *gin.Context) {
	// For now, ready = healthy
	healthCheck(c)
}

// Max file size (2MB)
const MaxUploadSize = 2 << 20 // 2MB

// Allowed MIME types
var allowedMimes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// Upload image function
func uploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		utils.RespondValidationError(c, "Image file is required")
		return
	}

	// Check file size
	if file.Size > MaxUploadSize {
		utils.RespondValidationError(c, "File size exceeds 2MB limit")
		return
	}

	// Open and validate MIME
	src, err := file.Open()
	if err != nil {
		utils.RespondInternalError(c, "Failed to open uploaded file")
		return
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		utils.RespondInternalError(c, "Failed to read file for MIME detection")
		return
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		utils.RespondInternalError(c, "Failed to reset file reader")
		return
	}

	mimeType := http.DetectContentType(buffer[:n])
	if !allowedMimes[mimeType] {
		utils.RespondValidationError(c, fmt.Sprintf("Invalid file type: %s. Only JPEG and PNG allowed", mimeType))
		return
	}

	// Extension from MIME
	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, 0755); err != nil {
			utils.RespondInternalError(c, "Failed to create upload directory")
			return
		}
	}

	uuidStr := uuid.New().String()
	filename := fmt.Sprintf("%s%s", uuidStr, ext)
	filePath := filepath.Join(uploadDir, filename)

	// Path traversal protection
	absUploadDir, _ := filepath.Abs(uploadDir)
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, absUploadDir) {
		utils.RespondInternalError(c, "Invalid file path")
		return
	}

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.RespondInternalError(c, "Failed to save image")
		return
	}

	imageURL := fmt.Sprintf("http://localhost:8084/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"message": "Image uploaded successfully", "image_url": imageURL})
}

// rXA7/2^9Rs1*
