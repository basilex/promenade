package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/usecase"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register godoc
// @Summary      Register new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Registration data"
// @Success      201 {object} response.Response{data=dto.AuthResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse "Email already exists"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	user, tokens, err := h.authUseCase.Register(c.Request.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if err == usecase.ErrEmailAlreadyExists {
			response.Error(c, http.StatusConflict, "email already exists", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to register", err)
		return
	}

	resp := dto.ToAuthResponse(user, tokens)
	response.Success(c, http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login
// @Description  Authenticate user and return JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200 {object} response.Response{data=dto.AuthResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse "Invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	user, tokens, err := h.authUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials || err == usecase.ErrUserNotActive {
			response.Error(c, http.StatusUnauthorized, "invalid credentials", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to login", err)
		return
	}

	resp := dto.ToAuthResponse(user, tokens)
	response.Success(c, http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} response.Response{data=dto.RefreshTokenResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse "Invalid token"
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	accessToken, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid refresh token", err)
		return
	}

	resp := dto.RefreshTokenResponse{
		AccessToken: accessToken,
	}
	response.Success(c, http.StatusOK, resp)
}

// GetMe godoc
// @Summary      Get current user
// @Description  Returns current authenticated user info
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.Response{data=dto.UserResponse}
// @Failure      401 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	user, err := h.authUseCase.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found", err)
		return
	}

	resp := dto.ToUserResponse(user)
	response.Success(c, http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout
// @Description  Logout user (client should delete tokens)
// @Tags         auth
// @Success      200 {object} response.Response
// @Security     BearerAuth
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// В stateless JWT logout обрабатывается на клиенте
	// Здесь можно добавить логику с blacklist токенов если нужно
	response.Success(c, http.StatusOK, gin.H{"message": "logged out successfully"})
}
