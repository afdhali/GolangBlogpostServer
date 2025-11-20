package middleware

import (
	"net/http"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...entity.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, exists := ctx.Get("user")
		if !exists {
			response.Error(ctx, http.StatusUnauthorized, "User Not Authenticated", nil)
			ctx.Abort()
			return 
		}

		currentUser := user.(*entity.User)

		hasPermission := false
		for _, role := range allowedRoles {
			if currentUser.Role == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Error(ctx, http.StatusForbidden, "You Dont Have Permission to access this resource",nil)
			ctx.Abort()
			return 
		}

		ctx.Next()
	}
}

func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(entity.RoleSuperAdmin)
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(entity.RoleSuperAdmin, entity.RoleAdmin)
}

func RequireUser() gin.HandlerFunc {
	return RequireRole(entity.RoleSuperAdmin, entity.RoleAdmin, entity.RoleUser)
}

// RBACMiddleware is a wrapper for RequireRole that accepts string roles
func RBACMiddleware(roles ...string) gin.HandlerFunc {
	entityRoles := make([]entity.UserRole, len(roles))
	for i, role := range roles {
		entityRoles[i] = entity.UserRole(role)
	}
	return RequireRole(entityRoles...)
}