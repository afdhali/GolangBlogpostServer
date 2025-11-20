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

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// GetByPostID get all comments for a post with pagination
func (h *CommentHandler) GetByPostID(c *gin.Context) {
	// Changed from "postId" to "id" to match the route parameter
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	params := &dto.CommentQueryParams{
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	comments, total, err := h.commentService.GetByPostID(c.Request.Context(), postID, params)
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get comments", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, page, limit, total, comments)
}

// Create create a new comment on a post
func (h *CommentHandler) Create(c *gin.Context) {
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

	// Changed from "postId" to "id" to match the route parameter
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	comment, err := h.commentService.Create(c.Request.Context(), postID, &req, user.ID)
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "parent comment not found" || err.Error() == "parent comment does not belong to this post" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create comment", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, comment)
}

// Update update an existing comment
func (h *CommentHandler) Update(c *gin.Context) {
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

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err.Error())
		return
	}

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	comment, err := h.commentService.Update(c.Request.Context(), commentID, &req, user)
	if err != nil {
		if err.Error() == "comment not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "you don't have permission to update this comment" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
		return
	}

	response.Success(c, http.StatusOK, comment)
}

// Delete delete a comment
func (h *CommentHandler) Delete(c *gin.Context) {
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

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err.Error())
		return
	}

	err = h.commentService.Delete(c.Request.Context(), commentID, user)
	if err != nil {
		if err.Error() == "comment not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "you don't have permission to delete this comment" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete comment", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}