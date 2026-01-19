package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	contracterrors "github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract/aggregate"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract/dto"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ContractHandler handles HTTP requests for contract operations
type ContractHandler struct {
	u usecase.IContractUseCase
}

// NewContractHandler creates a new contract handler
func NewContractHandler(u usecase.IContractUseCase) *ContractHandler {
	return &ContractHandler{
		u: u,
	}
}

// handleContractError discriminates domain errors and returns appropriate HTTP responses
func (h *ContractHandler) handleContractError(c *gin.Context, err error) {
	switch {
	// Repository errors → 404 Not Found
	case errors.Is(err, contracterrors.ErrContractNotFound):
		response.NotFound(c, "Contract not found")

	// State transition errors → 400 Bad Request with specific message
	case errors.Is(err, contracterrors.ErrCannotSubmitNonDraft):
		response.BadRequest(c, "Can only submit draft contracts for signature")

	case errors.Is(err, contracterrors.ErrCannotSignNonPending):
		response.BadRequest(c, "Can only sign contracts in pending signature status")

	case errors.Is(err, contracterrors.ErrCannotCompleteNonActive):
		response.BadRequest(c, "Can only complete active contracts")

	case errors.Is(err, contracterrors.ErrCannotTerminateNonActive):
		response.BadRequest(c, "Can only terminate active contracts")

	case errors.Is(err, contracterrors.ErrCannotRenewInactiveContract):
		response.BadRequest(c, "Can only renew active or completed contracts")

	case errors.Is(err, contracterrors.ErrCannotSetExpirationForActive):
		response.BadRequest(c, "Cannot set expiration date for signed contracts")

	// Validation errors → 400 Bad Request with specific message
	case errors.Is(err, contracterrors.ErrOrderIDRequired):
		response.BadRequest(c, "Order ID is required")

	case errors.Is(err, contracterrors.ErrCustomerIDRequired):
		response.BadRequest(c, "Customer ID is required")

	case errors.Is(err, contracterrors.ErrTermsRequired):
		response.BadRequest(c, "Contract terms are required")

	case errors.Is(err, contracterrors.ErrStatusRequired):
		response.BadRequest(c, "Contract status is required")

	case errors.Is(err, contracterrors.ErrSignerNameRequired):
		response.BadRequest(c, "Signer name is required")

	case errors.Is(err, contracterrors.ErrSignerEmailRequired):
		response.BadRequest(c, "Signer email is required")

	case errors.Is(err, contracterrors.ErrTerminationReasonRequired):
		response.BadRequest(c, "Termination reason is required")

	case errors.Is(err, contracterrors.ErrExpirationInPast):
		response.BadRequest(c, "Expiration date must be in the future")

	case errors.Is(err, contracterrors.ErrInvalidDaysValue):
		response.BadRequest(c, "Days value must be greater than 0")

	// Unknown/unexpected errors → 500 Internal Server Error
	default:
		response.InternalError(c, "An unexpected error occurred")
	}
}

// Create handles POST /contracts - Create contract
// @Summary Create contract
// @Description Create a new contract for an order
// @Tags contracts
// @Accept json
// @Produce json
// @Param contract body dto.CreateContractRequest true "Contract details"
// @Success 201 {object} response.Response{data=dto.ContractResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts [post]
func (h *ContractHandler) Create(c *gin.Context) {
	var req dto.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	created, err := h.u.CreateContract(c.Request.Context(), req.OrderID, req.CustomerID, req.Terms)
	if err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Created(c, dto.ToContractResponse(created))
}

// GetByID handles GET /contracts/:id - Get contract by ID
// @Summary Get contract by ID
// @Description Get contract details by ID
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Success 200 {object} response.Response{data=dto.ContractResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id} [get]
func (h *ContractHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	cntr, err := h.u.GetContract(c.Request.Context(), id)
	if err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, dto.ToContractResponse(cntr))
}

// Update handles PUT /contracts/:id - Update contract
// @Summary Update contract
// @Description Update contract terms (only for Draft status)
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Param contract body dto.UpdateContractRequest true "Updated contract details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id} [put]
func (h *ContractHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	var req dto.UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.u.UpdateContract(c.Request.Context(), id, req.Terms); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract updated successfully"})
}

// Delete handles DELETE /contracts/:id - Delete contract
// @Summary Delete contract
// @Description Delete a contract (soft delete)
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id} [delete]
func (h *ContractHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	if err := h.u.DeleteContract(c.Request.Context(), id); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract deleted successfully"})
}

// SubmitForSignature handles POST /contracts/:id/submit - Submit contract for signature
// @Summary Submit contract for signature
// @Description Submit a draft contract for signature
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id}/submit [post]
func (h *ContractHandler) SubmitForSignature(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	if err := h.u.SubmitForSignature(c.Request.Context(), id); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract submitted for signature"})
}

// Sign handles POST /contracts/:id/sign - Sign contract
// @Summary Sign contract
// @Description Sign a contract that is pending signature
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Param signature body dto.SignContractRequest true "Signature details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id}/sign [post]
func (h *ContractHandler) Sign(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	var req dto.SignContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.u.SignContract(c.Request.Context(), id, req.SignerName, req.SignerEmail, req.SignatureID); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract signed successfully"})
}

// Complete handles POST /contracts/:id/complete - Complete contract
// @Summary Complete contract
// @Description Mark an active contract as completed
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id}/complete [post]
func (h *ContractHandler) Complete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	if err := h.u.CompleteContract(c.Request.Context(), id); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract completed successfully"})
}

// Terminate handles POST /contracts/:id/terminate - Terminate contract
// @Summary Terminate contract
// @Description Terminate an active contract with a reason
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Param termination body dto.TerminateContractRequest true "Termination details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id}/terminate [post]
func (h *ContractHandler) Terminate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	var req dto.TerminateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.u.TerminateContract(c.Request.Context(), id, req.Reason); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract terminated successfully"})
}

// Renew handles POST /contracts/:id/renew - Renew contract
// @Summary Renew contract
// @Description Create a new version of the contract
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Success 201 {object} response.Response{data=dto.ContractResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id}/renew [post]
func (h *ContractHandler) Renew(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	if err := h.u.RenewContract(c.Request.Context(), id); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "contract renewed successfully"})
}

// SetExpirationDate handles PUT /contracts/:id/expiration - Set expiration date
// @Summary Set expiration date
// @Description Set expiration date for a draft contract
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Param expiration body dto.SetExpirationDateRequest true "Expiration date"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts/{id}/expiration [put]
func (h *ContractHandler) SetExpirationDate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid contract ID")
		return
	}

	var req dto.SetExpirationDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.u.SetExpirationDate(c.Request.Context(), id, req.ExpirationDate); err != nil {
		h.handleContractError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "expiration date set successfully"})
}

// ListByOrder handles GET /orders/:id/contracts - List contracts by order
// @Summary List contracts by order
// @Description Get all contracts for a specific order
// @Tags contracts
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=[]dto.ContractResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/contracts [get]
func (h *ContractHandler) ListByOrder(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	contracts, err := h.u.ListContractsByOrder(c.Request.Context(), orderID)
	if err != nil {
		h.handleContractError(c, err)
		return
	}

	// Convert to response DTOs
	respContracts := make([]dto.ContractResponse, 0, len(contracts))
	for _, cntr := range contracts {
		respContracts = append(respContracts, dto.ToContractResponse(cntr))
	}

	response.Success(c, respContracts)
}

// List handles GET /contracts - List contracts with filters
// @Summary List contracts
// @Description List contracts with optional filters (customer_id, status, expiring_days)
// @Tags contracts
// @Produce json
// @Param customer_id query string false "Customer ID (UUID)"
// @Param status query string false "Contract status (draft, pending_signature, active, completed, terminated)"
// @Param expiring_days query int false "Days until expiration (for filtering expiring contracts)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} response.Response{data=dto.ContractListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts [get]
func (h *ContractHandler) List(c *gin.Context) {
	// Parse query parameters
	customerIDStr := c.Query("customer_id")
	statusStr := c.Query("status")
	expiringDaysStr := c.Query("expiring_days")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Filter by expiring contracts
	if expiringDaysStr != "" {
		days, err := strconv.Atoi(expiringDaysStr)
		if err != nil || days < 1 {
			response.BadRequest(c, "invalid expiring_days parameter")
			return
		}

		contracts, err := h.u.ListExpiringSoon(c.Request.Context(), days)
		if err != nil {
			h.handleContractError(c, err)
			return
		}

		resp := dto.ToContractListResponse(contracts, len(contracts))
		response.Success(c, resp)
		return
	}

	// Filter by customer
	if customerIDStr != "" {
		customerID, err := uuidv7.Parse(customerIDStr)
		if err != nil {
			response.BadRequest(c, "invalid customer_id")
			return
		}

		contracts, total, err := h.u.ListContractsByCustomer(c.Request.Context(), customerID, page, pageSize)
		if err != nil {
			h.handleContractError(c, err)
			return
		}

		resp := dto.ToContractListResponse(contracts, total)
		response.Success(c, resp)
		return
	}

	// Filter by status
	if statusStr != "" {
		status := aggregate.ContractStatus(statusStr)
		contracts, total, err := h.u.ListContractsByStatus(c.Request.Context(), status, page, pageSize)
		if err != nil {
			h.handleContractError(c, err)
			return
		}

		resp := dto.ToContractListResponse(contracts, total)
		response.Success(c, resp)
		return
	}

	// Get all active contracts (default)
	contracts, err := h.u.GetActiveContracts(c.Request.Context())
	if err != nil {
		h.handleContractError(c, err)
		return
	}

	resp := dto.ToContractListResponse(contracts, len(contracts))
	response.Success(c, resp)
}
