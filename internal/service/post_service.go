package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/security"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
	"github.com/google/uuid"
)

type PostService interface {
	// GetAll(ctx context.Context, params *dto.PostQueryParams) ([]*dto.PostListResponse, int64, error)
	GetAll(ctx context.Context, params *dto.PostQueryParams, currentUser *entity.User) ([]*dto.PostListResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PostResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.PostResponse, error)
	Create(ctx context.Context, req *dto.CreatePostRequest, userID uuid.UUID) (*dto.PostResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePostRequest, user *entity.User) (*dto.PostResponse, error)
	Delete(ctx context.Context, id uuid.UUID, user *entity.User) error
	Publish(ctx context.Context, id uuid.UUID, user *entity.User) (*dto.PostResponse, error)
	Unpublish(ctx context.Context, id uuid.UUID, user *entity.User) (*dto.PostResponse, error)
	IncrementViews(ctx context.Context, id uuid.UUID) error
}

type postService struct {
	postRepo     repository.PostRepository
	categoryRepo repository.CategoryRepository
	commentRepo  repository.CommentRepository
	sanitizer    security.Sanitizer
	validator    *validator.CustomValidator
}

func NewPostService(
	postRepo repository.PostRepository,
	categoryRepo repository.CategoryRepository,
	commentRepo repository.CommentRepository,
	sanitizer security.Sanitizer,
	validator *validator.CustomValidator,
) PostService {
	return &postService{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		commentRepo:  commentRepo,
		sanitizer:    sanitizer,
		validator:    validator,
	}
}

// func (s *postService) GetAll(ctx context.Context, params *dto.PostQueryParams) ([]*dto.PostListResponse, int64, error) {
// 	// Validate params
// 	if err := s.validator.Validate(params); err != nil {
// 		return nil, 0, fmt.Errorf("validation error: %w", err)
// 	}

// 	// Default pagination
// 	if params.Page < 1 {
// 		params.Page = 1
// 	}
// 	if params.Limit < 1 {
// 		params.Limit = 10
// 	}

// 	// Get posts
// 	posts, total, err := s.postRepo.FindAll(ctx, params.Page, params.Limit, params.Search, params.Status, params.CategoryID, params.Tag, params.AuthorID, params.SortBy, params.SortOrder)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to get posts: %w", err)
// 	}

// 	// Bulk count comments for all posts
// 	postIDs := make([]uuid.UUID, len(posts))
// 	for i, post := range posts {
// 		postIDs[i] = post.ID
// 	}

// 	commentCounts, err := s.commentRepo.CountByPostIDs(ctx, postIDs)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
// 	}

// 	// Convert to response with comment counts
// 	responses := make([]*dto.PostListResponse, len(posts))
// 	for i, post := range posts {
// 		commentCount := commentCounts[post.ID]
// 		responses[i] = dto.ToPostListResponse(post, commentCount)
// 	}

// 	return responses, total, nil
// }

func (s *postService) GetAll(ctx context.Context, params *dto.PostQueryParams, currentUser *entity.User) ([]*dto.PostListResponse, int64, error) {
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

	// Get posts
	posts, total, err := s.postRepo.FindAll(ctx, params.Page, params.Limit, params.Search, params.Status, params.CategoryID, params.Tag, params.AuthorID, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts: %w", err)
	}

	// ✅ Filter posts based on user role
	var filteredPosts []*entity.Post
	if currentUser != nil && !currentUser.IsAdmin() {
		// Non-admin user: show published posts + own draft/archived posts
		for _, post := range posts {
			// Show published posts from anyone
			if post.Status == entity.PostStatusPublished {
				filteredPosts = append(filteredPosts, post)
				continue
			}
			// Show own draft/archived posts
			if post.AuthorID == currentUser.ID {
				filteredPosts = append(filteredPosts, post)
				continue
			}
		}
		// Update total count after filtering
		total = int64(len(filteredPosts))
	} else {
		// Admin/SuperAdmin or public user: show all posts
		filteredPosts = posts
	}

	// Bulk count comments for all posts
	postIDs := make([]uuid.UUID, len(filteredPosts))
	for i, post := range filteredPosts {
		postIDs[i] = post.ID
	}

	commentCounts, err := s.commentRepo.CountByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Convert to response with comment counts
	responses := make([]*dto.PostListResponse, len(filteredPosts))
	for i, post := range filteredPosts {
		commentCount := commentCounts[post.ID]
		responses[i] = dto.ToPostListResponse(post, commentCount)
	}

	return responses, total, nil
}

func (s *postService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Count comments
	commentCount, err := s.commentRepo.CountByPostID(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}

	return dto.ToPostResponse(post, commentCount), nil
}

func (s *postService) GetBySlug(ctx context.Context, slug string) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Count comments
	commentCount, err := s.commentRepo.CountByPostID(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}

	return dto.ToPostResponse(post, commentCount), nil
}

func (s *postService) Create(ctx context.Context, req *dto.CreatePostRequest, userID uuid.UUID) (*dto.PostResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if category exists
	_, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Check if slug exists
	existingPost, _ := s.postRepo.FindBySlug(ctx, req.Slug)
	if existingPost != nil {
		return nil, errors.New("slug already exists")
	}

	// Sanitize content
	sanitizedContent := s.sanitizer.SanitizeHTML(req.Content)

	// Default status
	status := entity.PostStatusDraft
	if req.Status != "" {
		status = entity.PostStatus(req.Status)
	}

	// Create post
	post := &entity.Post{
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       sanitizedContent,
		Excerpt:       req.Excerpt,
		FeaturedImage: req.FeaturedImage,
		Status:        status,
		AuthorID:      userID,
		CategoryID:    req.CategoryID,
		Tags:          req.Tags,
		ViewCount:     0,
	}

	// Set published_at if status is published
	if status == entity.PostStatusPublished {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Reload with relations
	post, _ = s.postRepo.FindByID(ctx, post.ID)

	return dto.ToPostResponse(post, 0), nil
}

func (s *postService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePostRequest, user *entity.User) (*dto.PostResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get post
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Permission check
	// ✅ Author dapat update post milik sendiri
	// ✅ Admin dapat update post apapun
	// ❌ User lain tidak dapat update
	if post.AuthorID != user.ID && !user.IsAdmin() {
		return nil, errors.New("you don't have permission to update this post")
	}


	// Check category if provided
	if req.CategoryID != nil {
		_, err := s.categoryRepo.FindByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, errors.New("category not found")
		}
		post.CategoryID = *req.CategoryID
	}

	// Check slug uniqueness if changed
	if req.Slug != "" && req.Slug != post.Slug {
		existingPost, _ := s.postRepo.FindBySlug(ctx, req.Slug)
		if existingPost != nil && existingPost.ID != post.ID {
			return nil, errors.New("slug already exists")
		}
		post.Slug = req.Slug
	}

	// Update fields
	if req.Title != "" {
		post.Title = req.Title
	}

	if req.Content != "" {
		post.Content = s.sanitizer.SanitizeHTML(req.Content)
	}

	if req.Excerpt != "" {
		post.Excerpt = req.Excerpt
	}

	if req.FeaturedImage != "" {
		post.FeaturedImage = req.FeaturedImage
	}

	if len(req.Tags) > 0 {
		post.Tags = req.Tags
	}

	if req.Status != "" {
		newStatus := entity.PostStatus(req.Status)
		// Set published_at when changing to published
		if newStatus == entity.PostStatusPublished && post.Status != entity.PostStatusPublished {
			now := time.Now()
			post.PublishedAt = &now
		}
		post.Status = newStatus
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Reload with relations
	post, _ = s.postRepo.FindByID(ctx, post.ID)

	// Count comments
	commentCount, _ := s.commentRepo.CountByPostID(ctx, post.ID)

	return dto.ToPostResponse(post, commentCount), nil
}

func (s *postService) Delete(ctx context.Context, id uuid.UUID, user *entity.User) error {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("post not found")
	}

	// Permission check - Author atau Admin dapat delete
	// ✅ Author dapat delete post milik sendiri
	// ✅ Admin dapat delete post apapun
	// ❌ User lain tidak dapat delete
	if post.AuthorID != user.ID && !user.IsAdmin() {
		return errors.New("you don't have permission to delete this post")
	}

	return s.postRepo.Delete(ctx, id)
}

// Publish post - only Super Admin and Admin can publish
func (s *postService) Publish(ctx context.Context, id uuid.UUID, user *entity.User) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Permission check - only Super Admin and Admin can publish
	if !user.CanPublishPost(post) {
		return nil, errors.New("you don't have permission to publish this post")
	}

	// Check if already published
	if post.IsPublished() {
		return nil, errors.New("post is already published")
	}

	// Use entity method to publish
	post.Publish()

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to publish post: %w", err)
	}

	// Reload with relations
	post, _ = s.postRepo.FindByID(ctx, post.ID)

	// Count comments
	commentCount, _ := s.commentRepo.CountByPostID(ctx, post.ID)

	return dto.ToPostResponse(post, commentCount), nil
}

// Unpublish post - only Super Admin and Admin can unpublish
func (s *postService) Unpublish(ctx context.Context, id uuid.UUID, user *entity.User) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Permission check - only Super Admin and Admin can unpublish
	if !user.CanPublishPost(post) {
		return nil, errors.New("you don't have permission to unpublish this post")
	}

	// Check if not published
	if !post.IsPublished() {
		return nil, errors.New("post is not published")
	}

	// Update status to draft
	post.Status = entity.PostStatusDraft

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to unpublish post: %w", err)
	}

	// Reload with relations
	post, _ = s.postRepo.FindByID(ctx, post.ID)

	// Count comments
	commentCount, _ := s.commentRepo.CountByPostID(ctx, post.ID)

	return dto.ToPostResponse(post, commentCount), nil
}

func (s *postService) IncrementViews(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("post not found")
	}

	// Increment view count
	post.ViewCount++

	return s.postRepo.Update(ctx, post)
}