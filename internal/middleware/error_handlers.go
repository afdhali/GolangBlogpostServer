package middleware

import (
	"net/http"

	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.Error(ctx, http.StatusInternalServerError, "Internal server error", err)
				ctx.Abort()
			}
		}()

		ctx.Next()

		if len(ctx.Errors) >0 {
			err := ctx.Errors.Last()
			response.Error(ctx, http.StatusInternalServerError, err.Error(),nil)
		}
	}
}