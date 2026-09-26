package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a structured logging middleware using slog
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID
		requestID, _ := c.Get("request_id")
		requestIDStr := "unknown"
		if requestID != nil {
			requestIDStr = requestID.(string)
		}

		// Prepare log attributes
		attrs := []slog.Attr{
			slog.String("request_id", requestIDStr),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if raw != "" {
			attrs = append(attrs, slog.String("query", raw))
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "HTTP request", attrs...)
		} else if c.Writer.Status() >= 400 {
			logger.LogAttrs(c.Request.Context(), slog.LevelWarn, "HTTP request", attrs...)
		} else {
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP request", attrs...)
		}
	}
}

// NewLogger creates a structured logger with service name
func NewLogger(serviceName string, isProduction bool) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if !isProduction {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler).With(slog.String("service", serviceName))
	return logger
}