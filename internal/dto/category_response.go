package dto

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
)

type CategoryResponse struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Slug        string    `json:"slug"`
    Description string    `json:"description"`
    PostCount   int64     `json:"post_count"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryListResponse struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Slug        string    `json:"slug"`
    Description string    `json:"description"`
    PostCount   int64     `json:"post_count"`
}

type CategoryBasic struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
    Slug string    `json:"slug"`
}

// Converter functions
func ToCategoryResponse(category *entity.Category, postCount int64) *CategoryResponse {
    return &CategoryResponse{
        ID:          category.ID,
        Name:        category.Name,
        Slug:        category.Slug,
        Description: category.Description,
        PostCount:   postCount,
        CreatedAt:   category.CreatedAt,
        UpdatedAt:   category.UpdatedAt,
    }
}

func ToCategoryListResponse(category *entity.Category, postCount int64) *CategoryListResponse {
    return &CategoryListResponse{
        ID:          category.ID,
        Name:        category.Name,
        Slug:        category.Slug,
        Description: category.Description,
        PostCount:   postCount,
    }
}

func ToCategoryBasic(category *entity.Category) *CategoryBasic {
    return &CategoryBasic{
        ID:   category.ID,
        Name: category.Name,
        Slug: category.Slug,
    }
}