package dto

import "github.com/google/uuid"

type UploadMediaRequest struct {
	AltText     string     `form:"alt_text" validate:"omitempty,max=500"`
	Description string     `form:"description" validate:"omitempty,max=2000"`
	PostID      *uuid.UUID `form:"post_id" validate:"omitempty,uuid"`
	IsFeatured  bool       `form:"is_featured"`
}

type UpdateMediaRequest struct {
	AltText     string     `json:"alt_text" validate:"omitempty,max=500"`
	Description string     `json:"description" validate:"omitempty,max=2000"`
	PostID      *uuid.UUID `json:"post_id" validate:"omitempty,uuid"`
	IsFeatured  *bool      `json:"is_featured" validate:"omitempty"`
}

type MediaQueryParams struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	Limit      int    `form:"limit" validate:"omitempty,min=1,max=100"`
	MediaType  string `form:"media_type" validate:"omitempty,oneof=image video document audio"`
	PostID     *uuid.UUID `form:"post_id" validate:"omitempty,uuid"`
	UserID     *uuid.UUID `form:"user_id" validate:"omitempty,uuid"`
	IsFeatured *bool  `form:"is_featured" validate:"omitempty"`
	SortBy     string `form:"sort_by" validate:"omitempty,oneof=created_at updated_at size"`
	SortOrder  string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}