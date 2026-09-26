package main

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/middleware"
	"SanyuktNamdev/routes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
	r.Use(middleware.CORSMiddleware())

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

// Allowed file types
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// Upload image function
func uploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	// Check file size
	if file.Size > MaxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 2MB limit"})
		return
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, JPEG, PNG are allowed"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join("uploads", filename)

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Return image URL
	imageURL := fmt.Sprintf("http://localhost:8084/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"message": "Image uploaded successfully", "image_url": imageURL})
}

// rXA7/2^9Rs1*
