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

func OptionalAuthMiddleware(jwtService security.JWTService, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		
		// ✅ STEP 1: Jika tidak ada Authorization header → continue as public
		if authHeader == "" {
			// User tidak login, continue as public
			ctx.Next()
			return
		}

		// ✅ STEP 2: Parse Authorization header
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format tapi jangan abort, lanjut as public
			ctx.Next()
			return
		}

		token := parts[1]

		// ✅ STEP 3: Verify token
		claims, err := jwtService.VerifyToken(token)
		if err != nil {
			// Token invalid/expired tapi jangan abort, lanjut as public
			ctx.Next()
			return
		}

		// ✅ STEP 4: Extract user_id dari claims
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			// Invalid token claims tapi jangan abort, lanjut as public
			ctx.Next()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			// Invalid user ID tapi jangan abort, lanjut as public
			ctx.Next()
			return
		}

		// ✅ STEP 5: Get user dari database
		user, err := userRepo.FindByID(ctx.Request.Context(), userID)
		if err != nil {
			// User tidak ditemukan tapi jangan abort, lanjut as public
			ctx.Next()
			return
		}

		// ✅ STEP 6: Check if user is active
		if !user.IsActive {
			// User inactive tapi jangan abort, lanjut as public
			ctx.Next()
			return
		}

		// ✅ STEP 7: User valid → set ke context (berbeda dengan public)
		ctx.Set("user", user)
		ctx.Set("user_id", user.ID)
		ctx.Set("user_role", user.Role)

		ctx.Next()
	}
}