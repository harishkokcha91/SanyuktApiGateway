package main

import (
	"SanyuktNamdev/config"
	"SanyuktNamdev/routes"
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
	port := os.Getenv("AUTHSERVICE_PORT")
	if port == "" {
		port = "8081" // Default for local run
	}

	// Initialize Database
	config.InitDB()

	// Create Router
	r := gin.Default()
	// Setup Routes
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

	log.Printf("Achievement service running on port %s", port)
	r.Run(":" + port)
}
