package middleware

import (
	"net/http"

	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestAPIKey := ctx.GetHeader("X-API-KEY")

		if requestAPIKey == "" {
			response.Error(ctx, http.StatusUnauthorized, "API Key is required", nil)
			ctx.Abort()
			return
		}

		if requestAPIKey != apiKey {
			response.Error(ctx, http.StatusUnauthorized, "Invalid API KEY", nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}