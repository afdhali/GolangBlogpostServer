package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
	"github.com/google/uuid"
)

type CommentService interface {
	Create(ctx context.Context, postID uuid.UUID, req *dto.CreateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error)
	GetByPostID(ctx context.Context, postID uuid.UUID, params *dto.CommentQueryParams) ([]*dto.CommentResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCommentRequest, user *entity.User) (*dto.CommentResponse, error)
	Delete(ctx context.Context, id uuid.UUID, user *entity.User) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	sanitizer   security.Sanitizer
	validator   *validator.CustomValidator
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	sanitizer security.Sanitizer,
	validator *validator.CustomValidator,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		sanitizer:   sanitizer,
		validator:   validator,
	}
}

func (s *commentService) Create(ctx context.Context, postID uuid.UUID, req *dto.CreateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if post exists
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// If has parent, check parent exists and belongs to same post
	if req.ParentID != nil {
		parentComment, err := s.commentRepo.FindByID(ctx, *req.ParentID)
		if err != nil {
			return nil, errors.New("parent comment not found")
		}
		if parentComment.PostID != postID {
			return nil, errors.New("parent comment does not belong to this post")
		}
	}

	// Sanitize content - use StrictSanitize for comments
	sanitizedContent := s.sanitizer.StrictSanitize(req.Content)

	// Create comment
	comment := &entity.Comment{
		Content:  sanitizedContent,
		PostID:   postID,
		UserID:   userID,
		ParentID: req.ParentID,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Load relations (User, Replies)
	comment, _ = s.commentRepo.FindByID(ctx, comment.ID)

	return dto.ToCommentResponse(comment), nil
}

func (s *commentService) GetByPostID(ctx context.Context, postID uuid.UUID, params *dto.CommentQueryParams) ([]*dto.CommentResponse, int64, error) {
	// Validate params
	if err := s.validator.Validate(params); err != nil {
		return nil, 0, fmt.Errorf("validation error: %w", err)
	}

	// Check if post exists
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, 0, errors.New("post not found")
	}

	// Default pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	}

	// Get comments
	comments, total, err := s.commentRepo.FindByPostID(ctx, postID, params.Page, params.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}

	// Convert to response
	responses := make([]*dto.CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = dto.ToCommentResponse(comment)
	}

	return responses, total, nil
}

func (s *commentService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCommentRequest, user *entity.User) (*dto.CommentResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get comment
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("comment not found")
	}

	// Permission check - Author atau Admin dapat update
	// ✅ Author dapat update comment milik sendiri
	// ✅ Admin dapat update comment apapun
	// ❌ User lain tidak dapat update
	if comment.UserID != user.ID && !user.IsAdmin() {
		return nil, errors.New("you don't have permission to update this comment")
	}

	// Sanitize content
	comment.Content = s.sanitizer.StrictSanitize(req.Content)

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	// Reload with relations
	comment, _ = s.commentRepo.FindByID(ctx, comment.ID)

	return dto.ToCommentResponse(comment), nil
}

func (s *commentService) Delete(ctx context.Context, id uuid.UUID, user *entity.User) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("comment not found")
	}

	// Permission check - Author atau Admin dapat delete
	// ✅ Author dapat delete comment milik sendiri
	// ✅ Admin dapat delete comment apapun
	// ❌ User lain tidak dapat delete
	if comment.UserID != user.ID && !user.IsAdmin() {
		return errors.New("you don't have permission to delete this comment")
	}

	return s.commentRepo.Delete(ctx, id)
}