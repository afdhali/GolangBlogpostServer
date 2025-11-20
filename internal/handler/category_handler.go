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

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// GetAll get all categories with pagination and filters
func (h *CategoryHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.DefaultQuery("search", "")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	params := &dto.CategoryQueryParams{
		Page:      page,
		Limit:     limit,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	categories, total, err := h.categoryService.GetAll(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get categories", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, page, limit, total, categories)
}

// GetByID get category by ID
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	category, err := h.categoryService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Category not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, category)
}

// GetBySlug get category by slug
func (h *CategoryHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.categoryService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Category not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, category)
}

// Create create a new category - admin only
func (h *CategoryHandler) Create(c *gin.Context) {
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

	// Check admin permission
	if !user.IsAdmin() {
		response.Error(c, http.StatusForbidden, "Forbidden", "Admin access required to create category")
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), &req, user)
	if err != nil {
		if err.Error() == "slug already exists" {
			response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, category)
}

// Update update an existing category - admin only
func (h *CategoryHandler) Update(c *gin.Context) {
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

	// Check admin permission
	if !user.IsAdmin() {
		response.Error(c, http.StatusForbidden, "Forbidden", "Admin access required to update category")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	category, err := h.categoryService.Update(c.Request.Context(), id, &req, user)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "slug already exists" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}

	response.Success(c, http.StatusOK, category)
}

// Delete delete a category - admin only
func (h *CategoryHandler) Delete(c *gin.Context) {
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

	// Check admin permission
	if !user.IsAdmin() {
		response.Error(c, http.StatusForbidden, "Forbidden", "Admin access required to delete category")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid category ID", err.Error())
		return
	}

	err = h.categoryService.Delete(c.Request.Context(), id, user)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, "Not found", err.Error())
			return
		}
		if err.Error() == "cannot delete category with posts" {
			response.Error(c, http.StatusBadRequest, "Bad request", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Category deleted successfully"})
}