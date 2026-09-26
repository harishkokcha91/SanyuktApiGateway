package middleware

import (
	"SanyuktNamdev/utils"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RecoveryMiddleware recovers from panics and logs with request ID
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				requestIDStr := "unknown"
				if requestID != nil {
					requestIDStr = fmt.Sprintf("%v", requestID)
				}

				// Log the panic with stack trace and request ID
				utils.Logger.Error("PANIC recovered",
					"request_id", requestIDStr,
					"error", err,
					"stack", string(debug.Stack()),
				)

				utils.RespondInternalError(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}