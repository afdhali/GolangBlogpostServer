package dto

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Slug        string `json:"slug" validate:"required,min=2,max=100,slug"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	Slug        string `json:"slug" validate:"omitempty,min=2,max=100,slug"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

type CategoryQueryParams struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Search    string `form:"search" validate:"omitempty,max=100"`
	SortBy    string `form:"sort_by" validate:"omitempty,oneof=created_at name post_count"`
	SortOrder string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}