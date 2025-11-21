package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/afdhali/GolangBlogpostServer/pkg/logger"
	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/gin-gonic/gin"
)

func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				log.Error("PANIC RECOVERED: %v\nStack Trace:\n%s", err, debug.Stack())
				response.Error(ctx, http.StatusInternalServerError, "Internal server error", nil)
				ctx.Abort()
			}
		}()

		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last()
			// Log error with request details
			log.Error("Request Error - Method: %s, Path: %s, Error: %s", 
				ctx.Request.Method, 
				ctx.Request.URL.Path, 
				err.Error())
			response.Error(ctx, http.StatusInternalServerError, err.Error(), nil)
		}
	}
}