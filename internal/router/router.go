package router

import (
	"github.com/afdhali/GolangBlogpostServer/config"
	"github.com/afdhali/GolangBlogpostServer/internal/handler"
	"github.com/afdhali/GolangBlogpostServer/internal/middleware"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/logger"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/gin-gonic/gin"
)

type Router struct {
	cfg             *config.Config
	logger          *logger.Logger
	jwtService      security.JWTService
	userRepo        repository.UserRepository
	authHandler     *handler.AuthHandler
	userHandler     *handler.UserHandler
	categoryHandler *handler.CategoryHandler
	postHandler     *handler.PostHandler
	commentHandler  *handler.CommentHandler
	mediaHandler    *handler.MediaHandler 
}

func NewRouter(
	cfg *config.Config,
	logger *logger.Logger,
	jwtService security.JWTService,
	userRepo repository.UserRepository,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	categoryHandler *handler.CategoryHandler,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
	mediaHandler *handler.MediaHandler, 
) *Router {
	return &Router{
		cfg:             cfg,
		logger:          logger,
		jwtService:      jwtService,
		userRepo:        userRepo,
		authHandler:     authHandler,
		userHandler:     userHandler,
		categoryHandler: categoryHandler,
		postHandler:     postHandler,
		commentHandler:  commentHandler,
		mediaHandler:    mediaHandler, 
	}
}

func (r *Router) Setup() *gin.Engine {
	// Set Gin mode
	if r.cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware(r.cfg))
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.ErrorHandler())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Serve static files (uploads)
	router.Static("/uploads", r.cfg.Storage.BasePath)

	// API routes
	api := router.Group("/api/v1")
	api.Use(middleware.APIKeyMiddleware(r.cfg.Security.APIKey))
	{
		// Public routes - Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.RefreshToken)
			auth.POST("/logout", r.authHandler.Logout)
		}

		// Public routes - Categories (read only)
		categories := api.Group("/categories")
		{
			categories.GET("", r.categoryHandler.GetAll)
			categories.GET("/:id", r.categoryHandler.GetByID)
			categories.GET("/slug/:slug", r.categoryHandler.GetBySlug)
		}

		// Public routes - Posts (read only)
		posts := api.Group("/posts")
		{
			posts.GET("", r.postHandler.GetAll)
			posts.GET("/slug/:slug", r.postHandler.GetBySlug)
			posts.GET("/:id", r.postHandler.GetByID)
			posts.POST("/:id/views", r.postHandler.IncrementViews)

			// Comments for specific post
			posts.GET("/:id/comments", r.commentHandler.GetByPostID)

			// 👇 ADD THESE MEDIA ROUTES (PUBLIC)
			// Get all media for a post (use :id not :postId to avoid conflicts)
			posts.GET("/:id/media", r.mediaHandler.GetByPostID)
			// Get featured media for a post
			posts.GET("/:id/media/featured", r.mediaHandler.GetFeaturedByPostID)
		}

		// Protected routes - require authentication
		authMiddleware := middleware.AuthMiddleware(r.jwtService, r.userRepo)

		// User profile routes
		profile := api.Group("/profile")
		profile.Use(authMiddleware)
		{
			profile.GET("", r.userHandler.GetProfile)
			profile.PUT("", r.userHandler.UpdateProfile)
			profile.PUT("/password", r.userHandler.ChangePassword)
			profile.POST("/avatar", r.userHandler.UploadAvatar)
			profile.DELETE("/avatar", r.userHandler.DeleteAvatar)
		}

		// User management routes (Admin only)
		users := api.Group("/users")
		users.Use(authMiddleware)
		{
			users.GET("", middleware.RequireAdmin(), r.userHandler.GetAll)
			users.GET("/:id", r.userHandler.GetByID)
			users.POST("", middleware.RequireSuperAdmin(), r.userHandler.CreateUser)
			users.PUT("/:id", middleware.RequireAdmin(), r.userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireSuperAdmin(), r.userHandler.DeleteUser)
		}

		// Category management routes (Admin only)
		categoryManagement := api.Group("/categories")
		categoryManagement.Use(authMiddleware, middleware.RequireAdmin())
		{
			categoryManagement.POST("", r.categoryHandler.Create)
			categoryManagement.PUT("/:id", r.categoryHandler.Update)
			categoryManagement.DELETE("/:id", r.categoryHandler.Delete)
		}

		// Post management routes
		postManagement := api.Group("/posts")
		postManagement.Use(authMiddleware)
		{
			postManagement.POST("", r.postHandler.Create)
			postManagement.PUT("/:id", r.postHandler.Update)
			postManagement.DELETE("/:id", r.postHandler.Delete)
			postManagement.POST("/:id/publish", middleware.RequireAdmin(), r.postHandler.Publish)
			postManagement.POST("/:id/unpublish", middleware.RequireAdmin(), r.postHandler.Unpublish)
		}

		// Comment management routes
		commentManagement := api.Group("/posts/:id/comments")
		commentManagement.Use(authMiddleware)
		{
			commentManagement.POST("", r.commentHandler.Create)
		}

		comments := api.Group("/comments")
		comments.Use(authMiddleware)
		{
			comments.PUT("/:commentId", r.commentHandler.Update)
			comments.DELETE("/:commentId", r.commentHandler.Delete)
		}

		// 👇 ADD THESE MEDIA ROUTES (PROTECTED & PUBLIC)
		// Media routes - Public (list & detail)
		mediaPublic := api.Group("/media")
		{
			mediaPublic.GET("", r.mediaHandler.GetAll)
			mediaPublic.GET("/:id", r.mediaHandler.GetByID)
		}

		// Media routes - Protected (upload, update, delete)
		mediaProtected := api.Group("/media")
		mediaProtected.Use(authMiddleware)
		{
			mediaProtected.POST("", r.mediaHandler.Upload)
			mediaProtected.PUT("/:id", r.mediaHandler.Update)
			mediaProtected.DELETE("/:id", r.mediaHandler.Delete)
		}
	}

	return router
}