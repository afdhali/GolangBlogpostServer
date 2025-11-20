package middleware

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
        AllowMethods:     cfg.CORS.AllowedMethods,
        AllowHeaders:     cfg.CORS.AllowedHeaders,
        AllowCredentials: cfg.CORS.AllowCredentials,
        MaxAge:           12 * time.Hour,
	})
}