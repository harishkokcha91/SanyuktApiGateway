package main

import (
	"business-microservice/database"
	"business-microservice/routes"

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
	routes.BusinessRoutes(r)
	r.Run(":8088")
}
