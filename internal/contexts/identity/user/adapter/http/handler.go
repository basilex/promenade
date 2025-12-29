package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	usecase       user.IUseCase
	jwtManager    *jwt.Manager
	tokenRevoker  *jwt.TokenRevoker
}

// NewUserHandler creates a new user handler
func NewUserHandler(usecase user.IUseCase, jwtManager *jwt.Manager, tokenRevoker *jwt.TokenRevoker) *UserHandler {
	return &UserHandler{
		usecase:      usecase,
		jwtManager:   jwtManager,
		tokenRevoker: tokenRevoker,
	}
}

// Register handles POST /users/register
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	created, err := h.usecase.Register(
		c.Request.Context(),
		req.Email,
		req.Name,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			response.BadRequest(c, "email already exists")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, ToUserResponse(created))
}

// Login handles POST /users/login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userEntity, err := h.usecase.Authenticate(
		c.Request.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid credentials")
			return
		}
		if errors.Is(err, user.ErrAccountLocked) {
			response.Unauthorized(c, "account is locked")
			return
		}
		if errors.Is(err, user.ErrAccountNotActive) {
			response.Unauthorized(c, "account is not active")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Generate JWT token pair
	tokenPair, err := h.jwtManager.GenerateTokenPair(
		userEntity.ID,
		userEntity.Email.Value(),
		userEntity.Roles, // Roles loaded from database via repository
	)
	if err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}

	response.Success(c, AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User:         ToUserResponse(userEntity),
	})
}

// RefreshToken handles POST /auth/refresh
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate and refresh the token
	tokenPair, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid or expired refresh token")
		return
	}

	// Get user info from token claims
	claims, err := h.jwtManager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		response.InternalError(c, "failed to extract user info")
		return
	}

	// Parse user ID from claims
	userID, err := uuidv7.Parse(claims.UserID)
	if err != nil {
		response.InternalError(c, "invalid user ID in token")
		return
	}

	// Get user from database
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.Unauthorized(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User:         ToUserResponse(userEntity),
	})
}

// GetByID handles GET /users/:id
// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// GetByEmail handles GET /users/email/:email
// @Summary Get user by email
// @Description Get user details by email address
// @Tags users
// @Accept json
// @Produce json
// @Param email path string true "User email"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/email/{email} [get]
func (h *UserHandler) GetByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		response.BadRequest(c, "email is required")
		return
	}

	userEntity, err := h.usecase.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// VerifyEmail handles POST /users/:id/verify-email
// @Summary Verify user email
// @Description Mark user email as verified
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/verify-email [post]
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	err = h.usecase.VerifyEmail(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Get updated user
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// ChangePassword handles PUT /users/:id/password
// @Summary Change user password
// @Description Change user password (requires old password)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.usecase.ChangePassword(
		c.Request.Context(),
		userUUID,
		req.OldPassword,
		req.NewPassword,
	)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		if errors.Is(err, user.ErrInvalidCredentials) {
			response.BadRequest(c, "old password is incorrect")
			return
		}
		if errors.Is(err, user.ErrInvalidPassword) {
			response.BadRequest(c, "new password does not meet requirements")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Get updated user
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// Suspend handles PUT /users/:id/suspend
// @Summary Suspend user account
// @Description Suspend user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/suspend [put]
func (h *UserHandler) Suspend(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	err = h.usecase.SuspendUser(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Get updated user
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// Ban handles PUT /users/:id/ban
// @Summary Ban user account
// @Description Ban user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/ban [put]
func (h *UserHandler) Ban(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	err = h.usecase.BanUser(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Get updated user
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// Activate handles PUT /users/:id/activate
// @Summary Activate user account
// @Description Activate suspended or banned user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/activate [put]
func (h *UserHandler) Activate(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	err = h.usecase.ActivateUser(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Get updated user
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// Unlock handles PUT /users/:id/unlock
// @Summary Unlock user account
// @Description Unlock account that was locked due to failed login attempts
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID v7)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/unlock [put]
func (h *UserHandler) Unlock(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	err = h.usecase.UnlockUser(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	// Get updated user
	userEntity, err := h.usecase.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToUserResponse(userEntity))
}

// List handles GET /users
// @Summary List users
// @Description List all users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} []UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	users, total, err := h.usecase.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   ToUserListResponse(users),
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// RevokeToken handles POST /auth/revoke
// @Summary Revoke current token
// @Description Revoke the current JWT access token (logout)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/revoke [post]
func (h *UserHandler) RevokeToken(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "missing authorization header")
		return
	}

	// Parse Bearer token
	parts := c.Request.Header.Get("Authorization")
	if len(parts) < 7 || parts[:7] != "Bearer " {
		response.Unauthorized(c, "invalid authorization header format")
		return
	}
	token := parts[7:]

	// Validate token to get expiration time
	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	// Revoke token
	if h.tokenRevoker != nil {
		err = h.tokenRevoker.Revoke(c.Request.Context(), token, claims.ExpiresAt.Time)
		if err != nil {
			response.InternalError(c, "failed to revoke token")
			return
		}
	}

	response.Success(c, gin.H{
		"message": "token revoked successfully",
	})
}

