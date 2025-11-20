package entity

import (
	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeDocument MediaType = "document"
	MediaTypeAudio    MediaType = "audio"
)

type Media struct {
	BaseEntity
	Filename    string     `gorm:"type:varchar(255);not null" json:"filename"`
	OriginalName string    `gorm:"type:varchar(255);not null" json:"original_name"`
	MimeType    string     `gorm:"type:varchar(100);not null" json:"mime_type"`
	Path        string     `gorm:"type:varchar(500);not null;index" json:"path"`
	URL         string     `gorm:"type:varchar(500);not null" json:"url"`
	Size        int64      `gorm:"not null" json:"size"`
	Width       *int       `gorm:"type:integer" json:"width,omitempty"`
	Height      *int       `gorm:"type:integer" json:"height,omitempty"`
	AltText     string     `gorm:"type:varchar(500)" json:"alt_text,omitempty"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	MediaType   MediaType  `gorm:"type:varchar(50);not null;default:'image'" json:"media_type"`
	PostID      *uuid.UUID `gorm:"type:uuid;index" json:"post_id,omitempty"`
	Post        *Post      `gorm:"foreignKey:PostID" json:"post,omitempty"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IsFeatured  bool       `gorm:"default:false" json:"is_featured"`
}

func (Media) TableName() string {
	return "media"
}

func (m *Media) IsImage() bool {
	return m.MediaType == MediaTypeImage
}

func (m *Media) IsVideo() bool {
	return m.MediaType == MediaTypeVideo
}

func (m *Media) GetDimensions() (width, height int, ok bool) {
	if m.Width != nil && m.Height != nil {
		return *m.Width, *m.Height, true
	}
	return 0, 0, false
}

func (m *Media) SetDimensions(width, height int) {
	m.Width = &width
	m.Height = &height
}