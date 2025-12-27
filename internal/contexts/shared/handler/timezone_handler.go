package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/dto"
	"github.com/basilex/promenade/pkg/reference"
	"github.com/basilex/promenade/pkg/response"
)

// TimezoneHandler handles timezone-related HTTP requests
type TimezoneHandler struct {
	repo *reference.Repository
}

// NewTimezoneHandler creates a new timezone handler
func NewTimezoneHandler(repo *reference.Repository) *TimezoneHandler {
	return &TimezoneHandler{repo: repo}
}

// ListTimezones godoc
// @Summary List all timezones
// @Description Get list of all active IANA timezones
// @Tags Reference Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.TimezoneResponse}
// @Failure 500 {object} response.Response
// @Router /timezones [get]
func (h *TimezoneHandler) ListTimezones(c *gin.Context) {
	ctx := c.Request.Context()

	timezones, err := h.repo.ListTimezones(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch timezones")
		return
	}

	response.Success(c, dto.ToTimezoneResponses(timezones))
}

// GetTimezoneByName godoc
// @Summary Get timezone by IANA name
// @Description Get timezone details by IANA timezone name
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param name path string true "Timezone Name (IANA, e.g., America/New_York, Europe/Kyiv)"
// @Success 200 {object} response.Response{data=dto.TimezoneResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /timezones/{name} [get]
func (h *TimezoneHandler) GetTimezoneByName(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_NAME", "Timezone name is required")
		return
	}

	timezone, err := h.repo.GetTimezoneByName(ctx, name)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "TIMEZONE_NOT_FOUND", "Timezone not found")
		return
	}

	response.Success(c, dto.ToTimezoneResponse(timezone))
}
