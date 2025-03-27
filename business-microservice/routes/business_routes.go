package routes

import (
	"business-microservice/controllers"

	"github.com/gin-gonic/gin"
)

func BusinessRoutes(router *gin.Engine) {
	businessGroup := router.Group("/businesses")
	{
		businessGroup.GET("/", controllers.GetBusinesses)
		businessGroup.GET("/:id", controllers.GetBusinessByID)
		businessGroup.POST("/", controllers.CreateBusiness)
		businessGroup.PUT("/:id", controllers.UpdateBusiness)
		businessGroup.DELETE("/:id", controllers.DeleteBusiness)
	}
}
