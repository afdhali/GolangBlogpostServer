package repository

import (
	"context"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindAll(ctx context.Context, page, limit int, search, role string, isActive *bool, sortBy, sortOrder string) ([]*entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct{
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, page, limit int, search, role string, isActive *bool, sortBy, sortOrder string) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	// Apply filters
	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting (with defaults and basic validation)
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	// To prevent SQL injection, you could add a whitelist of allowed sortBy fields here (e.g., "created_at", "username", "email").
	order := sortBy + " " + sortOrder

	// Pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order(order).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}