package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	user, err := h.authUseCase.Register(c.Request.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, "email already exists", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to register user", err)
		return
	}

	resp := &dto.RegisterResponse{
		User:    dto.ToUserResponse(user),
		Message: "Registration successful. Please verify your email.",
	}

	response.Success(c, http.StatusCreated, resp)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get client info
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	accessToken, refreshToken, user, err := h.authUseCase.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
		userAgent,
		ipAddress,
	)

	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "invalid email or password", err)
			return
		}
		if errors.Is(err, usecase.ErrUserSuspended) {
			response.Error(c, http.StatusForbidden, "user account is suspended", err)
			return
		}
		if errors.Is(err, usecase.ErrUserBanned) {
			response.Error(c, http.StatusForbidden, "user account is banned", err)
			return
		}
		if errors.Is(err, usecase.ErrUserNotActive) {
			response.Error(c, http.StatusForbidden, "user account is not active", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to login", err)
		return
	}

	resp := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User:         dto.ToUserResponse(user),
	}

	response.Success(c, http.StatusOK, resp)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate refresh token and end session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Logout request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := h.authUseCase.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) {
			response.Error(c, http.StatusUnauthorized, "invalid refresh token", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to logout", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "logout successful"})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access and refresh tokens using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} dto.RefreshTokenResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) {
			response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token", err)
			return
		}
		if errors.Is(err, usecase.ErrUserNotActive) {
			response.Error(c, http.StatusForbidden, "user account is not active", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to refresh token", err)
		return
	}

	resp := &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
	}

	response.Success(c, http.StatusOK, resp)
}

// GetMe godoc
// @Summary Get current user
// @Description Get authenticated user profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userID, ok := userIDVal.(uuidv7.UUID)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid user id format", nil)
		return
	}

	user, err := h.authUseCase.GetMe(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "user not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get user", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserResponse(user))
}

// GetUserSessions godoc
// @Summary Get user sessions
// @Description Get all active sessions for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []dto.SessionResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/sessions [get]
func (h *AuthHandler) GetUserSessions(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	userID, ok := userIDVal.(uuidv7.UUID)
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid user id format", nil)
		return
	}

	sessions, err := h.authUseCase.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get sessions", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToSessionResponses(sessions))
}

// SuspendUser godoc
// @Summary Suspend user
// @Description Suspend a user account temporarily (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.SuspendUserRequest true "Suspend user request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/users/{id}/suspend [post]
func (h *AuthHandler) SuspendUser(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id", err)
		return
	}

	var req dto.SuspendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err = h.authUseCase.SuspendUser(c.Request.Context(), userID, req.Reason, req.Until)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "user not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to suspend user", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "user suspended successfully"})
}

// BanUser godoc
// @Summary Ban user
// @Description Ban a user account permanently (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.BanUserRequest true "Ban user request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/users/{id}/ban [post]
func (h *AuthHandler) BanUser(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id", err)
		return
	}

	var req dto.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err = h.authUseCase.BanUser(c.Request.Context(), userID, req.Reason)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "user not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to ban user", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "user banned successfully"})
}

// ReactivateUser godoc
// @Summary Reactivate user
// @Description Reactivate a suspended or inactive user account (admin only)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/users/{id}/reactivate [post]
func (h *AuthHandler) ReactivateUser(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id", err)
		return
	}

	err = h.authUseCase.ReactivateUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "user not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to reactivate user", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "user reactivated successfully"})
}
