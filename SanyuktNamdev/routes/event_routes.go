package routes

import (
	"SanyuktNamdev/controllers"

	"github.com/gin-gonic/gin"
)

func EventRoutes(router *gin.Engine) {
	eventGroup := router.Group("/events")
	{
		eventGroup.GET("/", controllers.GetEvents)
		eventGroup.GET("/:id", controllers.GetEventByID)
		eventGroup.POST("/", controllers.CreateEvent)
		eventGroup.PUT("/:id", controllers.UpdateEvent)
		eventGroup.DELETE("/:id", controllers.DeleteEvent)
	}
}
