package main

import (
	"log"
	"os"
	"user-micro-service/database"
	"user-micro-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load(".env") // Load local env file if present
}

func main() {
	// Get port from environment variable
	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "8086" // Default for local run
	}

	// Initialize Database
	database.Connect()

	// Set up Router
	router := gin.Default()

	// Register Routes
	routes.UserRoutes(router)

	// Start Server
	log.Printf("User service running on port %s", port)
	router.Run(":" + port)
}
