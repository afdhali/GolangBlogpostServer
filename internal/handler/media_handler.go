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

type MediaHandler struct {
	mediaService service.MediaService
}

func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// Upload upload a new media file
func (h *MediaHandler) Upload(c *gin.Context) {
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

	// Get file from request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid file", err.Error())
		return
	}
	defer file.Close()

	// Parse request
	req := &dto.UploadMediaRequest{
		AltText:     c.PostForm("alt_text"),
		Description: c.PostForm("description"),
		IsFeatured:  c.PostForm("is_featured") == "true",
	}

	// Parse post_id if provided
	if postIDStr := c.PostForm("post_id"); postIDStr != "" {
		if postID, err := uuid.Parse(postIDStr); err == nil {
			req.PostID = &postID
		}
	}

	// Upload media
	media, err := h.mediaService.Upload(c.Request.Context(), user.ID, file, header, req)
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "invalid image" || err.Error() == "image size too large" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to upload media", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, media)
}

// GetAll get all media with pagination and filters
func (h *MediaHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	mediaType := c.DefaultQuery("media_type", "")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse optional UUID filters
	var postID, userID *uuid.UUID
	if postIDStr := c.Query("post_id"); postIDStr != "" {
		if pID, err := uuid.Parse(postIDStr); err == nil {
			postID = &pID
		}
	}
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uID, err := uuid.Parse(userIDStr); err == nil {
			userID = &uID
		}
	}

	// Parse is_featured filter
	var isFeatured *bool
	if isFeaturedStr := c.Query("is_featured"); isFeaturedStr != "" {
		if isFeaturedStr == "true" {
			isFeatured = &[]bool{true}[0]
		} else if isFeaturedStr == "false" {
			isFeatured = &[]bool{false}[0]
		}
	}

	params := &dto.MediaQueryParams{
		Page:       page,
		Limit:      limit,
		MediaType:  mediaType,
		PostID:     postID,
		UserID:     userID,
		IsFeatured: isFeatured,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
	}

	medias, total, err := h.mediaService.GetAll(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get medias", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, page, limit, total, medias)
}

// GetByID get media by ID
func (h *MediaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid media ID", err.Error())
		return
	}

	media, err := h.mediaService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Media not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, media)
}

// GetByPostID get all media for a post
// 👇 Note: Router uses :id parameter, NOT :postId (to avoid conflicts with /posts/:id)
func (h *MediaHandler) GetByPostID(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	medias, err := h.mediaService.GetByPostID(c.Request.Context(), postID)
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get medias", err.Error())
		return
	}

	response.Success(c, http.StatusOK, medias)
}

// GetFeaturedByPostID get featured media for a post
// 👇 Note: Router uses :id parameter, NOT :postId (to avoid conflicts with /posts/:id)
func (h *MediaHandler) GetFeaturedByPostID(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err.Error())
		return
	}

	media, err := h.mediaService.GetFeaturedByPostID(c.Request.Context(), postID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Featured media not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, media)
}

// Update update media metadata
func (h *MediaHandler) Update(c *gin.Context) {
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
		response.Error(c, http.StatusBadRequest, "Invalid media ID", err.Error())
		return
	}

	var req dto.UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	media, err := h.mediaService.Update(c.Request.Context(), id, &req, user)
	if err != nil {
		if err.Error() == "media not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "you don't have permission to update this media" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		if err.Error() == "post not found" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update media", err.Error())
		return
	}

	response.Success(c, http.StatusOK, media)
}

// Delete delete a media
func (h *MediaHandler) Delete(c *gin.Context) {
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
		response.Error(c, http.StatusBadRequest, "Invalid media ID", err.Error())
		return
	}

	err = h.mediaService.Delete(c.Request.Context(), id, user)
	if err != nil {
		if err.Error() == "media not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "you don't have permission to delete this media" {
			response.Error(c, http.StatusForbidden, "Forbidden", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete media", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Media deleted successfully"})
}