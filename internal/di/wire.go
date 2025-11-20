//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
)

// InitializeApp initializes the entire application with all dependencies
func InitializeApp() (*AppContainer, error) {
	wire.Build(
		// ============================================================================
		// CORE INFRASTRUCTURE (Config, Logger, Database)
		// ============================================================================
		ProvideConfig,
		ProvideLogger,
		ProvideDatabase,
		ProvideValidator,

		// ============================================================================
		// SECURITY & STORAGE (depends on Config)
		// ============================================================================
		ProvidePasswordHasher,
		ProvideJWTService,
		ProvideSanitizer,
		ProvideStorage,
		ProvideImageValidator,
		ProvideImageProcessor,

		// ============================================================================
		// LAYER 1: REPOSITORIES (depends on Database)
		// ============================================================================
		ProvideUserRepository,
		ProvideCategoryRepository,
		ProvidePostRepository,
		ProvideCommentRepository,
		ProvideRefreshTokenRepository,
		ProvideMediaRepository, 

		// ============================================================================
		// LAYER 2: SERVICES (depends on Repositories + Security/Storage)
		// ============================================================================
		ProvideAuthService,
		ProvideUserService,
		ProvideCategoryService,
		ProvidePostService,
		ProvideCommentService,
		ProvideMediaService, 

		// ============================================================================
		// LAYER 3: HANDLERS (depends on Services)
		// ============================================================================
		ProvideAuthHandler,
		ProvideUserHandler,
		ProvideCategoryHandler,
		ProvidePostHandler,
		ProvideCommentHandler,
		ProvideMediaHandler, 

		// ============================================================================
		// ROUTER & CONTAINER (depends on Handlers)
		// ============================================================================
		ProvideRouter,
		ProvideAppContainer,
	)

	return nil, nil
}

/*
DEPENDENCY INJECTION ORDER:

  1. Infrastructure
     ├─ Config
     ├─ Logger
     ├─ Database
     └─ Validator

  2. Security & Storage (requires Config)
     ├─ PasswordHasher
     ├─ JWTService
     ├─ Sanitizer
     ├─ Storage
     ├─ ImageValidator
     └─ ImageProcessor

  3. REPOSITORIES (requires Database)
     ├─ UserRepository
     ├─ CategoryRepository
     ├─ PostRepository
     ├─ CommentRepository
     ├─ RefreshTokenRepository
     └─ MediaRepository 

  4. SERVICES (requires Repositories + Security/Storage)
     ├─ AuthService
     ├─ UserService
     ├─ CategoryService
     ├─ PostService
     ├─ CommentService
     └─ MediaService 

  5. HANDLERS (requires Services)
     ├─ AuthHandler
     ├─ UserHandler
     ├─ CategoryHandler
     ├─ PostHandler
     ├─ CommentHandler
     └─ MediaHandler 

  6. ROUTER & CONTAINER (requires Handlers)
     ├─ Router
     └─ AppContainer

KEY POINTS:
- Order MUST follow dependency chain
- Repositories before Services
- Services before Handlers
- Handlers before Router
- Wire will fail if order is wrong
*/