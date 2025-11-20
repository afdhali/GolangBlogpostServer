package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

type Post struct {
	BaseEntity
	Title         string         `gorm:"type:varchar(200);not null" json:"title"`
	Slug          string         `gorm:"type:varchar(200);uniqueIndex;not null" json:"slug"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	Excerpt       string         `gorm:"type:varchar(500)" json:"excerpt"`
	FeaturedImage string         `gorm:"type:varchar(500)" json:"featured_image,omitempty"`
	Tags          pq.StringArray `gorm:"type:text[]" json:"tags,omitempty"`
	Status        PostStatus     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	ViewCount     int64          `gorm:"default:0" json:"view_count"`
	AuthorID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"author_id"`
	Author        *User          `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	CategoryID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"category_id"`
	Category      *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Comments      []Comment      `gorm:"foreignKey:PostID" json:"comments,omitempty"`
	PublishedAt   *time.Time     `gorm:"index" json:"published_at,omitempty"`
}

func (Post) TableName() string {
	return "posts"
}

func (p *Post) IsPublished() bool {
	return p.Status == PostStatusPublished
}

func (p *Post) Publish() {
	p.Status = PostStatusPublished
	now := time.Now()
	p.PublishedAt = &now
}

func (p *Post) IncrementViewCount() {
	p.ViewCount++
}