package handler

import (
	"net/http"
	"strconv"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/service"
	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostHandler struct {
	postService service.PostService
}

func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

// GetAll get all posts with pagination and filters
func (h *PostHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.DefaultQuery("search", "")
	status := c.DefaultQuery("status", "")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse optional UUID filters
	var categoryID, authorID *uuid.UUID
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if cID, err := uuid.Parse(categoryIDStr); err == nil {
			categoryID = &cID
		}
	}
	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		if aID, err := uuid.Parse(authorIDStr); err == nil {
			authorID = &aID
		}
	}

	params := &dto.PostQueryParams{
		Page:       page,
		Limit:      limit,
		Search:     search,
		Status:     status,
		CategoryID: categoryID,
		AuthorID:   authorID,
		Tag:        c.DefaultQuery("tag", ""),
		SortBy:     sortBy,
		SortOrder:  sortOrder,
	}

	posts, total, err := h.postService.GetAll(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get posts", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, page, limit, total, posts)
}

// GetByID get post by ID
func (h *PostHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	post, err := h.postService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Post not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, post)
}

// GetBySlug get post by slug
func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	post, err := h.postService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Post not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, post)
}

// Create create a new post
func (h *PostHandler) Create(c *gin.Context) {
	// Get user from context
	userValue, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found")
		return
	}

	user, ok := userValue.(*entity.User)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Invalid user")
		return
	}

	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	post, err := h.postService.Create(c.Request.Context(), &req, user.ID)
	if err != nil {
		// Check error type for appropriate status code
		if err.Error() == "category not found" || err.Error() == "slug already exists" {
			response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create post", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, post)
}

// Update update an existing post
func (h *PostHandler) Update(c *gin.Context) {
	// Get user from context
	userValue, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found")
		return
	}

	user, ok := userValue.(*entity.User)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Invalid user")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	post, err := h.postService.Update(c.Request.Context(), id, &req, user)
	if err != nil {
		// Check permission errors
		if err.Error() == "you don't have permission to update this post" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		// Check not found errors
		if err.Error() == "post not found" || err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
		return
	}

	response.Success(c, http.StatusOK, post)
}

// Delete delete a post
func (h *PostHandler) Delete(c *gin.Context) {
	// Get user from context
	userValue, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found")
		return
	}

	user, ok := userValue.(*entity.User)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Invalid user")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	err = h.postService.Delete(c.Request.Context(), id, user)
	if err != nil {
		// Check permission errors
		if err.Error() == "you don't have permission to delete this post" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		// Check not found errors
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete post", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// Publish publish a post - only Super Admin and Admin can publish
func (h *PostHandler) Publish(c *gin.Context) {
	// Get user from context
	userValue, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found")
		return
	}

	user, ok := userValue.(*entity.User)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Invalid user")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	post, err := h.postService.Publish(c.Request.Context(), id, user)
	if err != nil {
		// Check permission errors
		if err.Error() == "you don't have permission to publish this post" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		// Check not found or state errors
		if err.Error() == "post not found" || err.Error() == "post is already published" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to publish post", err.Error())
		return
	}

	response.Success(c, http.StatusOK, post)
}

// Unpublish unpublish a post - only Super Admin and Admin can unpublish
func (h *PostHandler) Unpublish(c *gin.Context) {
	// Get user from context
	userValue, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found")
		return
	}

	user, ok := userValue.(*entity.User)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "Invalid user")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	post, err := h.postService.Unpublish(c.Request.Context(), id, user)
	if err != nil {
		// Check permission errors
		if err.Error() == "you don't have permission to unpublish this post" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		// Check not found or state errors
		if err.Error() == "post not found" || err.Error() == "post is not published" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to unpublish post", err.Error())
		return
	}

	response.Success(c, http.StatusOK, post)
}

// IncrementViews increment post view count
func (h *PostHandler) IncrementViews(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	err = h.postService.IncrementViews(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Post not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "View count incremented"})
}