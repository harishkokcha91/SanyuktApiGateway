package main

import (
	"business-microservice/database"
	"business-microservice/routes"
	"log"
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
	_ = godotenv.Load(".env")
}

func main() {
	// Load environment variables
	businessPort := os.Getenv("BUSINESS_PORT")
	if businessPort == "" {
		businessPort = "8083" // Default for local run
	}

	// Initialize Database
	database.Connect()

	r := gin.Default()

	// Healthcheck route
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "BusinessService is running"})
	})

	routes.BusinessRoutes(r)

	log.Printf("Business service running on port %s", businessPort)
	r.Run(":" + businessPort)
}
