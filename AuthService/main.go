package main

import (
	"auth-service/config"
	"auth-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	config.InitDB()

	// Create Router
	r := gin.Default()
	// Setup Routes
	routes.SetupAuthRoutes(r)

	// Start Server
	r.Run(":8080")
}
