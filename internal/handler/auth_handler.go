package handler

import (
	"errors"
	"net/http"

	"SHIBA/internal/domain"
	"SHIBA/internal/dto"
	"SHIBA/internal/middleware"
	"SHIBA/internal/repository"
	"SHIBA/internal/service"
	"SHIBA/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, domain.UserRole(req.Role))
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			response.Conflict(c, "email already registered")
			return
		}
		response.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registered successfully, please verify your email",
		"user_id": user.ID,
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), token); err != nil {
		if errors.Is(err, service.ErrTokenExpired) {
			response.BadRequest(c, "token is invalid or expired")
			return
		}
		response.InternalError(c)
		return
	}

	response.Message(c, http.StatusOK, "email verified successfully")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accessToken, refreshToken, user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Unauthorized(c, "invalid email or password")
		case errors.Is(err, service.ErrEmailNotVerified):
			response.Unauthorized(c, "email not verified")
		case errors.Is(err, service.ErrAccountBlocked):
			response.Unauthorized(c, "account is blocked")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         string(user.Role),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid or expired refresh token")
		return
	}

	response.OK(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		response.InternalError(c)
		return
	}
	response.Message(c, http.StatusOK, "logged out")
}
