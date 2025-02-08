package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Max file size (2MB)
const MaxUploadSize = 2 << 20 // 2MB

// Allowed file types
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

func main() {
	// Create uploads directory if not exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}

	router := gin.Default()

	// Route for image upload
	router.POST("/upload", uploadImage)

	// Serve uploaded images
	router.Static("/uploads", "./uploads")

	// Start server
	port := "8086"
	log.Printf("Server started on http://localhost:%s", port)
	router.Run(":" + port)
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
	imageURL := fmt.Sprintf("http://localhost:8086/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"message": "Image uploaded successfully", "image_url": imageURL})
}
