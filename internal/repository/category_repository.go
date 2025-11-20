package repository

import (
	"context"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	FindAll(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]*entity.Category, int64, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 👇 Counting Posts by Category
    CountByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
    return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var category entity.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
    var category entity.Category
    err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error
    if err != nil {
        return nil, err
    }
    return &category, nil
}

func (r *categoryRepository) FindAll(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]*entity.Category, int64, error) {
	var categories []*entity.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Category{})

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	// Default sorting
	order := "name ASC"
	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "ASC" // Default to ASC if not provided
		}
		order = sortBy + " " + sortOrder
	}

	err := query.Offset(offset).Limit(limit).Order(order).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
    return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&entity.Category{}, id).Error
}

// 👇 NEW METHOD: Count posts for multiple categories
func (r *categoryRepository) CountByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
    if len(categoryIDs) == 0 {
        return make(map[uuid.UUID]int64), nil
    }

    type Result struct {
        CategoryID uuid.UUID
        Count      int64
    }

    var results []Result
    err := r.db.WithContext(ctx).
        Table("posts").
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