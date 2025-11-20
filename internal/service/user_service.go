package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/image"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/afdhali/GolangBlogpostServer/pkg/storage"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
	"github.com/google/uuid"
)

type UserService interface {
	GetAll(ctx context.Context, page, limit int, search, role string) ([]*dto.UserListResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
	GetByUsername(ctx context.Context, username string) (*dto.UserResponse, error)
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest, currentUser *entity.User) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, id uuid.UUID, req *dto.ChangePasswordRequest) error
	Delete(ctx context.Context, id uuid.UUID, currentUser *entity.User) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*dto.UserResponse, error)
	DeleteAvatar(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

type userService struct {
	userRepo       repository.UserRepository
	postRepo       repository.PostRepository
	passwordHasher security.PasswordHasher
	validator      *validator.CustomValidator
	storage        storage.Storage
	imageValidator *image.Validator
	imageProcessor *image.Processor
}

func NewUserService(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	passwordHasher security.PasswordHasher,
	validator *validator.CustomValidator,
	storage storage.Storage,
	imageValidator *image.Validator,
	imageProcessor *image.Processor,
) UserService {
	return &userService{
		userRepo:       userRepo,
		postRepo:       postRepo,
		passwordHasher: passwordHasher,
		validator:      validator,
		storage:        storage,
		imageValidator: imageValidator,
		imageProcessor: imageProcessor,
	}
}

func (s *userService) GetAll(ctx context.Context, page, limit int, search, role string) ([]*dto.UserListResponse, int64, error) {
	params := &dto.UserQueryParams{
		Page:   page,
		Limit:  limit,
		Search: search,
		Role:   role,
	}
	// Validate params
	if err := s.validator.Validate(params); err != nil {
		return nil, 0, fmt.Errorf("validation error: %w", err)
	}

	// Default pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	}

	// Get users from repository
	users, total, err := s.userRepo.FindAll(ctx, params.Page, params.Limit, params.Search, params.Role, params.IsActive, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// Bulk count posts for all users
	userIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	postCounts, err := s.postRepo.CountByAuthorIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Convert to response with post counts
	responses := make([]*dto.UserListResponse, len(users))
	for i, user := range users {
		postCount := postCounts[user.ID]
		responses[i] = dto.ToUserListResponse(user, postCount)
	}

	return responses, total, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Count posts
	postCount, err := s.postRepo.CountByAuthorID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	return dto.ToUserResponse(user, postCount), nil
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Count posts
	postCount, err := s.postRepo.CountByAuthorID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	return dto.ToUserResponse(user, postCount), nil
}

func (s *userService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if email exists
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

	// Generate default avatar if not provided
	avatar := req.Avatar
	if avatar == "" {
		name := req.FullName
		if name == "" {
			name = strings.Split(req.Email, "@")[0]
		}
		name = strings.ReplaceAll(name, " ", "+")
		avatar = fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random&size=200", name)
	}

	// Create user
	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		Role:     entity.UserRole(req.Role),
		Avatar: avatar,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return dto.ToUserResponse(user, 0), nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest, currentUser *entity.User) (*dto.UserResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Permission check
	if !currentUser.CanManageUser(user.ID) {
		return nil, errors.New("you don't have permission to update this user")
	}

	// Check email uniqueness if changed
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, errors.New("email already registered")
		}
		user.Email = req.Email
	}

	// Check username uniqueness if changed
	if req.Username != "" && req.Username != user.Username {
		existingUser, _ := s.userRepo.FindByUsername(ctx, req.Username)
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, errors.New("username already taken")
		}
		user.Username = req.Username
	}

	// Update fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}

	if req.Password != "" {
		hashedPassword, err := s.passwordHasher.Hash(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if req.Role != "" {
		user.Role = entity.UserRole(req.Role)
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Count posts
	postCount, _ := s.postRepo.CountByAuthorID(ctx, user.ID)

	return dto.ToUserResponse(user, postCount), nil
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check username uniqueness if changed
	if req.Username != "" && req.Username != user.Username {
		existingUser, _ := s.userRepo.FindByUsername(ctx, req.Username)
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, errors.New("username already taken")
		}
		user.Username = req.Username
	}

	// Update fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Count posts
	postCount, _ := s.postRepo.CountByAuthorID(ctx, user.ID)

	return dto.ToUserResponse(user, postCount), nil
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, req *dto.ChangePasswordRequest) error {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if !user.CheckPassword(req.OldPassword) {
		return errors.New("incorrect old password")
	}

	// Hash new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID, currentUser *entity.User) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Permission check
	if !currentUser.CanManageUser(user.ID) {
		return errors.New("you don't have permission to delete this user")
	}

	// Prevent self-deletion
	if user.ID == currentUser.ID {
		return errors.New("cannot delete your own account")
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *userService) UploadAvatar(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*dto.UserResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Validate image
	if err := s.imageValidator.Validate(header); err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	// Process image (compress & resize)
	processedFile, err := s.imageProcessor.Process(file, header)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Convert io.Reader to multipart.File
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, processedFile); err != nil {
		return nil, fmt.Errorf("failed to read processed image: %w", err)
	}

	// Create a new multipart.File from buffer
	processedMultipart := &bytesFile{
		Reader: bytes.NewReader(buf.Bytes()),
		size:   int64(buf.Len()),
	}

	// Save to storage
	fileInfo, err := s.storage.Save(ctx, processedMultipart, header, "avatars")
	if err != nil {
		return nil, fmt.Errorf("failed to save avatar: %w", err)
	}

	// Delete old avatar if exists and not default
	if user.Avatar != "" && !isDefaultAvatar(user.Avatar) {
		// Extract path from URL (assuming URL format: http://domain/uploads/avatars/filename.jpg)
		// We need to get the relative path: avatars/filename.jpg
		if oldPath := extractPathFromURL(user.Avatar); oldPath != "" {
			s.storage.Delete(ctx, oldPath)
		}
	}

	// Update user avatar
	user.Avatar = fileInfo.URL

	if err := s.userRepo.Update(ctx, user); err != nil {
		// Rollback: delete uploaded file
		s.storage.Delete(ctx, fileInfo.Path)
		return nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	// Count posts
	postCount, _ := s.postRepo.CountByAuthorID(ctx, user.ID)

	return dto.ToUserResponse(user, postCount), nil
}

func (s *userService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Delete file from storage if not default
	if user.Avatar != "" && !isDefaultAvatar(user.Avatar) {
		if oldPath := extractPathFromURL(user.Avatar); oldPath != "" {
			s.storage.Delete(ctx, oldPath)
		}
	}

	// Set to empty or default avatar
	user.Avatar = ""

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Count posts
	postCount, _ := s.postRepo.CountByAuthorID(ctx, user.ID)

	return dto.ToUserResponse(user, postCount), nil
}


// Helper types and functions
type bytesFile struct {
	*bytes.Reader
	size int64
}

func (b *bytesFile) Close() error {
	return nil
}

func isDefaultAvatar(avatar string) bool {
	return avatar == "" ||
		len(avatar) > len("https://ui-avatars.com") && avatar[:len("https://ui-avatars.com")] == "https://ui-avatars.com" ||
		len(avatar) > len("https://www.gravatar.com") && avatar[:len("https://www.gravatar.com")] == "https://www.gravatar.com"
}

func extractPathFromURL(url string) string {
	// Simple extraction: get everything after "/uploads/"
	const uploadsPrefix = "/uploads/"
	if idx := findString(url, uploadsPrefix); idx != -1 {
		return url[idx+len(uploadsPrefix):]
	}
	return ""
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}