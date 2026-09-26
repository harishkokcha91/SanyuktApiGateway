package routes

import (
	"SanyuktNamdev/controllers"
	"SanyuktNamdev/middleware"

	"github.com/gin-gonic/gin"
)

// NotificationPreferenceRoutes sets up notification preference endpoints
func NotificationPreferenceRoutes(router *gin.Engine) {
	prefGroup := router.Group("/notification-preferences")
	prefGroup.Use(middleware.AuthMiddleware())
	{
		prefController := controllers.NewNotificationPreferenceController()
		prefGroup.GET("", prefController.GetNotificationPreferences)
		prefGroup.PATCH("/:channel", prefController.UpdateNotificationPreference)
		prefGroup.POST("/reset", prefController.ResetNotificationPreferences)
	}
}