package handler

import (
	"net/http"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

type CountryHandler struct {
	countryUseCase usecase.CountryUseCase
}

func NewCountryHandler(countryUseCase usecase.CountryUseCase) *CountryHandler {
	return &CountryHandler{
		countryUseCase: countryUseCase,
	}
}

// Create creates a new country
// @Summary Create country
// @Tags countries
// @Accept json
// @Produce json
// @Param request body dto.CreateCountryRequest true "Country data"
// @Success 201 {object} response.Response{data=dto.CountryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries [post]
func (h *CountryHandler) Create(c *gin.Context) {
	var req dto.CreateCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	country := &entity.Country{
		Name:   req.Name,
		Code:   req.Code,
		ISO2:   req.ISO2,
		ISO3:   req.ISO3,
		Region: req.Region,
	}

	if err := h.countryUseCase.Create(c.Request.Context(), country); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create country", err)
		return
	}

	resp := h.toCountryResponse(country, false)
	response.Success(c, http.StatusCreated, resp)
}

// GetByID retrieves a country by ID
// @Summary Get country by ID
// @Tags countries
// @Produce json
// @Param id path string true "Country ID"
// @Param with_currencies query bool false "Include currencies"
// @Success 200 {object} response.Response{data=dto.CountryResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id} [get]
func (h *CountryHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	withCurrencies := c.Query("with_currencies") == "true"

	country, err := h.countryUseCase.GetByID(c.Request.Context(), id, withCurrencies)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "country not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get country", err)
		return
	}

	resp := h.toCountryResponse(country, withCurrencies)
	response.Success(c, http.StatusOK, resp)
}

// GetByCode retrieves a country by code
// @Summary Get country by code
// @Tags countries
// @Produce json
// @Param code path string true "Country code (iso2, iso3, or code)"
// @Param with_currencies query bool false "Include currencies"
// @Success 200 {object} response.Response{data=dto.CountryResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/code/{code} [get]
func (h *CountryHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	withCurrencies := c.Query("with_currencies") == "true"

	country, err := h.countryUseCase.GetByCode(c.Request.Context(), code, withCurrencies)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "country not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get country", err)
		return
	}

	resp := h.toCountryResponse(country, withCurrencies)
	response.Success(c, http.StatusOK, resp)
}

// List retrieves all countries
// @Summary List countries
// @Tags countries
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param region query string false "Filter by region" Enums(north_america, south_america, western_europe, eastern_europe, asia, middle_east, africa, oceania)
// @Param with_currencies query bool false "Include currencies"
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]dto.CountryResponse}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries [get]
func (h *CountryHandler) List(c *gin.Context) {
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)
	region := c.Query("region")
	withCurrencies := c.Query("with_currencies") == "true"

	var countries []entity.Country
	var total int
	var err error

	if region != "" {
		// Validate region
		validRegions := map[string]bool{
			"north_america":  true,
			"south_america":  true,
			"western_europe": true,
			"eastern_europe": true,
			"asia":           true,
			"middle_east":    true,
			"africa":         true,
			"oceania":        true,
		}
		if !validRegions[region] {
			response.Error(c, http.StatusBadRequest, "invalid region value", nil)
			return
		}
		countries, total, err = h.countryUseCase.ListByRegion(c.Request.Context(), region, page, pageSize, withCurrencies)
	} else {
		countries, total, err = h.countryUseCase.List(c.Request.Context(), page, pageSize, withCurrencies)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list countries", err)
		return
	}

	items := make([]dto.CountryResponse, len(countries))
	for i, country := range countries {
		items[i] = *h.toCountryResponse(&country, withCurrencies)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	response.Success(c, http.StatusOK, resp)
}

// Update updates a country
// @Summary Update country
// @Tags countries
// @Accept json
// @Produce json
// @Param id path string true "Country ID"
// @Param request body dto.UpdateCountryRequest true "Country data"
// @Success 200 {object} response.Response{data=dto.CountryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id} [put]
func (h *CountryHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	var req dto.UpdateCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	country := &entity.Country{
		ID:     id,
		Name:   req.Name,
		Code:   req.Code,
		ISO2:   req.ISO2,
		ISO3:   req.ISO3,
		Region: req.Region,
	}

	if err := h.countryUseCase.Update(c.Request.Context(), country); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "country not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update country", err)
		return
	}

	resp := h.toCountryResponse(country, false)
	response.Success(c, http.StatusOK, resp)
}

// Delete deletes a country
// @Summary Delete country
// @Tags countries
// @Produce json
// @Param id path string true "Country ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id} [delete]
func (h *CountryHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	if err := h.countryUseCase.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "country not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete country", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "country deleted successfully"})
}

// AddCurrency adds a currency to a country
// @Summary Add currency to country
// @Tags countries
// @Accept json
// @Produce json
// @Param id path string true "Country ID"
// @Param request body dto.AddCurrencyToCountryRequest true "Currency data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id}/currencies [post]
func (h *CountryHandler) AddCurrency(c *gin.Context) {
	countryID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	var req dto.AddCurrencyToCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	currencyID, err := uuidv7.Parse(req.CurrencyID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	if err := h.countryUseCase.AddCurrency(c.Request.Context(), countryID, currencyID, req.IsPrimary); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to add currency to country", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "currency added to country successfully"})
}

// RemoveCurrency removes a currency from a country
// @Summary Remove currency from country
// @Tags countries
// @Produce json
// @Param id path string true "Country ID"
// @Param currency_id path string true "Currency ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id}/currencies/{currency_id} [delete]
func (h *CountryHandler) RemoveCurrency(c *gin.Context) {
	countryID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	currencyID, err := uuidv7.Parse(c.Param("currency_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	if err := h.countryUseCase.RemoveCurrency(c.Request.Context(), countryID, currencyID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to remove currency from country", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "currency removed from country successfully"})
}

// GetCurrencies retrieves all currencies for a country
// @Summary Get country currencies
// @Tags countries
// @Produce json
// @Param id path string true "Country ID"
// @Success 200 {object} response.Response{data=[]dto.CurrencyResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id}/currencies [get]
func (h *CountryHandler) GetCurrencies(c *gin.Context) {
	countryID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	currencies, err := h.countryUseCase.GetCurrencies(c.Request.Context(), countryID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get currencies", err)
		return
	}

	items := make([]dto.CurrencyResponse, len(currencies))
	for i, currency := range currencies {
		items[i] = *h.toCurrencyResponse(&currency, false)
	}

	response.Success(c, http.StatusOK, items)
}

// Helper methods
func (h *CountryHandler) toCountryResponse(country *entity.Country, withCurrencies bool) *dto.CountryResponse {
	resp := &dto.CountryResponse{
		ID:        country.ID,
		Name:      country.Name,
		Code:      country.Code,
		ISO2:      country.ISO2,
		ISO3:      country.ISO3,
		Region:    country.Region,
		CreatedAt: country.CreatedAt,
		UpdatedAt: country.UpdatedAt,
	}

	if withCurrencies && len(country.Currencies) > 0 {
		resp.Currencies = make([]dto.CurrencyResponse, len(country.Currencies))
		for i, currency := range country.Currencies {
			resp.Currencies[i] = *h.toCurrencyResponse(&currency, false)
		}
	}

	return resp
}

func (h *CountryHandler) toCurrencyResponse(currency *entity.Currency, withCountries bool) *dto.CurrencyResponse {
	return &dto.CurrencyResponse{
		ID:        currency.ID,
		Name:      currency.Name,
		Code:      currency.Code,
		Symbol:    currency.Symbol,
		CreatedAt: currency.CreatedAt,
		UpdatedAt: currency.UpdatedAt,
	}
}
