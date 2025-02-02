package main

import (
	"userprofile-service/database"
	"userprofile-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	database.Connect()

	// Set up Router
	router := gin.Default()

	// Register Routes
	routes.UserRoutes(router)

	// Start Server
	router.Run(":8082") // Listen on port 8080
}
