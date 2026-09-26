package main

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/middleware"
	"SanyuktNamdev/routes"
	"SanyuktNamdev/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
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

	// Initialize Database
	// config.InitDB()
	// Initialize Database
	database.Connect()
	// Create Router
	r := gin.New()

	// Add middleware
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.SecurityCORSMiddleware(middleware.DefaultCORSConfig()))
	r.Use(middleware.SecurityHeaders(false)) // false = dev mode (no HSTS/SSL redirect)
	r.Use(middleware.BodyLimitMiddleware())
	r.Use(middleware.GlobalRateLimiter())

	// Healthcheck route
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "AuthService is running"})
	})
	routes.SetupAuthRoutes(r)
	routes.AchievementRoutes(r)
	routes.BusinessRoutes(r)
	routes.EventRoutes(r)
	routes.UserRoutes(r)
	routes.ProfileRoutes(r)
	// Route for image upload (requires auth)
	r.POST("/upload", middleware.AuthMiddleware(), uploadImage)

	log.Printf("Achievement service running on port %s", port)
	r.Run(":" + port)
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
