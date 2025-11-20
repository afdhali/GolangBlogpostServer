package dto

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
)

type UserResponse struct {
    ID        uuid.UUID `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    FullName  string    `json:"full_name"`
    Avatar    string    `json:"avatar,omitempty"`
    Role      string    `json:"role"`
    IsActive  bool      `json:"is_active"`
    PostCount int64     `json:"post_count"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse struct {
    ID        uuid.UUID `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    FullName  string    `json:"full_name"`
    Avatar    string    `json:"avatar,omitempty"`
    Role      string    `json:"role"`
    IsActive  bool      `json:"is_active"`
    PostCount int64     `json:"post_count"`
    CreatedAt time.Time `json:"created_at"`
}

type UserAuthor struct {
    ID       uuid.UUID `json:"id"`
    Username string    `json:"username"`
    FullName string    `json:"full_name"`
    Avatar   string    `json:"avatar,omitempty"`
}

// Converter functions
func ToUserResponse(user *entity.User, postCount int64) *UserResponse {
    return &UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        FullName:  user.FullName,
        Avatar:    user.Avatar,
        Role:      string(user.Role),
        IsActive:  user.IsActive,
        PostCount: postCount,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
}

func ToUserListResponse(user *entity.User, postCount int64) *UserListResponse {
    return &UserListResponse{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        FullName:  user.FullName,
        Avatar:    user.Avatar,
        Role:      string(user.Role),
        IsActive:  user.IsActive,
        PostCount: postCount,
        CreatedAt: user.CreatedAt,
    }
}

func ToUserAuthor(user *entity.User) *UserAuthor {
    return &UserAuthor{
        ID:       user.ID,
        Username: user.Username,
        FullName: user.FullName,
        Avatar:   user.Avatar,
    }
}
