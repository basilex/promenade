package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/country/dto"
	"github.com/basilex/promenade/internal/contexts/shared/country/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Handler handles country-related HTTP requests
type Handler struct {
	u usecase.ICountryUseCase
}

// NewHandler creates a new country handler
func NewHandler(u usecase.ICountryUseCase) *Handler {
	return &Handler{u: u}
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
func (h *Handler) ListCountries(c *gin.Context) {
	ctx := c.Request.Context()

	countries, err := h.u.List(ctx)
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
func (h *Handler) GetCountryByCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := strings.ToUpper(c.Param("code"))

	if len(code) != 2 {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Country code must be 2 characters (ISO 3166-1 alpha-2)")
		return
	}

	country, err := h.u.GetByCode(ctx, code)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "COUNTRY_NOT_FOUND", "Country not found")
		return
	}

	response.Success(c, dto.ToCountryResponse(country))
}

// CreateCountry godoc
// @Summary Create a new country
// @Description Create a new country with ISO codes
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param country body dto.CreateCountryRequest true "Country data"
// @Success 201 {object} response.Response{data=dto.CountryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries [post]
func (h *Handler) CreateCountry(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateCountryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	country, err := aggregate.NewCountry(req.Code, req.Name, req.PhoneCode)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}

	country.Code3 = req.Code3
	country.NumericCode = req.NumericCode
	country.NameLocal = req.NameLocal

	if err := h.u.Create(ctx, country); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create country")
		return
	}

	response.Created(c, dto.ToCountryResponse(country))
}

// UpdateCountry godoc
// @Summary Update a country
// @Description Update country details by ID
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param id path string true "Country ID (UUID)"
// @Param country body dto.UpdateCountryRequest true "Country data"
// @Success 200 {object} response.Response{data=dto.CountryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id} [put]
func (h *Handler) UpdateCountry(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var req dto.UpdateCountryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Parse UUID
	uuidID, err := uuidv7.Parse(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	// Get existing country
	existing, err := h.u.GetByID(ctx, uuidID)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "COUNTRY_NOT_FOUND", "Country not found")
		return
	}

	// Update fields
	existing.Code3 = req.Code3
	existing.NumericCode = req.NumericCode
	existing.Name = req.Name
	existing.NameLocal = req.NameLocal
	existing.PhoneCode = req.PhoneCode
	existing.IsActive = req.IsActive

	if err := h.u.Update(ctx, existing); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update country")
		return
	}

	response.Success(c, dto.ToCountryResponse(existing))
}

// DeleteCountry godoc
// @Summary Delete a country
// @Description Delete a country by ID
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param id path string true "Country ID (UUID)"
// @Success 204 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id} [delete]
func (h *Handler) DeleteCountry(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	uuidID, err := uuidv7.Parse(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	if err := h.u.Delete(ctx, uuidID); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete country")
		return
	}

	c.Status(http.StatusNoContent)
}
