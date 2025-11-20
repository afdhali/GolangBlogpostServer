package entity

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
    BaseEntity
    Token     string    `gorm:"type:varchar(500);uniqueIndex;not null" json:"token"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
    User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
    ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
    IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
}

func (RefreshToken) TableName() string {
    return "refresh_tokens"
}

func (rt *RefreshToken) IsExpired() bool {
    return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
    return !rt.IsRevoked && !rt.IsExpired()
}