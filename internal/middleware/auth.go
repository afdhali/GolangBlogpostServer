package middleware

import (
	"net/http"
	"strings"

	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware(jwtService security.JWTService, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(ctx, http.StatusUnauthorized, "Authorization header is required", nil)
			ctx.Abort()
			return 
		}

		 parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Error(ctx, http.StatusUnauthorized, "Invalid authorization header format", nil)
            ctx.Abort()
            return
        }

        token := parts[1]

        claims, err := jwtService.VerifyToken(token)
        if err != nil {
            response.Error(ctx, http.StatusUnauthorized, "Invalid or expired token", nil)
            ctx.Abort()
            return
        }

        userIDStr, ok := claims["user_id"].(string)
        if !ok {
            response.Error(ctx, http.StatusUnauthorized, "Invalid token claims", nil)
            ctx.Abort()
            return
        }

		userID, err := uuid.Parse(userIDStr)
        if err != nil {
            response.Error(ctx, http.StatusUnauthorized, "Invalid user ID in token", nil)
            ctx.Abort()
            return
        }

        user, err := userRepo.FindByID(ctx.Request.Context(), userID)
        if err != nil {
            response.Error(ctx, http.StatusUnauthorized, "User not found", nil)
            ctx.Abort()
            return
        }

        if !user.IsActive {
            response.Error(ctx, http.StatusUnauthorized, "User account is inactive", nil)
            ctx.Abort()
            return
        }

        ctx.Set("user", user)
        ctx.Set("user_id", user.ID)
        ctx.Set("user_role", user.Role)

        ctx.Next()
	}
}