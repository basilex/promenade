package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ContractHandler handles HTTP requests for contract operations
type ContractHandler struct {
	usecase contract.IUseCase
}

// NewContractHandler creates a new contract handler
func NewContractHandler(usecase contract.IUseCase) *ContractHandler {
	return &ContractHandler{
		usecase: usecase,
	}
}

// Create handles POST /contracts - Create contract
// @Summary Create contract
// @Description Create a new contract for an order
// @Tags contracts
// @Accept json
// @Produce json
// @Param contract body CreateContractRequest true "Contract details"
// @Success 201 {object} response.Response{data=ContractResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /contracts [post]
func (h *ContractHandler) Create(c *gin.Context) {
	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	created, err := h.usecase.CreateContract(c.Request.Context(), req.OrderID, req.CustomerID, req.Terms)
	if err != nil {
		response.InternalError(c, "Failed to create contract")
		return
	}

	response.Created(c, ToContractResponse(created))
}

// GetByID handles GET /contracts/:id - Get contract by ID
// @Summary Get contract by ID
// @Description Get contract details by ID
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Success 200 {object} response.Response{data=ContractResponse}
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

	cntr, err := h.usecase.GetContract(c.Request.Context(), id)
	if errors.Is(err, contract.ErrContractNotFound) {
		response.NotFound(c, "Contract not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to retrieve contract")
		return
	}

	response.Success(c, ToContractResponse(cntr))
}

// Update handles PUT /contracts/:id - Update contract
// @Summary Update contract
// @Description Update contract terms (only for Draft status)
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID (UUID)"
// @Param contract body UpdateContractRequest true "Updated contract details"
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

	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateContract(c.Request.Context(), id, req.Terms); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		response.InternalError(c, "Failed to update contract")
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

	if err := h.usecase.DeleteContract(c.Request.Context(), id); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		response.InternalError(c, "Failed to delete contract")
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

	if err := h.usecase.SubmitForSignature(c.Request.Context(), id); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		if errors.Is(err, contract.ErrInvalidContractTransition) {
			response.BadRequest(c, "Invalid contract status transition")
			return
		}
		response.InternalError(c, "Failed to submit contract for signature")
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
// @Param signature body SignContractRequest true "Signature details"
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

	var req SignContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.SignContract(c.Request.Context(), id, req.SignerName, req.SignerEmail, req.SignatureID); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		if errors.Is(err, contract.ErrInvalidContractTransition) {
			response.BadRequest(c, "Invalid contract status transition")
			return
		}
		response.InternalError(c, "Failed to sign contract")
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

	if err := h.usecase.CompleteContract(c.Request.Context(), id); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		response.InternalError(c, "Failed to complete contract")
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
// @Param termination body TerminateContractRequest true "Termination details"
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

	var req TerminateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.TerminateContract(c.Request.Context(), id, req.Reason); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		if errors.Is(err, contract.ErrInvalidContractTransition) {
			response.BadRequest(c, "Invalid contract status transition")
			return
		}
		response.InternalError(c, "Failed to terminate contract")
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
// @Success 201 {object} response.Response{data=ContractResponse}
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

	if err := h.usecase.RenewContract(c.Request.Context(), id); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		response.InternalError(c, "Failed to renew contract")
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
// @Param expiration body SetExpirationDateRequest true "Expiration date"
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

	var req SetExpirationDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.SetExpirationDate(c.Request.Context(), id, req.ExpirationDate); err != nil {
		if errors.Is(err, contract.ErrContractNotFound) {
			response.NotFound(c, "Contract not found")
			return
		}
		response.InternalError(c, "Failed to set expiration date")
		return
	}

	response.Success(c, gin.H{"message": "expiration date set successfully"})
}

// ListByOrder handles GET /orders/:order_id/contracts - List contracts by order
// @Summary List contracts by order
// @Description Get all contracts for a specific order
// @Tags contracts
// @Produce json
// @Param order_id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=[]ContractResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{order_id}/contracts [get]
func (h *ContractHandler) ListByOrder(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("order_id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	contracts, err := h.usecase.ListContractsByOrder(c.Request.Context(), orderID)
	if err != nil {
		response.InternalError(c, "Failed to list contracts by order")
		return
	}

	// Convert to response DTOs
	respContracts := make([]ContractResponse, 0, len(contracts))
	for _, cntr := range contracts {
		respContracts = append(respContracts, ToContractResponse(cntr))
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
// @Success 200 {object} response.Response{data=ContractListResponse}
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

		contracts, err := h.usecase.ListExpiringSoon(c.Request.Context(), days)
		if err != nil {
			response.InternalError(c, "Failed to list expiring contracts")
			return
		}

		resp := ToContractListResponse(contracts, len(contracts))
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

		contracts, total, err := h.usecase.ListContractsByCustomer(c.Request.Context(), customerID, page, pageSize)
		if err != nil {
			response.InternalError(c, "Failed to list contracts by customer")
			return
		}

		resp := ToContractListResponse(contracts, total)
		response.Success(c, resp)
		return
	}

	// Filter by status
	if statusStr != "" {
		status := contract.ContractStatus(statusStr)
		contracts, total, err := h.usecase.ListContractsByStatus(c.Request.Context(), status, page, pageSize)
		if err != nil {
			response.InternalError(c, "Failed to list contracts by status")
			return
		}

		resp := ToContractListResponse(contracts, total)
		response.Success(c, resp)
		return
	}

	// Get all active contracts (default)
	contracts, err := h.usecase.GetActiveContracts(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list active contracts")
		return
	}

	resp := ToContractListResponse(contracts, len(contracts))
	response.Success(c, resp)
}
