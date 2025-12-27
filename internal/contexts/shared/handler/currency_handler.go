package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/dto"
	"github.com/basilex/promenade/pkg/reference"
	"github.com/basilex/promenade/pkg/response"
)

// CurrencyHandler handles currency-related HTTP requests
type CurrencyHandler struct {
	repo *reference.Repository
}

// NewCurrencyHandler creates a new currency handler
func NewCurrencyHandler(repo *reference.Repository) *CurrencyHandler {
	return &CurrencyHandler{repo: repo}
}

// ListCurrencies godoc
// @Summary List all currencies
// @Description Get list of all active currencies with ISO codes
// @Tags Reference Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.CurrencyResponse}
// @Failure 500 {object} response.Response
// @Router /currencies [get]
func (h *CurrencyHandler) ListCurrencies(c *gin.Context) {
	ctx := c.Request.Context()

	currencies, err := h.repo.ListCurrencies(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch currencies")
		return
	}

	response.Success(c, dto.ToCurrencyResponses(currencies))
}

// GetCurrencyByCode godoc
// @Summary Get currency by ISO code
// @Description Get currency details by ISO 4217 alpha code
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param code path string true "Currency Code (ISO 4217, e.g., USD, EUR, UAH)"
// @Success 200 {object} response.Response{data=dto.CurrencyResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /currencies/{code} [get]
func (h *CurrencyHandler) GetCurrencyByCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := strings.ToUpper(c.Param("code"))

	if len(code) != 3 {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Currency code must be 3 characters (ISO 4217)")
		return
	}

	currency, err := h.repo.GetCurrencyByCode(ctx, code)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "CURRENCY_NOT_FOUND", "Currency not found")
		return
	}

	response.Success(c, dto.ToCurrencyResponse(currency))
}
