package main

import (
	"brilliant-student/database"
	"brilliant-student/routes"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file for local development
	_ = godotenv.Load(".env")
}

func main() {
	// Get port from environment variable
	port := os.Getenv("ACHIEVEMENT_PORT")
	if port == "" {
		port = "8082" // Default for local run
	}

	// Initialize Database
	database.Connect()

	r := gin.Default()

	// Healthcheck route
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Achievement is running"})
	})
	routes.AchievementRoutes(r)

	log.Printf("Achievement service running on port %s", port)
	r.Run(":" + port)
}
