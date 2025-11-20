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

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func(h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(),id)
	if err != nil {
		response.Error(c, http.StatusNotFound,"User not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *UserHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page","1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit","10"))
	search := c.Query("search")
	role := c.Query("role")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := h.userService.GetAll(c.Request.Context(), page, limit, search, role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, page, limit, total, users)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err !=nil {
		response.Error(c, http.StatusBadRequest, "Failed to create user", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	currentUser, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found in context")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, &req, currentUser.(*entity.User))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Update failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	currentUser, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User not found in context")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	err = h.userService.Delete(c.Request.Context(), id, currentUser.(*entity.User))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Delete failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uuid.UUID),&req)
	if err !=nil {
		response.Error(c, http.StatusBadRequest, "Password change failed", err.Error())
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message":"Password chaged successfully"})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized","User ID not found")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err !=nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), &req)
	if err !=nil {
		response.Error(c, http.StatusInternalServerError, "Update profile failed", err.Error())
		return
	}
	response.Success(c, http.StatusOK, user)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid file", err.Error())
		return
	}
	defer file.Close()

	user, err := h.userService.UploadAvatar(c.Request.Context(), userID.(uuid.UUID), file, header)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Upload failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	user, err := h.userService.DeleteAvatar(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Delete failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}