package routes

import (
	"auth-service/controllers"
	"auth-service/middleware"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Setup authentication routes
func SetupAuthRoutes(router *gin.Engine) {

	router.POST("/register", controllers.Register)
	router.POST("/login", controllers.Login)
	router.POST("/registerOne", controllers.RegisterUserIfExistReturnUser)

	auth := router.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/protected", func(c *gin.Context) {
		fmt.Println(c.Get("userid"))
		userid, _ := c.Get("userid")
		c.JSON(200, gin.H{"message": "Welcome " + userid.(string)})
	})
}
