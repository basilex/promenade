package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	cashregistererrors "github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/dto"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CashRegisterHandler handles HTTP requests for cash register operations
type CashRegisterHandler struct {
	u usecase.ICashRegisterUseCase
}

// NewCashRegisterHandler creates a new cash register handler
func NewCashRegisterHandler(u usecase.ICashRegisterUseCase) *CashRegisterHandler {
	return &CashRegisterHandler{u: u}
}

// Create creates a new cash register
// @Summary Create cash register
// @Description Create a new fiscal cash register
// @Tags CashRegister
// @Accept json
// @Produce json
// @Param request body dto.CreateCashRegisterRequest true "Cash register creation request"
// @Success 201 {object} response.Response{data=dto.CashRegisterResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers [post]
func (h *CashRegisterHandler) Create(c *gin.Context) {
	var req dto.CreateCashRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse UUIDs
	organizationID, err := uuidv7.Parse(req.OrganizationID)
	if err != nil {
		response.BadRequest(c, "Invalid organization ID format")
		return
	}

	createdBy, err := uuidv7.Parse(req.CreatedBy)
	if err != nil {
		response.BadRequest(c, "Invalid created_by format")
		return
	}

	// Create cash register via use case
	cr, err := h.u.CreateCashRegister(c.Request.Context(), organizationID, req.FiscalNumber, req.Model, createdBy)
	if err != nil {
		// Map domain errors to user-friendly messages
		switch {
		case errors.Is(err, cashregistererrors.ErrOrganizationIDRequired):
			response.BadRequest(c, "Organization ID is required")
		case errors.Is(err, cashregistererrors.ErrFiscalNumberRequired):
			response.BadRequest(c, "Fiscal number is required")
		case errors.Is(err, cashregistererrors.ErrModelRequired):
			response.BadRequest(c, "Model is required")
		case errors.Is(err, cashregistererrors.ErrCashRegisterFiscalNumberExists):
			response.BadRequest(c, "Fiscal number already exists")
		default:
			// System errors - generic message (no detail leakage)
			response.InternalError(c, "Failed to create cash register")
		}
		return
	}

	response.Created(c, dto.ToCashRegisterResponse(cr))
}

// GetByID retrieves cash register by ID
// @Summary Get cash register by ID
// @Description Get cash register by ID
// @Tags CashRegister
// @Produce json
// @Param id path string true "Cash Register ID"
// @Success 200 {object} response.Response{data=dto.CashRegisterResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers/{id} [get]
func (h *CashRegisterHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cash register ID format")
		return
	}

	cr, err := h.u.GetCashRegister(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, cashregistererrors.ErrCashRegisterNotFound) {
			response.NotFound(c, "Cash register not found")
			return
		}
		response.InternalError(c, "Failed to retrieve cash register")
		return
	}

	response.Success(c, dto.ToCashRegisterResponse(cr))
}

// List retrieves all cash registers
// @Summary List cash registers
// @Description Get list of all cash registers
// @Tags CashRegister
// @Produce json
// @Param organization_id query string false "Organization ID"
// @Success 200 {object} response.Response{data=[]dto.CashRegisterResponse}
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers [get]
func (h *CashRegisterHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	// Optional organization filter
	var organizationID *uuidv7.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		orgID, err := uuidv7.Parse(orgIDStr)
		if err != nil {
			response.BadRequest(c, "Invalid organization ID format")
			return
		}
		organizationID = &orgID
	}

	cashRegisters, err := h.u.ListCashRegisters(ctx, organizationID)
	if err != nil {
		response.InternalError(c, "Failed to list cash registers")
		return
	}

	response.Success(c, dto.ToCashRegisterListResponse(cashRegisters))
}

// Activate activates a cash register
// @Summary Activate cash register
// @Description Activate a cash register with license key
// @Tags CashRegister
// @Accept json
// @Produce json
// @Param id path string true "Cash Register ID"
// @Param request body dto.ActivateCashRegisterRequest true "Activation request"
// @Success 200 {object} response.Response{data=dto.CashRegisterResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers/{id}/activate [post]
func (h *CashRegisterHandler) Activate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cash register ID format")
		return
	}

	var req dto.ActivateCashRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	activatedBy, err := uuidv7.Parse(req.ActivatedBy)
	if err != nil {
		response.BadRequest(c, "Invalid activated_by format")
		return
	}

	cr, err := h.u.ActivateCashRegister(c.Request.Context(), id, req.LicenseKey, activatedBy)
	if err != nil {
		switch {
		case errors.Is(err, cashregistererrors.ErrCashRegisterNotFound):
			response.NotFound(c, "Cash register not found")
		case errors.Is(err, cashregistererrors.ErrLicenseKeyRequired):
			response.BadRequest(c, "License key is required")
		case errors.Is(err, cashregistererrors.ErrCashRegisterAlreadyActive):
			response.BadRequest(c, "Cash register is already active")
		default:
			response.InternalError(c, "Failed to activate cash register")
		}
		return
	}

	response.Success(c, dto.ToCashRegisterResponse(cr))
}

// Deactivate deactivates a cash register
// @Summary Deactivate cash register
// @Description Deactivate a cash register
// @Tags CashRegister
// @Produce json
// @Param id path string true "Cash Register ID"
// @Success 200 {object} response.Response{data=dto.CashRegisterResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers/{id}/deactivate [post]
func (h *CashRegisterHandler) Deactivate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cash register ID format")
		return
	}

	// Get deactivatedBy from context (JWT middleware)
	deactivatedByStr := c.GetString("user_id")
	if deactivatedByStr == "" {
		response.BadRequest(c, "User not authenticated")
		return
	}

	deactivatedBy, err := uuidv7.Parse(deactivatedByStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	cr, err := h.u.DeactivateCashRegister(c.Request.Context(), id, deactivatedBy)
	if err != nil {
		switch {
		case errors.Is(err, cashregistererrors.ErrCashRegisterNotFound):
			response.NotFound(c, "Cash register not found")
		case errors.Is(err, cashregistererrors.ErrCashRegisterAlreadyInactive):
			response.BadRequest(c, "Cash register is already inactive")
		default:
			response.InternalError(c, "Failed to deactivate cash register")
		}
		return
	}

	response.Success(c, dto.ToCashRegisterResponse(cr))
}

// Delete deletes a cash register
// @Summary Delete cash register
// @Description Delete a cash register (soft delete)
// @Tags CashRegister
// @Produce json
// @Param id path string true "Cash Register ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers/{id} [delete]
func (h *CashRegisterHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cash register ID format")
		return
	}

	if err := h.u.DeleteCashRegister(c.Request.Context(), id); err != nil {
		if errors.Is(err, cashregistererrors.ErrCashRegisterNotFound) {
			response.NotFound(c, "Cash register not found")
			return
		}
		response.InternalError(c, "Failed to delete cash register")
		return
	}

	response.Success(c, gin.H{"message": "Cash register deleted successfully"})
}

// Sync synchronizes cash register with fiscal service
// @Summary Sync cash register
// @Description Sync cash register with fiscal service
// @Tags CashRegister
// @Produce json
// @Param id path string true "Cash Register ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/cash-registers/{id}/sync [post]
func (h *CashRegisterHandler) Sync(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cash register ID format")
		return
	}

	// Get operator ID from JWT claims (placeholder for now)
	operatorID := uuidv7.New()

	_, err = h.u.SyncCashRegister(c.Request.Context(), id, operatorID)
	if err != nil {
		if errors.Is(err, cashregistererrors.ErrCashRegisterNotFound) {
			response.NotFound(c, "Cash register not found")
			return
		}
		response.InternalError(c, "Failed to sync cash register")
		return
	}

	response.Success(c, gin.H{"message": "Cash register synced successfully"})
}
