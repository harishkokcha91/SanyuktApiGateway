package main

import (
	"log"
	"os"
	"userprofile-service/database"
	"userprofile-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file if present
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found or error loading it")
	}
}

func main() {
	// Get port from environment variable
	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "8087" // Default for local run
	}

	// Initialize Database
	database.Connect()

	// Set up Router
	router := gin.Default()

	// Register Routes
	routes.UserRoutes(router)

	// Start Server
	log.Printf("User service running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
