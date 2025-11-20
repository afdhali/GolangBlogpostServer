package dto

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
)

type MediaResponse struct {
	ID           uuid.UUID `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	URL          string    `json:"url"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	AltText      string    `json:"alt_text,omitempty"`
	Description  string    `json:"description,omitempty"`
	MediaType    string    `json:"media_type"`
	PostID       *uuid.UUID `json:"post_id,omitempty"`
	UserID       uuid.UUID `json:"user_id"`
	User         *MediaAuthor `json:"user,omitempty"`
	IsFeatured   bool      `json:"is_featured"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MediaListResponse struct {
	ID           uuid.UUID `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	URL          string    `json:"url"`
	Size         int64     `json:"size"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	MediaType    string    `json:"media_type"`
	PostID       *uuid.UUID `json:"post_id,omitempty"`
	User         *MediaAuthor  `json:"user,omitempty"`
	IsFeatured   bool      `json:"is_featured"`
	CreatedAt    time.Time `json:"created_at"`
}

type MediaAuthor struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	FullName string    `json:"full_name"`
	Avatar   string    `json:"avatar,omitempty"`
}

type MediaBasic struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	URL      string    `json:"url"`
	Size     int64     `json:"size"`
}

// Converter functions
func ToMediaResponse(media *entity.Media) *MediaResponse {
	response := &MediaResponse{
		ID:           media.ID,
		Filename:     media.Filename,
		OriginalName: media.OriginalName,
		MimeType:     media.MimeType,
		URL:          media.URL,
		Path:         media.Path,
		Size:         media.Size,
		Width:        media.Width,
		Height:       media.Height,
		AltText:      media.AltText,
		Description:  media.Description,
		MediaType:    string(media.MediaType),
		PostID:       media.PostID,
		UserID:       media.UserID,
		IsFeatured:   media.IsFeatured,
		CreatedAt:    media.CreatedAt,
		UpdatedAt:    media.UpdatedAt,
	}

	// Add user info if exists
	if media.User != nil {
		response.User = &MediaAuthor{
			ID:       media.User.ID,
			Username: media.User.Username,
			FullName: media.User.FullName,
			Avatar:   media.User.Avatar,
		}
	}

	return response
}

func ToMediaListResponse(media *entity.Media) *MediaListResponse {
	response := &MediaListResponse{
		ID:           media.ID,
		Filename:     media.Filename,
		OriginalName: media.OriginalName,
		MimeType:     media.MimeType,
		URL:          media.URL,
		Size:         media.Size,
		Width:        media.Width,
		Height:       media.Height,
		MediaType:    string(media.MediaType),
		PostID:       media.PostID,
		IsFeatured:   media.IsFeatured,
		CreatedAt:    media.CreatedAt,
	}

	// ✅ TAMBAHKAN INI - Include user info
	if media.User != nil {
		response.User = &MediaAuthor{
			ID:       media.User.ID,
			Username: media.User.Username,
			FullName: media.User.FullName,
			Avatar:   media.User.Avatar,
		}
	}

	return response
}

func ToMediaBasic(media *entity.Media) *MediaBasic {
	return &MediaBasic{
		ID:       media.ID,
		Filename: media.Filename,
		URL:      media.URL,
		Size:     media.Size,
	}
}

func ToMediaResponses(medias []*entity.Media) []*MediaResponse {
	responses := make([]*MediaResponse, len(medias))
	for i, media := range medias {
		responses[i] = ToMediaResponse(media)
	}
	return responses
}

func ToMediaListResponses(medias []*entity.Media) []*MediaListResponse {
	responses := make([]*MediaListResponse, len(medias))
	for i, media := range medias {
		responses[i] = ToMediaListResponse(media)
	}
	return responses
}