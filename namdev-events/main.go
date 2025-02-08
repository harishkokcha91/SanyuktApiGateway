package main

import (
	"namdev-events/database"
	"namdev-events/routes"

	"github.com/gin-gonic/gin"
)

// func init() {
// 	initializers.ConnectDatabase()
// 	initializers.DB.AutoMigrate(&models.Achievement{})
// }

func main() {
	// Initialize Database
	database.Connect()
	r := gin.Default()
	routes.EventRoutes(r)
	r.Run(":8087")
}
