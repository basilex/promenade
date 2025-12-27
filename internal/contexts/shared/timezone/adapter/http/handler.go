package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/pkg/response"
)

// Handler handles timezone-related HTTP requests
type Handler struct {
	usecase timezone.IUseCase
}

// NewHandler creates a new timezone handler
func NewHandler(uc timezone.IUseCase) *Handler {
	return &Handler{usecase: uc}
}

// parseUTCOffset converts UTC offset string (e.g., "+02:00", "-05:30") to seconds
func parseUTCOffset(offset string) (int, error) {
	offset = strings.TrimSpace(offset)
	if len(offset) < 3 {
		return 0, strconv.ErrSyntax
	}
	
	// Validate sign
	if offset[0] != '+' && offset[0] != '-' {
		return 0, fmt.Errorf("offset must start with + or -")
	}
	
	// Parse sign
	sign := 1
	if offset[0] == '-' {
		sign = -1
	}
	offset = strings.TrimLeft(offset, "+-")
	
	// Split hours and minutes
	parts := strings.Split(offset, ":")
	if len(parts) != 2 {
		return 0, strconv.ErrSyntax
	}
	
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	
	// Validate ranges
	if hours < 0 || hours > 14 {
		return 0, fmt.Errorf("hours must be between 0 and 14")
	}
	
	if minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("minutes must be between 0 and 59")
	}
	
	return sign * (hours*3600 + minutes*60), nil
}

// ListTimezones godoc
// @Summary List all timezones
// @Description Get list of all active timezones with IANA names
// @Tags Reference Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]TimezoneResponse}
// @Failure 500 {object} response.Response
// @Router /timezones [get]
func (h *Handler) ListTimezones(c *gin.Context) {
	ctx := c.Request.Context()

	timezones, err := h.usecase.List(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch timezones")
		return
	}

	response.Success(c, ToTimezoneResponses(timezones))
}

// GetTimezoneByName godoc
// @Summary Get timezone by IANA name
// @Description Get timezone details by IANA timezone name
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param name path string true "Timezone Name (IANA, e.g., Europe/Kyiv, America/New_York)"
// @Success 200 {object} response.Response{data=TimezoneResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /timezones/{name} [get]
func (h *Handler) GetTimezoneByName(c *gin.Context) {
	ctx := c.Request.Context()
	// *name parameter includes leading slash, so trim it
	name := strings.TrimSpace(strings.TrimPrefix(c.Param("name"), "/"))

	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_NAME", "Timezone name is required")
		return
	}

	timezone, err := h.usecase.GetByName(ctx, name)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "TIMEZONE_NOT_FOUND", "Timezone not found")
		return
	}

	response.Success(c, ToTimezoneResponse(timezone))
}
// CreateTimezone creates a new timezone
func (h *Handler) CreateTimezone(c *gin.Context) {
	ctx := c.Request.Context()
	var req CreateTimezoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	utcOffset, err := parseUTCOffset(req.UTCOffset)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_UTC_OFFSET", "UTC offset must be in format +HH:MM or -HH:MM")
		return
	}

	timezone, err := timezone.NewTimezone(req.Name, req.Abbreviation, utcOffset)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}

	if err := h.usecase.Create(ctx, timezone); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create timezone")
		return
	}

	response.Created(c, ToTimezoneResponse(timezone))
}

// UpdateTimezone updates an existing timezone
func (h *Handler) UpdateTimezone(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var req UpdateTimezoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	uuidID, err := timezone.ParseUUID(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	existing, err := h.usecase.GetByID(ctx, uuidID)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "TIMEZONE_NOT_FOUND", "Timezone not found")
		return
	}

	utcOffset, err := parseUTCOffset(req.UTCOffset)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_UTC_OFFSET", "UTC offset must be in format +HH:MM or -HH:MM")
		return
	}

	existing.Abbreviation = req.Abbreviation
	existing.UTCOffset = utcOffset
	existing.IsActive = req.IsActive

	if err := h.usecase.Update(ctx, existing); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update timezone")
		return
	}

	response.Success(c, ToTimezoneResponse(existing))
}

// DeleteTimezone deletes a timezone
func (h *Handler) DeleteTimezone(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	uuidID, err := timezone.ParseUUID(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	if err := h.usecase.Delete(ctx, uuidID); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete timezone")
		return
	}

	c.Status(http.StatusNoContent)
}