package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/afdhali/GolangBlogpostServer/config"
	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error)
	Logout(ctx context.Context, refreshtToken string) error
}

type authService struct {
	userRepo 			repository.UserRepository
	refreshTokenRepo 	repository.RefreshTokenRepository
	passwordHasher 		security.PasswordHasher
	jwtService 			security.JWTService
	validator 			*validator.CustomValidator
	config 				*config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordHasher security.PasswordHasher,
	jwtService security.JWTService,
	validator *validator.CustomValidator,
	config *config.Config,
) AuthService {
	return &authService{
		userRepo: 			userRepo,
		refreshTokenRepo: 	refreshTokenRepo,
		passwordHasher: 	passwordHasher,
		jwtService: 		jwtService,
		validator: 			validator,
		config: 			config,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w",err)
	}

	// check if email exists
	existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Check if username exists
	existingUser, _ = s.userRepo.FindByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user with default avatar
	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		Avatar:   s.generateDefaultAvatar(req.Email, req.FullName),
		Role:     entity.RoleUser,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	tokenEntity := &entity.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWT.RefreshTokenExpiry) * time.Second),
	}

	if err := s.refreshTokenRepo.Create(ctx, tokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return dto.ToAuthResponse(user, accessToken, refreshToken, int64(s.config.JWT.AccessTokenExpiry)), nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := s.passwordHasher.Verify(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	tokenEntity := &entity.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWT.RefreshTokenExpiry) * time.Second),
	}

	if err := s.refreshTokenRepo.Create(ctx, tokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return dto.ToAuthResponse(user, accessToken, refreshToken, int64(s.config.JWT.AccessTokenExpiry)), nil
}

func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Verify refresh token
	claims, err := s.jwtService.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check token type
	if claims["type"] != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Check if token exists in database
	tokenEntity, err := s.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	// Check if token is expired
	if tokenEntity.ExpiresAt.Before(time.Now()) {
		if err := s.refreshTokenRepo.Revoke(ctx, req.RefreshToken); err != nil {
			return nil, fmt.Errorf("failed to revoke expired refresh token: %w", err)
		}
		return nil, errors.New("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, tokenEntity.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Revoke old refresh token
	if err := s.refreshTokenRepo.Revoke(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Store new refresh token
	newTokenEntity := &entity.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWT.RefreshTokenExpiry) * time.Second),
	}

	if err := s.refreshTokenRepo.Create(ctx, newTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return dto.ToTokenResponse(accessToken, newRefreshToken, int64(s.config.JWT.AccessTokenExpiry)), nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// Find refresh token
	_, err := s.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return errors.New("refresh token not found")
	}

	// Revoke the token
	return s.refreshTokenRepo.Revoke(ctx, refreshToken)
}

// Helper function to generate default avatar
func (s *authService) generateDefaultAvatar(email, fullName string) string {
	// Option 1: Gravatar (commented out)
	// hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	// return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=200", hash)

	// Option 2: UI Avatars (with initials)
	name := strings.ReplaceAll(fullName, " ", "+")
	if name == "" {
		name = strings.Split(email, "@")[0]
	}
	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random&size=200", name)
}