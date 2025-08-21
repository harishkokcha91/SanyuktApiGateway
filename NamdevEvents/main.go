package main

import (
	"log"
	"namdev-events/database"
	"namdev-events/routes"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// func init() {
// 	initializers.ConnectDatabase()
// 	initializers.DB.AutoMigrate(&models.Achievement{})
// }

func init() {
	_ = godotenv.Load(".env") // Load local env file if present
}

func main() {
	// Get port from environment variable
	port := os.Getenv("EVENT_SERVICE_PORT")
	if port == "" {
		port = "8085" // Default for local run
	}

	// Initialize Database
	database.Connect()

	r := gin.Default()
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Events is running"})
	})
	routes.EventRoutes(r)

	log.Printf("Event service running on port %s", port)
	r.Run(":" + port)
}
