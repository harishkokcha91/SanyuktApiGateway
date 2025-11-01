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
		userGroup.GET("/", controllers.GetProfiles)
		userGroup.GET("/:id", controllers.GetProfileByID)
		userGroup.GET("/byuserId/:id", controllers.GetProfilesByUserID)
		userGroup.POST("/", controllers.CreateProfile)
		userGroup.PUT("/:id", controllers.UpdateProfile)
		userGroup.DELETE("/:id", controllers.DeleteProfile)
		userGroup.POST("/upload/:id", controllers.UploadProfileImage)
	}
}
