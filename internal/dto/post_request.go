package dto

import "github.com/google/uuid"

type CreatePostRequest struct {
    Title         string     `json:"title" validate:"required,min=5,max=200"`
    Slug          string     `json:"slug" validate:"required,min=5,max=200,slug"`
    Content       string     `json:"content" validate:"required,min=10"`
    Excerpt       string     `json:"excerpt" validate:"omitempty,max=500"`
    CategoryID    uuid.UUID  `json:"category_id" validate:"required,uuid"`
    FeaturedImage string     `json:"featured_image" validate:"omitempty,url"`
    Tags          []string   `json:"tags" validate:"omitempty,dive,min=2,max=50"`
    Status        string     `json:"status" validate:"omitempty,oneof=draft published archived"`
}

type UpdatePostRequest struct {
    Title         string     `json:"title" validate:"omitempty,min=5,max=200"`
    Slug          string     `json:"slug" validate:"omitempty,min=5,max=200,slug"`
    Content       string     `json:"content" validate:"omitempty,min=10"`
    Excerpt       string     `json:"excerpt" validate:"omitempty,max=500"`
    CategoryID    *uuid.UUID `json:"category_id" validate:"omitempty,uuid"`
    FeaturedImage string     `json:"featured_image" validate:"omitempty,url"`
    Tags          []string   `json:"tags" validate:"omitempty,dive,min=2,max=50"`
    Status        string     `json:"status" validate:"omitempty,oneof=draft published archived"`
}

type PostQueryParams struct {
    Page       int        `form:"page" validate:"omitempty,min=1"`
    Limit      int        `form:"limit" validate:"omitempty,min=1,max=100"`
    Search     string     `form:"search" validate:"omitempty,max=100"`
    Status     string     `form:"status" validate:"omitempty,oneof=draft published archived"`
    CategoryID *uuid.UUID `form:"category_id" validate:"omitempty,uuid"`
    Tag        string     `form:"tag" validate:"omitempty,max=50"`
    AuthorID   *uuid.UUID `form:"author_id" validate:"omitempty,uuid"`
    SortBy     string     `form:"sort_by" validate:"omitempty,oneof=created_at updated_at title views"`
    SortOrder  string     `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}