package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type UserProfileHandler struct {
	profileUC usecase.UserProfileUseCase
}

func NewUserProfileHandler(profileUC usecase.UserProfileUseCase) *UserProfileHandler {
	return &UserProfileHandler{
		profileUC: profileUC,
	}
}

// CreateProfile godoc
// @Summary Create user profile
// @Description Create a new profile for the authenticated user
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body dto.CreateProfileRequest true "Profile data"
// @Success 201 {object} response.Response{data=dto.ProfileResponse} "Profile created"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 409 {object} response.Response "Profile already exists"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles [post]
func (h *UserProfileHandler) CreateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Bind request
	var req dto.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Convert to entity
	profile := req.ToEntity()

	// Create profile
	createdProfile, err := h.profileUC.CreateProfile(ctx, userID.(uuidv7.UUID), profile)
	if err != nil {
		if errors.Is(err, usecase.ErrProfileAlreadyExists) {
			response.Error(c, http.StatusConflict, "profile already exists", err)
			return
		}
		if errors.Is(err, usecase.ErrNicknameAlreadyTaken) {
			response.Error(c, http.StatusConflict, "nickname already taken", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidProfileData) {
			response.Error(c, http.StatusBadRequest, "invalid profile data", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create profile", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToProfileResponse(createdProfile))
}

// GetProfile godoc
// @Summary Get profile by ID
// @Description Get a user profile by ID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} response.Response{data=dto.ProfileResponse} "Profile found"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Profile not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profiles/{id} [get]
func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Get viewer ID (optional)
	var viewerID *uuidv7.UUID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(uuidv7.UUID)
		viewerID = &uid
	}

	// Get profile
	profile, err := h.profileUC.GetProfile(ctx, profileID, viewerID)
	if err != nil {
		if errors.Is(err, usecase.ErrProfileNotFound) {
			response.Error(c, http.StatusNotFound, "profile not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorizedProfileAccess) {
			response.Error(c, http.StatusForbidden, "forbidden", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get profile", err)
		return
	}

	// Increment profile views if viewer is not the owner
	if viewerID != nil {
		_ = h.profileUC.IncrementViews(ctx, profileID, viewerID)
	}

	response.Success(c, http.StatusOK, dto.ToProfileResponse(profile))
}

// GetMyProfile godoc
// @Summary Get my profile
// @Description Get the authenticated user's profile
// @Tags profiles
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.ProfileResponse} "Profile found"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "Profile not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/me [get]
func (h *UserProfileHandler) GetMyProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	uid := userID.(uuidv7.UUID)

	// Get profile
	profile, err := h.profileUC.GetProfileByUserID(ctx, uid, &uid)
	if err != nil {
		if errors.Is(err, usecase.ErrProfileNotFound) {
			response.Error(c, http.StatusNotFound, "profile not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get profile", err)
		return
	}

	// Update last seen
	_ = h.profileUC.UpdateLastSeen(ctx, profile.ID)

	response.Success(c, http.StatusOK, dto.ToProfileResponse(profile))
}

// GetProfileByNickname godoc
// @Summary Get profile by nickname
// @Description Get a user profile by nickname
// @Tags profiles
// @Accept json
// @Produce json
// @Param nickname path string true "Profile nickname"
// @Success 200 {object} response.Response{data=dto.ProfileResponse} "Profile found"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Profile not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profiles/nickname/{nickname} [get]
func (h *UserProfileHandler) GetProfileByNickname(c *gin.Context) {
	ctx := c.Request.Context()

	nickname := c.Param("nickname")
	if nickname == "" {
		response.Error(c, http.StatusBadRequest, "nickname is required", nil)
		return
	}

	// Get viewer ID (optional)
	var viewerID *uuidv7.UUID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(uuidv7.UUID)
		viewerID = &uid
	}

	// Get profile
	profile, err := h.profileUC.GetProfileByNickname(ctx, nickname, viewerID)
	if err != nil {
		if errors.Is(err, usecase.ErrProfileNotFound) {
			response.Error(c, http.StatusNotFound, "profile not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorizedProfileAccess) {
			response.Error(c, http.StatusForbidden, "forbidden", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get profile", err)
		return
	}

	// Increment profile views if viewer is not the owner
	if viewerID != nil {
		_ = h.profileUC.IncrementViews(ctx, profile.ID, viewerID)
	}

	response.Success(c, http.StatusOK, dto.ToProfileResponse(profile))
}

// UpdateProfile godoc
// @Summary Update profile
// @Description Update the authenticated user's profile
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body dto.UpdateProfileRequest true "Updated profile data"
// @Success 200 {object} response.Response{data=dto.ProfileResponse} "Profile updated"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Profile not found"
// @Failure 409 {object} response.Response "Nickname already taken"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/{id} [put]
func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Bind request
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Convert to entity
	updates := req.ToEntity()

	// Update profile
	updatedProfile, err := h.profileUC.UpdateProfile(ctx, profileID, userID.(uuidv7.UUID), updates)
	if err != nil {
		if errors.Is(err, usecase.ErrProfileNotFound) {
			response.Error(c, http.StatusNotFound, "profile not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorizedProfileAccess) {
			response.Error(c, http.StatusForbidden, "forbidden", err)
			return
		}
		if errors.Is(err, usecase.ErrNicknameAlreadyTaken) {
			response.Error(c, http.StatusConflict, "nickname already taken", err)
			return
		}
		if errors.Is(err, usecase.ErrInvalidProfileData) {
			response.Error(c, http.StatusBadRequest, "invalid profile data", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update profile", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToProfileResponse(updatedProfile))
}

// DeleteProfile godoc
// @Summary Delete profile
// @Description Delete the authenticated user's profile
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Success 204 "Profile deleted"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden"
// @Failure 404 {object} response.Response "Profile not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/{id} [delete]
func (h *UserProfileHandler) DeleteProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Delete profile
	if err := h.profileUC.DeleteProfile(ctx, profileID, userID.(uuidv7.UUID)); err != nil {
		if errors.Is(err, usecase.ErrProfileNotFound) {
			response.Error(c, http.StatusNotFound, "profile not found", err)
			return
		}
		if errors.Is(err, usecase.ErrUnauthorizedProfileAccess) {
			response.Error(c, http.StatusForbidden, "forbidden", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete profile", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProfiles godoc
// @Summary List profiles
// @Description List user profiles with pagination
// @Tags profiles
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param public_only query bool false "Show only public profiles" default(true)
// @Success 200 {object} response.Response{data=dto.ProfileListResponse} "Profiles list"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profiles [get]
func (h *UserProfileHandler) ListProfiles(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination params
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	publicOnly := true
	if p := c.Query("public_only"); p != "" {
		if parsed, err := strconv.ParseBool(p); err == nil {
			publicOnly = parsed
		}
	}

	// List profiles
	profiles, err := h.profileUC.ListProfiles(ctx, limit, offset, publicOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list profiles", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToProfileListResponse(profiles, limit, offset))
}

// SearchProfiles godoc
// @Summary Search profiles
// @Description Search profiles by nickname, display name, or bio
// @Tags profiles
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.ProfileListResponse} "Search results"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /profiles/search [get]
func (h *UserProfileHandler) SearchProfiles(c *gin.Context) {
	ctx := c.Request.Context()

	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "search query is required", nil)
		return
	}

	// Parse pagination params
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Search profiles
	profiles, err := h.profileUC.SearchProfiles(ctx, query, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to search profiles", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToProfileListResponse(profiles, limit, offset))
}

// BanProfile godoc
// @Summary Ban profile
// @Description Ban a user profile (admin only)
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param request body dto.BanProfileRequest true "Ban reason"
// @Success 200 {object} response.Response "Profile banned"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/{id}/ban [post]
func (h *UserProfileHandler) BanProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (admin user)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Bind request
	var req dto.BanProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Ban profile
	if err := h.profileUC.BanProfile(ctx, profileID, req.Reason, userID.(uuidv7.UUID)); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to ban profile", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "profile banned successfully"})
}

// UnbanProfile godoc
// @Summary Unban profile
// @Description Unban a user profile (admin only)
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} response.Response "Profile unbanned"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/{id}/unban [post]
func (h *UserProfileHandler) UnbanProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Check authorization
	if _, exists := c.Get("user_id"); !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Unban profile
	if err := h.profileUC.UnbanProfile(ctx, profileID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to unban profile", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "profile unbanned successfully"})
}

// VerifyProfile godoc
// @Summary Verify profile
// @Description Verify a user profile (admin only)
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} response.Response "Profile verified"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/{id}/verify [post]
func (h *UserProfileHandler) VerifyProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Check authorization
	if _, exists := c.Get("user_id"); !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Verify profile
	if err := h.profileUC.VerifyProfile(ctx, profileID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to verify profile", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "profile verified successfully"})
}

// UnverifyProfile godoc
// @Summary Unverify profile
// @Description Unverify a user profile (admin only)
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} response.Response "Profile unverified"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /profiles/{id}/unverify [post]
func (h *UserProfileHandler) UnverifyProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Check authorization
	if _, exists := c.Get("user_id"); !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Parse profile ID
	profileID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid profile ID", err)
		return
	}

	// Unverify profile
	if err := h.profileUC.UnverifyProfile(ctx, profileID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to unverify profile", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "profile unverified successfully"})
}
