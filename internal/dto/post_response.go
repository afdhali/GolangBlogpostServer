package dto

import (
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/google/uuid"
)

type PostResponse struct {
    ID            uuid.UUID      `json:"id"`
    Title         string         `json:"title"`
    Slug          string         `json:"slug"`
    Content       string         `json:"content"`
    Excerpt       string         `json:"excerpt"`
    FeaturedImage string         `json:"featured_image,omitempty"`
    Status        string         `json:"status"`
    Views         int64          `json:"views"`
    AuthorID      uuid.UUID      `json:"author_id"`
    Author        *PostAuthor    `json:"author"`
    CategoryID    uuid.UUID      `json:"category_id"`
    Category      *PostCategory  `json:"category"`
    Tags          []string       `json:"tags"`         // 👈 Changed to []string
    CommentCount  int64          `json:"comment_count"`
    PublishedAt   *time.Time     `json:"published_at,omitempty"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
}

type PostListResponse struct {
    ID            uuid.UUID     `json:"id"`
    Title         string        `json:"title"`
    Slug          string        `json:"slug"`
    Excerpt       string        `json:"excerpt"`
    FeaturedImage string        `json:"featured_image,omitempty"`
    Status        string        `json:"status"`
    Views         int64         `json:"views"`
    Author        *PostAuthor   `json:"author"`
    Category      *PostCategory `json:"category"`
    Tags          []string      `json:"tags"`         // 👈 Changed to []string
    CommentCount  int64         `json:"comment_count"`
    PublishedAt   *time.Time    `json:"published_at,omitempty"`
    CreatedAt     time.Time     `json:"created_at"`
    UpdatedAt     time.Time     `json:"updated_at"`
}

type PostAuthor struct {
    ID       uuid.UUID `json:"id"`
    Username string    `json:"username"`
    FullName string    `json:"full_name"`
    Avatar   string    `json:"avatar,omitempty"`
}

type PostCategory struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
    Slug string    `json:"slug"`
}

// 👇 ADD commentCount parameter
func ToPostResponse(post *entity.Post, commentCount int64) *PostResponse {
    response := &PostResponse{
        ID:            post.ID,
        Title:         post.Title,
        Slug:          post.Slug,
        Content:       post.Content,
        Excerpt:       post.Excerpt,
        FeaturedImage: post.FeaturedImage,
        Status:        string(post.Status),  // 👈 Convert PostStatus to string
        Views:         post.ViewCount,       // 👈 Use ViewCount from entity
        AuthorID:      post.AuthorID,
        CategoryID:    post.CategoryID,
        CommentCount:  commentCount,         // 👈 From parameter
        PublishedAt:   post.PublishedAt,
        CreatedAt:     post.CreatedAt,
        UpdatedAt:     post.UpdatedAt,
    }

    // Add author
    if post.Author != nil {
        response.Author = &PostAuthor{
            ID:       post.Author.ID,
            Username: post.Author.Username,
            FullName: post.Author.FullName,
            Avatar:   post.Author.Avatar,
        }
    }

    // Add category
    if post.Category != nil {
        response.Category = &PostCategory{
            ID:   post.Category.ID,
            Name: post.Category.Name,
            Slug: post.Category.Slug,
        }
    }

    // Add tags (already []string in entity)
    if len(post.Tags) > 0 {
        response.Tags = post.Tags  // 👈 Direct assignment
    } else {
        response.Tags = []string{} // 👈 Empty array instead of nil
    }

    return response
}

// 👇 ADD commentCount parameter
func ToPostListResponse(post *entity.Post, commentCount int64) *PostListResponse {
    response := &PostListResponse{
        ID:            post.ID,
        Title:         post.Title,
        Slug:          post.Slug,
        Excerpt:       post.Excerpt,
        FeaturedImage: post.FeaturedImage,
        Status:        string(post.Status),  // 👈 Convert PostStatus to string
        Views:         post.ViewCount,       // 👈 Use ViewCount from entity
        CommentCount:  commentCount,         // 👈 From parameter
        PublishedAt:   post.PublishedAt,
        CreatedAt:     post.CreatedAt,
        UpdatedAt:     post.UpdatedAt,
    }

    // Add author
    if post.Author != nil {
        response.Author = &PostAuthor{
            ID:       post.Author.ID,
            Username: post.Author.Username,
            FullName: post.Author.FullName,
            Avatar:   post.Author.Avatar,
        }
    }

    // Add category
    if post.Category != nil {
        response.Category = &PostCategory{
            ID:   post.Category.ID,
            Name: post.Category.Name,
            Slug: post.Category.Slug,
        }
    }

    // Add tags (already []string in entity)
    if len(post.Tags) > 0 {
        response.Tags = post.Tags  // 👈 Direct assignment
    } else {
        response.Tags = []string{} // 👈 Empty array instead of nil
    }

    return response
}

// ❌ REMOVE bulk converter
// func ToPostListResponses(posts []*entity.Post) []*PostListResponse