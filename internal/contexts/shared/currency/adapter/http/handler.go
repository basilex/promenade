package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/pkg/response"
)

// Handler handles currency-related HTTP requests
type Handler struct {
	usecase currency.IUseCase
}

// NewHandler creates a new currency handler
func NewHandler(uc currency.IUseCase) *Handler {
	return &Handler{usecase: uc}
}

// ListCurrencies godoc
// @Summary List all currencies
// @Description Get list of all active currencies with ISO codes
// @Tags Reference Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]CurrencyResponse}
// @Failure 500 {object} response.Response
// @Router /currencies [get]
func (h *Handler) ListCurrencies(c *gin.Context) {
	ctx := c.Request.Context()

	currencies, err := h.usecase.List(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch currencies")
		return
	}

	response.Success(c, ToCurrencyResponses(currencies))
}

// GetCurrencyByCode godoc
// @Summary Get currency by ISO code
// @Description Get currency details by ISO 4217 code
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param code path string true "Currency Code (ISO 4217, e.g., USD, EUR, UAH)"
// @Success 200 {object} response.Response{data=CurrencyResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{code} [get]
func (h *Handler) GetCurrencyByCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := strings.ToUpper(c.Param("code"))

	if len(code) != 3 {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Currency code must be 3 characters (ISO 4217)")
		return
	}

	currency, err := h.usecase.GetByCode(ctx, code)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "CURRENCY_NOT_FOUND", "Currency not found")
		return
	}

	response.Success(c, ToCurrencyResponse(currency))
}
// CreateCurrency creates a new currency
func (h *Handler) CreateCurrency(c *gin.Context) {
	ctx := c.Request.Context()
	var req CreateCurrencyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	currency, err := currency.NewCurrency(req.Code, req.Name, req.Symbol, req.DecimalPlaces)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}

	currency.NumericCode = req.NumericCode

	if err := h.usecase.Create(ctx, currency); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create currency")
		return
	}

	response.Created(c, ToCurrencyResponse(currency))
}

// UpdateCurrency updates an existing currency
func (h *Handler) UpdateCurrency(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var req UpdateCurrencyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	uuidID, err := currency.ParseUUID(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	existing, err := h.usecase.GetByID(ctx, uuidID)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "CURRENCY_NOT_FOUND", "Currency not found")
		return
	}

	existing.NumericCode = req.NumericCode
	existing.Name = req.Name
	existing.Symbol = req.Symbol
	existing.DecimalPlaces = req.DecimalPlaces
	existing.IsActive = req.IsActive

	if err := h.usecase.Update(ctx, existing); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update currency")
		return
	}

	response.Success(c, ToCurrencyResponse(existing))
}

// DeleteCurrency deletes a currency
func (h *Handler) DeleteCurrency(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	uuidID, err := currency.ParseUUID(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	if err := h.usecase.Delete(ctx, uuidID); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete currency")
		return
	}

	c.Status(http.StatusNoContent)
}