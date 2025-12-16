package handler

import (
	"net/http"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CurrencyHandler struct {
	currencyUseCase *usecase.CurrencyUseCase
}

func NewCurrencyHandler(currencyUseCase *usecase.CurrencyUseCase) *CurrencyHandler {
	return &CurrencyHandler{
		currencyUseCase: currencyUseCase,
	}
}

// Create creates a new currency
// @Summary Create currency
// @Tags currencies
// @Accept json
// @Produce json
// @Param request body dto.CreateCurrencyRequest true "Currency data"
// @Success 201 {object} response.Response{data=dto.CurrencyResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies [post]
func (h *CurrencyHandler) Create(c *gin.Context) {
	var req dto.CreateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	currency := &entity.Currency{
		Name:   req.Name,
		Code:   req.Code,
		Symbol: req.Symbol,
	}

	if err := h.currencyUseCase.Create(c.Request.Context(), currency); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create currency", err)
		return
	}

	resp := h.toCurrencyResponse(currency, false)
	response.Success(c, http.StatusCreated, resp)
}

// GetByID retrieves a currency by ID
// @Summary Get currency by ID
// @Tags currencies
// @Produce json
// @Param id path string true "Currency ID"
// @Param with_countries query bool false "Include countries"
// @Success 200 {object} response.Response{data=dto.CurrencyResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{id} [get]
func (h *CurrencyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	withCountries := c.Query("with_countries") == "true"

	currency, err := h.currencyUseCase.GetByID(c.Request.Context(), id, withCountries)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "currency not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get currency", err)
		return
	}

	resp := h.toCurrencyResponse(currency, withCountries)
	response.Success(c, http.StatusOK, resp)
}

// GetByCode retrieves a currency by code
// @Summary Get currency by code
// @Tags currencies
// @Produce json
// @Param code path string true "Currency code"
// @Param with_countries query bool false "Include countries"
// @Success 200 {object} response.Response{data=dto.CurrencyResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/code/{code} [get]
func (h *CurrencyHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	withCountries := c.Query("with_countries") == "true"

	currency, err := h.currencyUseCase.GetByCode(c.Request.Context(), code, withCountries)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "currency not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get currency", err)
		return
	}

	resp := h.toCurrencyResponse(currency, withCountries)
	response.Success(c, http.StatusOK, resp)
}

// List retrieves all currencies
// @Summary List currencies
// @Tags currencies
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param with_countries query bool false "Include countries"
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]dto.CurrencyResponse}}
// @Failure 500 {object} response.Response
// @Router /currencies [get]
func (h *CurrencyHandler) List(c *gin.Context) {
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)
	withCountries := c.Query("with_countries") == "true"

	currencies, total, err := h.currencyUseCase.List(c.Request.Context(), page, pageSize, withCountries)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list currencies", err)
		return
	}

	items := make([]dto.CurrencyResponse, len(currencies))
	for i, currency := range currencies {
		items[i] = *h.toCurrencyResponse(&currency, withCountries)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	response.Success(c, http.StatusOK, resp)
}

// Update updates a currency
// @Summary Update currency
// @Tags currencies
// @Accept json
// @Produce json
// @Param id path string true "Currency ID"
// @Param request body dto.UpdateCurrencyRequest true "Currency data"
// @Success 200 {object} response.Response{data=dto.CurrencyResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{id} [put]
func (h *CurrencyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	var req dto.UpdateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	currency := &entity.Currency{
		ID:     id,
		Name:   req.Name,
		Code:   req.Code,
		Symbol: req.Symbol,
	}

	if err := h.currencyUseCase.Update(c.Request.Context(), currency); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "currency not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update currency", err)
		return
	}

	resp := h.toCurrencyResponse(currency, false)
	response.Success(c, http.StatusOK, resp)
}

// Delete deletes a currency
// @Summary Delete currency
// @Tags currencies
// @Produce json
// @Param id path string true "Currency ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{id} [delete]
func (h *CurrencyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	if err := h.currencyUseCase.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "currency not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete currency", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "currency deleted successfully"})
}

// AddCountry adds a country to a currency
// @Summary Add country to currency
// @Tags currencies
// @Accept json
// @Produce json
// @Param id path string true "Currency ID"
// @Param request body dto.AddCountryToCurrencyRequest true "Country data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{id}/countries [post]
func (h *CurrencyHandler) AddCountry(c *gin.Context) {
	currencyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	var req dto.AddCountryToCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	countryID, err := uuid.Parse(req.CountryID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	if err := h.currencyUseCase.AddCountry(c.Request.Context(), currencyID, countryID, req.IsPrimary); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to add country to currency", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "country added to currency successfully"})
}

// RemoveCountry removes a country from a currency
// @Summary Remove country from currency
// @Tags currencies
// @Produce json
// @Param id path string true "Currency ID"
// @Param country_id path string true "Country ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{id}/countries/{country_id} [delete]
func (h *CurrencyHandler) RemoveCountry(c *gin.Context) {
	currencyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	countryID, err := uuid.Parse(c.Param("country_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	if err := h.currencyUseCase.RemoveCountry(c.Request.Context(), currencyID, countryID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to remove country from currency", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "country removed from currency successfully"})
}

// GetCountries retrieves all countries for a currency
// @Summary Get currency countries
// @Tags currencies
// @Produce json
// @Param id path string true "Currency ID"
// @Success 200 {object} response.Response{data=[]dto.CountryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{id}/countries [get]
func (h *CurrencyHandler) GetCountries(c *gin.Context) {
	currencyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid currency ID", err)
		return
	}

	countries, err := h.currencyUseCase.GetCountries(c.Request.Context(), currencyID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get countries", err)
		return
	}

	items := make([]dto.CountryResponse, len(countries))
	for i, country := range countries {
		items[i] = *h.toCountryResponse(&country, false)
	}

	response.Success(c, http.StatusOK, items)
}

// Helper methods
func (h *CurrencyHandler) toCurrencyResponse(currency *entity.Currency, withCountries bool) *dto.CurrencyResponse {
	resp := &dto.CurrencyResponse{
		ID:        currency.ID,
		Name:      currency.Name,
		Code:      currency.Code,
		Symbol:    currency.Symbol,
		CreatedAt: currency.CreatedAt,
		UpdatedAt: currency.UpdatedAt,
	}

	if withCountries && len(currency.Countries) > 0 {
		resp.Countries = make([]dto.CountryResponse, len(currency.Countries))
		for i, country := range currency.Countries {
			resp.Countries[i] = *h.toCountryResponse(&country, false)
		}
	}

	return resp
}

func (h *CurrencyHandler) toCountryResponse(country *entity.Country, withCurrencies bool) *dto.CountryResponse {
	return &dto.CountryResponse{
		ID:        country.ID,
		Name:      country.Name,
		Code:      country.Code,
		ISO2:      country.ISO2,
		ISO3:      country.ISO3,
		Region:    country.Region,
		CreatedAt: country.CreatedAt,
		UpdatedAt: country.UpdatedAt,
	}
}
