package dto

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
)

type AuthResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    TokenType    string       `json:"token_type"`
    ExpiresIn    int64        `json:"expires_in"`
    User         *UserProfile `json:"user"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int64  `json:"expires_in"`
}

type UserProfile struct {
    ID        uuid.UUID `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    FullName  string    `json:"full_name"`
    Avatar    string    `json:"avatar,omitempty"`
    Role      string    `json:"role"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

// Converter functions
func ToAuthResponse(user *entity.User, accessToken, refreshToken string, expiresIn int64) *AuthResponse {
    return &AuthResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    expiresIn,
        User:         ToUserProfile(user),
    }
}

func ToTokenResponse(accessToken, refreshToken string, expiresIn int64) *TokenResponse {
    return &TokenResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    expiresIn,
    }
}

func ToUserProfile(user *entity.User) *UserProfile {
    return &UserProfile{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        FullName:  user.FullName,
        Avatar:    user.Avatar,
        Role:      string(user.Role),
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
    }
}