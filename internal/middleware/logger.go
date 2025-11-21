package middleware

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/pkg/logger"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware injects logger into context and logs all requests
func LoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Inject logger into context for use in handlers
		c.Set("logger", log)

		// Log request start
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log request completion
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		if statusCode >= 400 {
			// Log errors
			log.Error("Request - Method: %s, Path: %s, Status: %d, Latency: %v, ClientIP: %s",
				method, path, statusCode, latency, clientIP)
		} else {
			// Log successful requests
			log.Info("Request - Method: %s, Path: %s, Status: %d, Latency: %v, ClientIP: %s",
				method, path, statusCode, latency, clientIP)
		}
	}
}