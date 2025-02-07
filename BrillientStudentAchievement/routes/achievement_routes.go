package routes

import (
	"brilliant-student/controllers"

	"github.com/gin-gonic/gin"
)

func AchievementRoutes(router *gin.Engine) {
	achievementGroup := router.Group("/brilliantstudent/achievements")
	{
		achievementGroup.POST("/", controllers.CreateAchievement)
		achievementGroup.GET("/", controllers.GetAchievements)
		achievementGroup.GET("/:id", controllers.GetAchievementByID)
		achievementGroup.PUT("/:id", controllers.UpdateAchievement)
		achievementGroup.DELETE("/:id", controllers.DeleteAchievement)
	}
}
