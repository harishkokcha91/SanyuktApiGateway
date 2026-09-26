package middleware

import (
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// SecurityHeaders returns secure headers middleware
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	config := secure.Config{
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}

	if isProduction {
		config.STSSeconds = 31536000
		config.STSIncludeSubdomains = true
		config.STSPreload = true
		config.SSLRedirect = true
		config.SSLProxyHeaders = map[string]string{"X-Forwarded-Proto": "https"}
	}

	return secure.New(config)
}