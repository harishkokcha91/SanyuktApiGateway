package main

import (
	"user-micro-service/database"
	"user-micro-service/routes"

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
	router.Run(":8083") // Listen on port 8080
}
