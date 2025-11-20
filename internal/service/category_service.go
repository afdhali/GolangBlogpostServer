package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/afdhali/GolangBlogpostServer/pkg/validator"
	"github.com/google/uuid"
)

type CategoryService interface {
	GetAll(ctx context.Context, params *dto.CategoryQueryParams) ([]*dto.CategoryListResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.CategoryResponse, error)
	Create(ctx context.Context, req *dto.CreateCategoryRequest, user *entity.User) (*dto.CategoryResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest, user *entity.User) (*dto.CategoryResponse, error)
	Delete(ctx context.Context, id uuid.UUID, user *entity.User) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
	postRepo     repository.PostRepository
	validator    *validator.CustomValidator
}

func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	postRepo repository.PostRepository,
	validator *validator.CustomValidator,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		postRepo:     postRepo,
		validator:    validator,
	}
}

func (s *categoryService) GetAll(ctx context.Context, params *dto.CategoryQueryParams) ([]*dto.CategoryListResponse, int64, error) {
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

	// Get categories
	categories, total, err := s.categoryRepo.FindAll(ctx, params.Page, params.Limit, params.Search, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get categories: %w", err)
	}

	// Bulk count posts for all categories
	categoryIDs := make([]uuid.UUID, len(categories))
	for i, category := range categories {
		categoryIDs[i] = category.ID
	}

	postCounts, err := s.postRepo.CountByCategoryIDs(ctx, categoryIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Convert to response with post counts
	responses := make([]*dto.CategoryListResponse, len(categories))
	for i, category := range categories {
		postCount := postCounts[category.ID]
		responses[i] = dto.ToCategoryListResponse(category, postCount)
	}

	return responses, total, nil
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Count posts
	postCount, err := s.postRepo.CountByCategoryID(ctx, category.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	return dto.ToCategoryResponse(category, postCount), nil
}

func (s *categoryService) GetBySlug(ctx context.Context, slug string) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Count posts
	postCount, err := s.postRepo.CountByCategoryID(ctx, category.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	return dto.ToCategoryResponse(category, postCount), nil
}

func (s *categoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest, user *entity.User) (*dto.CategoryResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if slug exists
	existingCategory, _ := s.categoryRepo.FindBySlug(ctx, req.Slug)
	if existingCategory != nil {
		return nil, errors.New("slug already exists")
	}

	// Create category
	category := &entity.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}

	// Permission Check
	if !user.CanManageCategory(category) {
		return nil, errors.New("you don't have permission to create category")
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return dto.ToCategoryResponse(category, 0), nil
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest, user *entity.User) (*dto.CategoryResponse, error) {
	// Validate request
	if err := s.validator.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get category
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Permission Check
	if !user.CanManageCategory(category) {
		return nil, errors.New("you don't have permission to update category")
	}

	// Check slug uniqueness if changed
	if req.Slug != "" && req.Slug != category.Slug {
		existingCategory, _ := s.categoryRepo.FindBySlug(ctx, req.Slug)
		if existingCategory != nil && existingCategory.ID != category.ID {
			return nil, errors.New("slug already exists")
		}
		category.Slug = req.Slug
	}

	// Update fields
	if req.Name != "" {
		category.Name = req.Name
	}

	if req.Description != "" {
		category.Description = req.Description
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	// Count posts
	postCount, _ := s.postRepo.CountByCategoryID(ctx, category.ID)

	return dto.ToCategoryResponse(category, postCount), nil
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID, user *entity.User) error {
	// Get category
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("category not found")
	}

	// Permission Check
	if !user.CanManageCategory(category) {
		return errors.New("you don't have permission to delete category")
	}

	// Check if category has posts
	postCount, err := s.postRepo.CountByCategoryID(ctx, category.ID)
	if err != nil {
		return fmt.Errorf("failed to count posts: %w", err)
	}

	if postCount > 0 {
		return errors.New("cannot delete category with posts")
	}

	return s.categoryRepo.Delete(ctx, id)
}