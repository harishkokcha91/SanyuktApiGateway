package middleware

import (
	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			utils.RespondUnauthorized(c, "No token provided")
			c.Abort()
			return
		}

		// Remove the "Bearer " prefix from the Authorization header to get the token
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			utils.RespondUnauthorized(c, "Invalid Authorization header format")
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		tokenString = tokenString[7:]

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.RespondUnauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		// Set the userid and role from the claims into the context
		c.Set("userid", claims["userid"])
		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		} else {
			c.Set("role", "user")
		}
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.RespondForbidden(c, "Role not found in token")
			c.Abort()
			return
		}

		if userRole != role {
			utils.RespondForbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}