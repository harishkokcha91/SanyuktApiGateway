package main

import (
	"brilliant-student/database"
	"brilliant-student/routes"

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
	routes.AchievementRoutes(r)
	r.Run(":8085")
}
