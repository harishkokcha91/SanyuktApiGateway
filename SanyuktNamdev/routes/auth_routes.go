package routes

import (
	"SanyuktNamdev/controllers"
	"SanyuktNamdev/middleware"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Setup authentication routes
func SetupAuthRoutes(router *gin.Engine) {
	// Stricter rate limiting on auth endpoints
	authRateLimiter := middleware.AuthRateLimiter()

	router.POST("/register", authRateLimiter, controllers.Register)
	router.POST("/login", authRateLimiter, controllers.Login)
	router.POST("/registerOne", authRateLimiter, controllers.RegisterUserIfExistReturnUser)

	auth := router.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/protected", func(c *gin.Context) {
		fmt.Println(c.Get("userid"))
		userid, _ := c.Get("userid")
		c.JSON(200, gin.H{"message": userid.(string)})
	})
}
