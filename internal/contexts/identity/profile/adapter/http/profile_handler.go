package profileHTTP

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	profileerrors "github.com/basilex/promenade/internal/contexts/identity/profile"
	"github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/profile/dto"
	"github.com/basilex/promenade/internal/contexts/identity/profile/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ProfileHandler handles HTTP requests for profile operations
type ProfileHandler struct {
	usecase usecase.IProfileUseCase
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(uc usecase.IProfileUseCase) *ProfileHandler {
	return &ProfileHandler{
		usecase: uc,
	}
}

// Create handles POST /profiles
// @Summary Create new profile
// @Description Create a new profile for a user
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body dto.CreateProfileRequest true "Profile creation request"
// @Success 201 {object} dto.ProfileResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /profiles [post]
func (h *ProfileHandler) Create(c *gin.Context) {
	var req dto.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: Get user ID from JWT context
	// For now, expect it in the request or query
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userID, err := uuidv7.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	profileEntity, err := h.usecase.CreateProfile(c.Request.Context(), userID, req.DisplayName)
	if err != nil {
		response.InternalError(c, "Failed to create profile")
		return
	}

	response.Created(c, dto.ToProfileResponse(profileEntity))
}

// GetByID handles GET /profiles/:id
// @Summary Get profile by ID
// @Description Retrieve a profile by its ID
// @Tags profiles
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} dto.ProfileResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /profiles/{id} [get]
func (h *ProfileHandler) GetByID(c *gin.Context) {
	profileID := c.Param("id")
	if profileID == "" {
		response.BadRequest(c, "profile_id is required")
		return
	}

	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	profileEntity, err := h.usecase.GetProfile(c.Request.Context(), profileUUID)
	if err != nil {
		if errors.Is(err, profileerrors.ErrNotFound) {
			response.NotFound(c, "profile not found")
			return
		}
		response.InternalError(c, "Failed to retrieve profile")
		return
	}

	response.Success(c, dto.ToProfileResponse(profileEntity))
}

// GetByUserID handles GET /profiles/user/:user_id
// @Summary Get profile by user ID
// @Description Retrieve a profile by user ID
// @Tags profiles
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} dto.ProfileResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /profiles/user/{user_id} [get]
func (h *ProfileHandler) GetByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, "user_id is required")
		return
	}

	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	profileEntity, err := h.usecase.GetProfileByUserID(c.Request.Context(), userUUID)
	if err != nil {
		if errors.Is(err, profileerrors.ErrNotFound) {
			response.NotFound(c, "profile not found")
			return
		}
		response.InternalError(c, "Failed to retrieve profile")
		return
	}

	response.Success(c, dto.ToProfileResponse(profileEntity))
}

// UpdateDisplayName handles PUT /profiles/:id/display-name
func (h *ProfileHandler) UpdateDisplayName(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateDisplayNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateDisplayName(c.Request.Context(), profileUUID, req.DisplayName); err != nil {
		response.InternalError(c, "Failed to update display name")
		return
	}

	response.SuccessWithMessage(c, "display name updated successfully")
}

// UpdateBio handles PUT /profiles/:id/bio
func (h *ProfileHandler) UpdateBio(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateBioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateBio(c.Request.Context(), profileUUID, req.Bio); err != nil {
		response.InternalError(c, "Failed to update bio")
		return
	}

	response.SuccessWithMessage(c, "bio updated successfully")
}

// UpdateAvatar handles PUT /profiles/:id/avatar
func (h *ProfileHandler) UpdateAvatar(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateAvatar(c.Request.Context(), profileUUID, req.AvatarURL); err != nil {
		response.InternalError(c, "Failed to update avatar")
		return
	}

	response.SuccessWithMessage(c, "avatar updated successfully")
}

// UpdatePersonalInfo handles PUT /profiles/:id/personal-info
func (h *ProfileHandler) UpdatePersonalInfo(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdatePersonalInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdatePersonalInfo(c.Request.Context(), profileUUID, req.FirstName, req.LastName, req.MiddleName); err != nil {
		response.InternalError(c, "Failed to update personal info")
		return
	}

	response.SuccessWithMessage(c, "personal info updated successfully")
}

// UpdateGender handles PUT /profiles/:id/gender
func (h *ProfileHandler) UpdateGender(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateGenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateGender(c.Request.Context(), profileUUID, aggregate.Gender(req.Gender)); err != nil {
		response.InternalError(c, "Failed to update gender")
		return
	}

	response.SuccessWithMessage(c, "gender updated successfully")
}

// UpdateDateOfBirth handles PUT /profiles/:id/date-of-birth
func (h *ProfileHandler) UpdateDateOfBirth(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateDateOfBirthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateDateOfBirth(c.Request.Context(), profileUUID, req.DateOfBirth); err != nil {
		response.InternalError(c, "Failed to update date of birth")
		return
	}

	response.SuccessWithMessage(c, "date of birth updated successfully")
}

// UpdateLocalization handles PUT /profiles/:id/localization
func (h *ProfileHandler) UpdateLocalization(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateLocalizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateLocalization(c.Request.Context(), profileUUID, req.Timezone, req.Language, req.Country); err != nil {
		response.InternalError(c, "Failed to update localization")
		return
	}

	response.SuccessWithMessage(c, "localization updated successfully")
}

// UpdateSocialLinks handles PUT /profiles/:id/social-links
func (h *ProfileHandler) UpdateSocialLinks(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	var req dto.UpdateSocialLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateSocialLinks(c.Request.Context(), profileUUID,
		req.Website, req.LinkedIn, req.Twitter, req.GitHub, req.Facebook, req.Instagram); err != nil {
		response.InternalError(c, "Failed to update social links")
		return
	}

	response.SuccessWithMessage(c, "social links updated successfully")
}

// SetPublic handles PUT /profiles/:id/public
func (h *ProfileHandler) SetPublic(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	if err := h.usecase.SetPublic(c.Request.Context(), profileUUID); err != nil {
		response.InternalError(c, "Failed to set profile to public")
		return
	}

	response.SuccessWithMessage(c, "profile set to public")
}

// SetPrivate handles PUT /profiles/:id/private
func (h *ProfileHandler) SetPrivate(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	if err := h.usecase.SetPrivate(c.Request.Context(), profileUUID); err != nil {
		response.InternalError(c, "Failed to set profile to private")
		return
	}

	response.SuccessWithMessage(c, "profile set to private")
}

// Delete handles DELETE /profiles/:id
func (h *ProfileHandler) Delete(c *gin.Context) {
	profileID := c.Param("id")
	profileUUID, err := uuidv7.Parse(profileID)
	if err != nil {
		response.BadRequest(c, "invalid profile_id")
		return
	}

	if err := h.usecase.DeleteProfile(c.Request.Context(), profileUUID); err != nil {
		response.InternalError(c, "Failed to delete profile")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListPublic handles GET /profiles/public
// @Summary List public profiles
// @Description Retrieve a list of public profiles
// @Tags profiles
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} dto.ProfileResponse
// @Router /profiles/public [get]
func (h *ProfileHandler) ListPublic(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := parseIntWithDefault(l, 20); err == nil {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := parseIntWithDefault(o, 0); err == nil {
			offset = parsed
		}
	}

	profiles, err := h.usecase.ListPublicProfiles(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list profiles")
		return
	}

	response.Success(c, dto.ToProfileResponseList(profiles))
}

// Helper function to parse int with default
func parseIntWithDefault(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue, err
	}
	return result, nil
}
