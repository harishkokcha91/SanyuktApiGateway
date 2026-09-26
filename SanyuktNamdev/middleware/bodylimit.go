package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize limits request body to 10MB globally (adjust as needed)
const MaxBodySize = 10 << 20 // 10 MB

func BodyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Wrap the request body with a max bytes reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodySize)
		c.Next()
	}
}