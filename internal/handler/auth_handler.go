package handler

import (
	"net/http"

	"github.com/afdhali/GolangBlogpostServer/internal/dto"
	"github.com/afdhali/GolangBlogpostServer/internal/service"
	"github.com/afdhali/GolangBlogpostServer/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Registration failed", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid request", err.Error())
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Token refresh failed", err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)

}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err := h.authService.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Logout failed", err.Error)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Logout successful"})
}