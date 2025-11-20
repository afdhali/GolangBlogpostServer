package repository

import (
	"context"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(ctx context.Context, media *entity.Media) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Media, error)
	FindAll(ctx context.Context, page, limit int, mediaType string, postID, userID *uuid.UUID, isFeatured *bool, sortBy, sortOrder string) ([]*entity.Media, int64, error)
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Media, error)
	FindFeaturedByPostID(ctx context.Context, postID uuid.UUID) (*entity.Media, error)
	Update(ctx context.Context, media *entity.Media) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Bulk operations
	DeleteByPostID(ctx context.Context, postID uuid.UUID) error
	CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *entity.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Media, error) {
	var media entity.Media
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Post").
		Where("id = ?", id).
		First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) FindAll(ctx context.Context, page, limit int, mediaType string, postID, userID *uuid.UUID, isFeatured *bool, sortBy, sortOrder string) ([]*entity.Media, int64, error) {
	var medias []*entity.Media
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Media{}).
		Preload("User")

	// Apply filters
	if mediaType != "" {
		query = query.Where("media_type = ?", mediaType)
	}

	if postID != nil {
		query = query.Where("post_id = ?", *postID)
	}

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if isFeatured != nil {
		query = query.Where("is_featured = ?", *isFeatured)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	// Default sorting
	order := "created_at DESC"
	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "DESC"
		}
		order = sortBy + " " + sortOrder
	}

	err := query.Offset(offset).Limit(limit).Order(order).Find(&medias).Error
	if err != nil {
		return nil, 0, err
	}

	return medias, total, nil
}

func (r *mediaRepository) FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Media, error) {
	var medias []*entity.Media
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&medias).Error
	return medias, err
}

func (r *mediaRepository) FindFeaturedByPostID(ctx context.Context, postID uuid.UUID) (*entity.Media, error) {
	var media entity.Media
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ? AND is_featured = ?", postID, true).
		First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) Update(ctx context.Context, media *entity.Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Media{}, id).Error
}

func (r *mediaRepository) DeleteByPostID(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("post_id = ?", postID).Delete(&entity.Media{}).Error
}

func (r *mediaRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Media{}).
		Where("post_id = ?", postID).
		Count(&count).Error
	return count, err
}

func (r *mediaRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Media{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}