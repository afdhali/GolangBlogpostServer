package repository

import (
	"context"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(ctx context.Context, post *entity.Post) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Post, error)
	FindAll(ctx context.Context, page, limit int, search, status string, categoryID *uuid.UUID, tag string, authorID *uuid.UUID, sortBy, sortOrder string) ([]*entity.Post, int64, error)
	Update(ctx context.Context, post *entity.Post) error
	Delete(ctx context.Context, id uuid.UUID) error

    // 👇 For Dynamic Counting Posts
    CountByAuthorID(ctx context.Context, authorID uuid.UUID) (int64, error)
    CountByAuthorIDs(ctx context.Context, authorIDs []uuid.UUID) (map[uuid.UUID]int64, error)
    CountByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error)
    CountByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type postRepository struct {
    db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
    return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *entity.Post) error {
    return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Post, error) {
    var post entity.Post
    err := r.db.WithContext(ctx).
        Preload("Author").
        Preload("Category").
        Where("id = ?", id).
        First(&post).Error
    if err != nil {
        return nil, err
    }
    return &post, nil
}

func (r *postRepository) FindBySlug(ctx context.Context, slug string) (*entity.Post, error) {
    var post entity.Post
    err := r.db.WithContext(ctx).
        Preload("Author").
        Preload("Category").
        Where("slug = ?", slug).
        First(&post).Error
    if err != nil {
        return nil, err
    }
    return &post, nil
}

func (r *postRepository) FindAll(ctx context.Context, page, limit int, search, status string, categoryID *uuid.UUID, tag string, authorID *uuid.UUID, sortBy, sortOrder string) ([]*entity.Post, int64, error) {
	var posts []*entity.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Post{}).
		Preload("Author").
		Preload("Category")

	// Apply filters
	if search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if authorID != nil {
		query = query.Where("author_id = ?", *authorID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if tag != "" {
		// Assuming a many-to-many relationship with tags table
		// Adjust based on actual schema: post_tags (post_id, tag_id), tags (id, name)
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.name ILIKE ?", "%"+tag+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	// Sorting
	order := "created_at DESC"
	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "DESC"
		}
		order = sortBy + " " + sortOrder
	}

	err := query.Offset(offset).Limit(limit).Order(order).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) Update(ctx context.Context, post *entity.Post) error {
    return r.db.WithContext(ctx).Save(post).Error
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&entity.Post{}, id).Error
}

// 👇 NEW METHOD: Count posts by single author
func (r *postRepository) CountByAuthorID(ctx context.Context, authorID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&entity.Post{}).
        Where("author_id = ?", authorID).
        Count(&count).Error
    return count, err
}

// 👇 NEW METHOD: Count posts for multiple authors (efficient bulk query)
func (r *postRepository) CountByAuthorIDs(ctx context.Context, authorIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
    if len(authorIDs) == 0 {
        return make(map[uuid.UUID]int64), nil
    }

    type Result struct {
        AuthorID uuid.UUID
        Count    int64
    }

    var results []Result
    err := r.db.WithContext(ctx).
        Model(&entity.Post{}).
        Select("author_id, COUNT(*) as count").
        Where("author_id IN ?", authorIDs).
        Group("author_id").
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    // Convert to map
    countMap := make(map[uuid.UUID]int64)
    for _, result := range results {
        countMap[result.AuthorID] = result.Count
    }

    // Fill in zeros for authors with no posts
    for _, authorID := range authorIDs {
        if _, exists := countMap[authorID]; !exists {
            countMap[authorID] = 0
        }
    }

    return countMap, nil
}

// 👇 NEW METHOD: Count posts for multiple categories (efficient bulk query)
func (r *postRepository) CountByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(categoryIDs) == 0 {
		return make(map[uuid.UUID]int64), nil
	}

	type Result struct {
		CategoryID uuid.UUID
		Count      int64
	}

	var results []Result
	err := r.db.WithContext(ctx).
		Model(&entity.Post{}).
		Select("category_id, COUNT(*) as count").
		Where("category_id IN ?", categoryIDs).
		Group("category_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to map
	countMap := make(map[uuid.UUID]int64)
	for _, result := range results {
		countMap[result.CategoryID] = result.Count
	}

	// Fill in zeros for categories with no posts
	for _, categoryID := range categoryIDs {
		if _, exists := countMap[categoryID]; !exists {
			countMap[categoryID] = 0
		}
	}

	return countMap, nil
}

// 👇 NEW METHOD: Count posts by category
func (r *postRepository) CountByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&entity.Post{}).
        Where("category_id = ?", categoryID).
        Count(&count).Error
    return count, err
}