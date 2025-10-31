package routes

import (
	"SanyuktNamdev/controllers"

	"github.com/gin-gonic/gin"
)

func ProfileRoutes(router *gin.Engine) {
	// Serve static files from the "uploads" directory
	router.Static("/uploads", "./uploads")
	userGroup := router.Group("/matrimonialProfiles")
	{
		userGroup.GET("/", controllers.GetUsers)
		userGroup.GET("/:id", controllers.GetUserByID)
		userGroup.GET("/byuserId/:id", controllers.GetProfileByUserID)
		userGroup.POST("/", controllers.CreateUser)
		userGroup.PUT("/:id", controllers.UpdateUser)
		userGroup.DELETE("/:id", controllers.DeleteUser)
		userGroup.POST("/upload/:id", controllers.UploadImageForUser)
	}
}
