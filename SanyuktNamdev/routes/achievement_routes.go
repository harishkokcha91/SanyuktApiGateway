package routes

import (
	"SanyuktNamdev/controllers"
	"SanyuktNamdev/middleware"

	"github.com/gin-gonic/gin"
)

func AchievementRoutes(router *gin.Engine) {
	achievementGroup := router.Group("/achievements")
	{
		// Public read
		achievementGroup.GET("/", controllers.GetAchievements)
		achievementGroup.GET("/:id", controllers.GetAchievementByID)

		// Admin write
		admin := achievementGroup.Group("")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			admin.POST("/", controllers.CreateAchievement)
			admin.PUT("/:id", controllers.UpdateAchievement)
			admin.DELETE("/:id", controllers.DeleteAchievement)
		}
	}
}
