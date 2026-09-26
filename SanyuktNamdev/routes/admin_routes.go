package routes

import (
	"SanyuktNamdev/controllers"
	"SanyuktNamdev/middleware"

	"github.com/gin-gonic/gin"
)

// AdminRoutes sets up admin-only approval endpoints
func AdminRoutes(router *gin.Engine) {
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		// Approve a record
		adminGroup.PATCH("/:type/:id/approve", controllers.AdminApprove)

		// Reject a record
		adminGroup.PATCH("/:type/:id/reject", controllers.AdminReject)
	}
}