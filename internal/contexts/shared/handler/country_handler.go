package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/dto"
	"github.com/basilex/promenade/pkg/reference"
	"github.com/basilex/promenade/pkg/response"
)

// CountryHandler handles country-related HTTP requests
type CountryHandler struct {
	repo *reference.Repository
}

// NewCountryHandler creates a new country handler
func NewCountryHandler(repo *reference.Repository) *CountryHandler {
	return &CountryHandler{repo: repo}
}

// ListCountries godoc
// @Summary List all countries
// @Description Get list of all active countries with ISO codes
// @Tags Reference Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.CountryResponse}
// @Failure 500 {object} response.Response
// @Router /countries [get]
func (h *CountryHandler) ListCountries(c *gin.Context) {
	ctx := c.Request.Context()

	countries, err := h.repo.ListCountries(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch countries")
		return
	}

	response.Success(c, dto.ToCountryResponses(countries))
}

// GetCountryByCode godoc
// @Summary Get country by ISO code
// @Description Get country details by ISO 3166-1 alpha-2 code
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param code path string true "Country Code (ISO 3166-1 alpha-2, e.g., US, UA, DE)"
// @Success 200 {object} response.Response{data=dto.CountryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{code} [get]
func (h *CountryHandler) GetCountryByCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := strings.ToUpper(c.Param("code"))

	if len(code) != 2 {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Country code must be 2 characters (ISO 3166-1 alpha-2)")
		return
	}

	country, err := h.repo.GetCountryByCode(ctx, code)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "COUNTRY_NOT_FOUND", "Country not found")
		return
	}

	response.Success(c, dto.ToCountryResponse(country))
}
