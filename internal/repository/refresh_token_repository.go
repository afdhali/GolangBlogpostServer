package repository

import (
	"context"
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error)
    Revoke(ctx context.Context, token string) error
    RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
    DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
    db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
    return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
    return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
    var refreshToken entity.RefreshToken
    err := r.db.WithContext(ctx).
        Preload("User").
        Where("token = ?", token).
        First(&refreshToken).Error
    if err != nil {
        return nil, err
    }
    return &refreshToken, nil
}

func (r *refreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error) {
    var tokens []*entity.RefreshToken
    err := r.db.WithContext(ctx).
        Where("user_id = ? AND is_revoked = ?", userID, false).
        Find(&tokens).Error
    return tokens, err
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
    return r.db.WithContext(ctx).
        Model(&entity.RefreshToken{}).
        Where("token = ?", token).
        Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&entity.RefreshToken{}).
        Where("user_id = ?", userID).
        Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
    return r.db.WithContext(ctx).
        Where("expires_at < ?", time.Now()).
        Delete(&entity.RefreshToken{}).Error
}
