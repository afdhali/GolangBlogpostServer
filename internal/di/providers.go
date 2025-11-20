package di

import (
	"github.com/afdhali/GolangBlogpostServer/config"
	"github.com/afdhali/GolangBlogpostServer/internal/handler"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/internal/router"
	"github.com/afdhali/GolangBlogpostServer/internal/service"
	"github.com/afdhali/GolangBlogpostServer/pkg/database"
	"github.com/afdhali/GolangBlogpostServer/pkg/image"
	"github.com/afdhali/GolangBlogpostServer/pkg/logger"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/afdhali/GolangBlogpostServer/pkg/storage"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
	"gorm.io/gorm"
)

// AppContainer holds all dependencies and provides cleanup
type AppContainer struct {
	Router *router.Router
	db     *gorm.DB
	logger *logger.Logger
}

// Cleanup performs graceful shutdown
func (c *AppContainer) Cleanup() {
	// Close database connection
	if c.db != nil {
		if sqlDB, err := c.db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// Close logger
	if c.logger != nil {
		c.logger.Close()
	}
}

// ProvideConfig loads application configuration
func ProvideConfig() (*config.Config, error) {
	return config.LoadConfig()
}

// ProvideLogger creates logger instance
func ProvideLogger() (*logger.Logger, error) {
	return logger.NewLogger("./logs")
}

// ProvideDatabase creates database connection
func ProvideDatabase(cfg *config.Config) (*gorm.DB, error) {
	return database.NewPostgresDB(cfg)
}

// ProvideValidator creates validator instance
func ProvideValidator() *validator.CustomValidator {
	return validator.NewValidator()
}

// ProvidePasswordHasher creates password hasher
func ProvidePasswordHasher(cfg *config.Config) security.PasswordHasher {
	return security.NewPasswordHasher(cfg.Security.BcryptCost, 8, 72)
}

// ProvideJWTService creates JWT service
func ProvideJWTService(cfg *config.Config) security.JWTService {
	return security.NewJWTService(cfg)
}

// ProvideSanitizer creates HTML sanitizer
func ProvideSanitizer() security.Sanitizer {
	return security.NewSanitizer()
}

// ProvideStorage creates storage instance
func ProvideStorage(cfg *config.Config) storage.Storage {
	return storage.NewLocalStorage(cfg.Storage.BasePath, cfg.Storage.BaseURL)
}

// ProvideImageValidator creates image validator
func ProvideImageValidator(cfg *config.Config) *image.Validator {
	return image.NewValidator(cfg.Storage.MaxSizeMB, []string{"image/jpeg", "image/jpg", "image/png", "image/webp"})
}

// ProvideImageProcessor creates image processor
func ProvideImageProcessor() *image.Processor {
	return image.DefaultAvatarProcessor()
}

// ============================================================================
// REPOSITORIES
// ============================================================================

// Repositories
func ProvideUserRepository(db *gorm.DB) repository.UserRepository {
	return repository.NewUserRepository(db)
}

func ProvideCategoryRepository(db *gorm.DB) repository.CategoryRepository {
	return repository.NewCategoryRepository(db)
}

func ProvidePostRepository(db *gorm.DB) repository.PostRepository {
	return repository.NewPostRepository(db)
}

func ProvideCommentRepository(db *gorm.DB) repository.CommentRepository {
	return repository.NewCommentRepository(db)
}

func ProvideRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return repository.NewRefreshTokenRepository(db)
}

// 👇 ADD THIS - Media Repository Provider
func ProvideMediaRepository(db *gorm.DB) repository.MediaRepository {
	return repository.NewMediaRepository(db)
}

// ============================================================================
// SERVICES
// ============================================================================

// Services
func ProvideAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordHasher security.PasswordHasher,
	jwtService security.JWTService,
	validator *validator.CustomValidator,
	cfg *config.Config,
) service.AuthService {
	return service.NewAuthService(userRepo, refreshTokenRepo, passwordHasher, jwtService, validator, cfg)
}

func ProvideUserService(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	passwordHasher security.PasswordHasher,
	validator *validator.CustomValidator,
	storage storage.Storage,
	imageValidator *image.Validator,
	imageProcessor *image.Processor,
) service.UserService {
	return service.NewUserService(userRepo, postRepo, passwordHasher, validator, storage, imageValidator, imageProcessor)
}

func ProvideCategoryService(
	categoryRepo repository.CategoryRepository,
	postRepo repository.PostRepository,
	validator *validator.CustomValidator,
) service.CategoryService {
	return service.NewCategoryService(categoryRepo, postRepo, validator)
}

func ProvidePostService(
	postRepo repository.PostRepository,
	categoryRepo repository.CategoryRepository,
	commentRepo repository.CommentRepository,
	sanitizer security.Sanitizer,
	validator *validator.CustomValidator,
) service.PostService {
	return service.NewPostService(postRepo, categoryRepo, commentRepo, sanitizer, validator)
}

func ProvideCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	sanitizer security.Sanitizer,
	validator *validator.CustomValidator,
) service.CommentService {
	return service.NewCommentService(commentRepo, postRepo, sanitizer, validator)
}

// 👇 ADD THIS - Media Service Provider
func ProvideMediaService(
	mediaRepo repository.MediaRepository,
	postRepo repository.PostRepository,
	storage storage.Storage,
	imageValidator *image.Validator,
	imageProcessor *image.Processor,
	validator *validator.CustomValidator,
) service.MediaService {
	return service.NewMediaService(mediaRepo, postRepo, storage, imageValidator, imageProcessor, validator)
}

// ============================================================================
// HANDLERS
// ============================================================================

// Handlers
func ProvideAuthHandler(authService service.AuthService) *handler.AuthHandler {
	return handler.NewAuthHandler(authService)
}

func ProvideUserHandler(userService service.UserService) *handler.UserHandler {
	return handler.NewUserHandler(userService)
}

func ProvideCategoryHandler(categoryService service.CategoryService) *handler.CategoryHandler {
	return handler.NewCategoryHandler(categoryService)
}

func ProvidePostHandler(postService service.PostService) *handler.PostHandler {
	return handler.NewPostHandler(postService)
}

func ProvideCommentHandler(commentService service.CommentService) *handler.CommentHandler {
	return handler.NewCommentHandler(commentService)
}

func ProvideMediaHandler(mediaService service.MediaService) *handler.MediaHandler {
	return handler.NewMediaHandler(mediaService)
}

// ============================================================================
// ROUTER
// ============================================================================

// Router
func ProvideRouter(
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
) *router.Router {
	return router.NewRouter(
		cfg,
		logger,
		jwtService,
		userRepo,
		authHandler,
		userHandler,
		categoryHandler,
		postHandler,
		commentHandler,
		mediaHandler, 
	)
}

// ProvideAppContainer creates the app container with cleanup capabilities
func ProvideAppContainer(
	router *router.Router,
	db *gorm.DB,
	logger *logger.Logger,
) *AppContainer {
	return &AppContainer{
		Router: router,
		db:     db,
		logger: logger,
	}
}