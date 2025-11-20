package repository

import (
	"context"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)
	FindByPostID(ctx context.Context, postID uuid.UUID, page, limit int) ([]*entity.Comment, int64, error)
    Update(ctx context.Context, comment *entity.Comment) error
    Delete(ctx context.Context, id uuid.UUID) error

    // Counting by Post
    CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
    CountByPostIDs(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

type commentRepository struct {
    db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
    return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *entity.Comment) error {
    return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
    var comment entity.Comment
    err := r.db.WithContext(ctx).
        Preload("User").
        Preload("Post").
        Where("id = ?", id).
        First(&comment).Error
    if err != nil {
        return nil, err
    }
    return &comment, nil
}

func (r *commentRepository) FindByPostID(ctx context.Context, postID uuid.UUID, page, limit int) ([]*entity.Comment, int64, error) {
    var comments []*entity.Comment
    var total int64

    query := r.db.WithContext(ctx).Model(&entity.Comment{}).
        Preload("User").
        Preload("Replies.User").
        Where("post_id = ? AND parent_id IS NULL", postID)

    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    offset := (page - 1) * limit
    err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&comments).Error
    if err != nil {
        return nil, 0, err
    }

    return comments, total, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *entity.Comment) error {
    return r.db.WithContext(ctx).Save(comment).Error
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&entity.Comment{}, id).Error
}

// 👇 NEW: Count comments by single post
func (r *commentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&entity.Comment{}).
        Where("post_id = ?", postID).
        Count(&count).Error
    return count, err
}

// 👇 NEW: Count comments for multiple posts (bulk)
func (r *commentRepository) CountByPostIDs(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
    if len(postIDs) == 0 {
        return make(map[uuid.UUID]int64), nil
    }

    type Result struct {
        PostID uuid.UUID
        Count  int64
    }

    var results []Result
    err := r.db.WithContext(ctx).
        Model(&entity.Comment{}).
        Select("post_id, COUNT(*) as count").
        Where("post_id IN ?", postIDs).
        Group("post_id").
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    // Convert to map
    countMap := make(map[uuid.UUID]int64)
    for _, result := range results {
        countMap[result.PostID] = result.Count
    }

    // Fill zeros for posts without comments
    for _, postID := range postIDs {
        if _, exists := countMap[postID]; !exists {
            countMap[postID] = 0
        }
    }

    return countMap, nil
}