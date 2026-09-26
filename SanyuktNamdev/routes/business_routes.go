package routes

import (
	"SanyuktNamdev/controllers"
	"SanyuktNamdev/middleware"

	"github.com/gin-gonic/gin"
)

func BusinessRoutes(router *gin.Engine) {
	businessGroup := router.Group("/businesses")
	{
		// Public read
		businessGroup.GET("/", controllers.GetBusinesses)
		businessGroup.GET("/:id", controllers.GetBusinessByID)

		// Admin write
		admin := businessGroup.Group("")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			admin.POST("/", controllers.CreateBusiness)
			admin.PUT("/:id", controllers.UpdateBusiness)
			admin.DELETE("/:id", controllers.DeleteBusiness)
		}
	}
}
