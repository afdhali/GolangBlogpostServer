package dto

import "github.com/google/uuid"

type CreateCommentRequest struct {
    Content  string     `json:"content" validate:"required,min=1,max=1000"`
    ParentID *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
}

type UpdateCommentRequest struct {
    Content string `json:"content" validate:"required,min=1,max=1000"`
}

type CommentQueryParams struct {
    Page      int    `form:"page" validate:"omitempty,min=1"`
    Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`
    SortBy    string `form:"sort_by" validate:"omitempty,oneof=created_at updated_at"`
    SortOrder string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}