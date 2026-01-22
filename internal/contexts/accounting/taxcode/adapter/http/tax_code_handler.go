package http

import (
	"errors"
	"strconv"

	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/dto"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

type TaxCodeHandler struct {
	uc usecase.ITaxCodeUseCase
}

func NewTaxCodeHandler(uc usecase.ITaxCodeUseCase) *TaxCodeHandler {
	return &TaxCodeHandler{uc: uc}
}

func (h *TaxCodeHandler) RegisterRoutes(router *gin.RouterGroup) {
	taxCodes := router.Group("/tax-codes")
	{
		taxCodes.POST("", h.CreateTaxCode)
		taxCodes.GET("/:id", h.GetTaxCode)
		taxCodes.GET("/code/:code", h.GetTaxCodeByCode)
		taxCodes.PUT("/:id", h.UpdateTaxCode)
		taxCodes.PUT("/:id/payable-account", h.SetTaxPayableAccount)
		taxCodes.PUT("/:id/receivable-account", h.SetTaxReceivableAccount)
		taxCodes.PUT("/:id/activate", h.ActivateTaxCode)
		taxCodes.PUT("/:id/deactivate", h.DeactivateTaxCode)
		taxCodes.DELETE("/:id", h.DeleteTaxCode)
		taxCodes.GET("", h.ListTaxCodes)
		taxCodes.GET("/type/:type", h.ListTaxCodesByType)
		taxCodes.GET("/active", h.ListActiveTaxCodes)
	}
}

// CreateTaxCode creates a new tax code
// @Summary Create a new tax code
// @Tags tax-codes
// @Accept json
// @Produce json
// @Param request body dto.CreateTaxCodeRequest true "Create tax code request"
// @Success 201 {object} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes [post]
func (h *TaxCodeHandler) CreateTaxCode(c *gin.Context) {
	var req dto.CreateTaxCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	taxType := aggregate.TaxType(req.TaxType)
	if !isValidTaxType(taxType) {
		response.BadRequest(c, "invalid tax type")
		return
	}

	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))
	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	tc, err := h.uc.CreateTaxCode(
		c.Request.Context(),
		organizationID,
		req.Code,
		req.Name,
		taxType,
		req.Rate,
		userUUID,
	)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Created(c, dto.ToTaxCodeResponse(tc))
}

// GetTaxCode gets a tax code by ID
// @Summary Get a tax code
// @Tags tax-codes
// @Produce json
// @Param id path string true "Tax Code ID"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id} [get]
func (h *TaxCodeHandler) GetTaxCode(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	tc, err := h.uc.GetTaxCodeByID(c.Request.Context(), id)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// GetTaxCodeByCode gets a tax code by code
// @Summary Get a tax code by code
// @Tags tax-codes
// @Produce json
// @Param code path string true "Tax code"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/code/{code} [get]
func (h *TaxCodeHandler) GetTaxCodeByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "code is required")
		return
	}

	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	tc, err := h.uc.GetTaxCodeByCode(c.Request.Context(), organizationID, code)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// UpdateTaxCode updates a tax code
// @Summary Update a tax code
// @Tags tax-codes
// @Accept json
// @Produce json
// @Param id path string true "Tax Code ID"
// @Param request body dto.UpdateTaxCodeRequest true "Update tax code request"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id} [put]
func (h *TaxCodeHandler) UpdateTaxCode(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	var req dto.UpdateTaxCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	tc, err := h.uc.UpdateTaxCode(c.Request.Context(), id, req.Rate, userUUID)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// SetTaxPayableAccount sets the tax payable account
// @Summary Set tax payable account
// @Tags tax-codes
// @Accept json
// @Produce json
// @Param id path string true "Tax Code ID"
// @Param request body dto.SetTaxAccountRequest true "Set account request"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id}/payable-account [put]
func (h *TaxCodeHandler) SetTaxPayableAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	var req dto.SetTaxAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accountID, err := uuidv7.Parse(req.AccountID)
	if err != nil {
		response.BadRequest(c, "invalid account_id")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	tc, err := h.uc.SetTaxPayableAccount(c.Request.Context(), id, accountID, userUUID)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// SetTaxReceivableAccount sets the tax receivable account
// @Summary Set tax receivable account
// @Tags tax-codes
// @Accept json
// @Produce json
// @Param id path string true "Tax Code ID"
// @Param request body dto.SetTaxAccountRequest true "Set account request"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id}/receivable-account [put]
func (h *TaxCodeHandler) SetTaxReceivableAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	var req dto.SetTaxAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accountID, err := uuidv7.Parse(req.AccountID)
	if err != nil {
		response.BadRequest(c, "invalid account_id")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	tc, err := h.uc.SetTaxReceivableAccount(c.Request.Context(), id, accountID, userUUID)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// ActivateTaxCode activates a tax code
// @Summary Activate a tax code
// @Tags tax-codes
// @Produce json
// @Param id path string true "Tax Code ID"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id}/activate [put]
func (h *TaxCodeHandler) ActivateTaxCode(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	tc, err := h.uc.ActivateTaxCode(c.Request.Context(), id, userUUID)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// DeactivateTaxCode deactivates a tax code
// @Summary Deactivate a tax code
// @Tags tax-codes
// @Produce json
// @Param id path string true "Tax Code ID"
// @Success 200 {object} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id}/deactivate [put]
func (h *TaxCodeHandler) DeactivateTaxCode(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	tc, err := h.uc.DeactivateTaxCode(c.Request.Context(), id, userUUID)
	if err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.Success(c, dto.ToTaxCodeResponse(tc))
}

// DeleteTaxCode deletes a tax code
// @Summary Delete a tax code
// @Tags tax-codes
// @Produce json
// @Param id path string true "Tax Code ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/{id} [delete]
func (h *TaxCodeHandler) DeleteTaxCode(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tax code ID")
		return
	}

	if err := h.uc.DeleteTaxCode(c.Request.Context(), id); err != nil {
		handleTaxCodeError(c, err)
		return
	}

	response.SuccessWithMessage(c, "tax code deleted successfully")
}

// ListTaxCodes lists all tax codes for an organization
// @Summary List tax codes
// @Tags tax-codes
// @Produce json
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} dto.TaxCodeResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes [get]
func (h *TaxCodeHandler) ListTaxCodes(c *gin.Context) {
	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	limit := 100
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	taxCodes, err := h.uc.ListTaxCodesByOrganization(c.Request.Context(), organizationID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list tax codes")
		return
	}

	response.Success(c, dto.ToTaxCodeResponseList(taxCodes))
}

// ListTaxCodesByType lists tax codes by type
// @Summary List tax codes by type
// @Tags tax-codes
// @Produce json
// @Param type path string true "Tax type"
// @Success 200 {array} dto.TaxCodeResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/type/{type} [get]
func (h *TaxCodeHandler) ListTaxCodesByType(c *gin.Context) {
	taxTypeStr := c.Param("type")
	taxType := aggregate.TaxType(taxTypeStr)
	if !isValidTaxType(taxType) {
		response.BadRequest(c, "invalid tax type")
		return
	}

	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	taxCodes, err := h.uc.ListTaxCodesByType(c.Request.Context(), organizationID, taxType)
	if err != nil {
		response.InternalError(c, "failed to list tax codes by type")
		return
	}

	response.Success(c, dto.ToTaxCodeResponseList(taxCodes))
}

// ListActiveTaxCodes lists all active tax codes
// @Summary List active tax codes
// @Tags tax-codes
// @Produce json
// @Success 200 {array} dto.TaxCodeResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/tax-codes/active [get]
func (h *TaxCodeHandler) ListActiveTaxCodes(c *gin.Context) {
	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	taxCodes, err := h.uc.ListActiveTaxCodes(c.Request.Context(), organizationID)
	if err != nil {
		response.InternalError(c, "failed to list active tax codes")
		return
	}

	response.Success(c, dto.ToTaxCodeResponseList(taxCodes))
}

func handleTaxCodeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, taxcode.ErrTaxCodeNotFound):
		response.NotFound(c, "tax code not found")
	case errors.Is(err, taxcode.ErrTaxCodeEmpty),
		errors.Is(err, taxcode.ErrTaxNameEmpty),
		errors.Is(err, taxcode.ErrInvalidTaxType),
		errors.Is(err, taxcode.ErrInvalidTaxRate),
		errors.Is(err, taxcode.ErrGLAccountRequired),
		errors.Is(err, taxcode.ErrTaxCodeInactive),
		errors.Is(err, taxcode.ErrTaxCodeDuplicate),
		errors.Is(err, taxcode.ErrCannotDeleteInUse),
		errors.Is(err, taxcode.ErrInvalidTaxableBase):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "operation failed")
	}
}

func isValidTaxType(t aggregate.TaxType) bool {
	switch t {
	case aggregate.TaxTypeVAT,
		aggregate.TaxTypeIncomeTax,
		aggregate.TaxTypePayrollTax,
		aggregate.TaxTypeWithholding,
		aggregate.TaxTypeExcise,
		aggregate.TaxTypeCustoms,
		aggregate.TaxTypeProperty,
		aggregate.TaxTypeOther:
		return true
	default:
		return false
	}
}
