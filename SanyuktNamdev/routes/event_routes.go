package routes

import (
	"SanyuktNamdev/controllers"
	"SanyuktNamdev/middleware"

	"github.com/gin-gonic/gin"
)

func EventRoutes(router *gin.Engine) {
	eventGroup := router.Group("/events")
	{
		// Public read
		eventGroup.GET("/", controllers.GetEvents)
		eventGroup.GET("/:id", controllers.GetEventByID)

		// Admin write
		admin := eventGroup.Group("")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			admin.POST("/", controllers.CreateEvent)
			admin.PUT("/:id", controllers.UpdateEvent)
			admin.DELETE("/:id", controllers.DeleteEvent)
		}
	}
}
