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
		const bearerPrefix = "Bearer "
		if len(tokenString) < len(bearerPrefix) || tokenString[:len(bearerPrefix)] != bearerPrefix {
			utils.RespondUnauthorized(c, "Invalid Authorization header format")
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		tokenString = tokenString[len(bearerPrefix):]

		// Check for empty token after Bearer prefix
		if tokenString == "" {
			utils.RespondUnauthorized(c, "Invalid Authorization header format")
			c.Abort()
			return
		}

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

		// Admin role has access to all roles
		if userRole == "admin" {
			c.Next()
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