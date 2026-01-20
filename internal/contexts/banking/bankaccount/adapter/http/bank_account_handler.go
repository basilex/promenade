package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	accounterrors "github.com/basilex/promenade/internal/contexts/banking/bankaccount"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/dto"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// System user ID for API operations (TODO: replace with actual authenticated user)
var systemUserID = uuidv7.MustParse("00000000-0000-0000-0000-000000000000")

// BankAccountHandler handles HTTP requests for bank account operations.
type BankAccountHandler struct {
	usecase usecase.IBankAccountUseCase
}

// NewBankAccountHandler creates a new BankAccountHandler.
func NewBankAccountHandler(uc usecase.IBankAccountUseCase) *BankAccountHandler {
	return &BankAccountHandler{
		usecase: uc,
	}
}

// CreateManual creates a new manual bank account.
// @Summary Create manual bank account
// @Description Create a manually entered bank account
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param request body dto.CreateManualAccountRequest true "Manual account creation request"
// @Success 201 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/manual [post]
func (h *BankAccountHandler) CreateManual(c *gin.Context) {
	var req dto.CreateManualAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, err := uuidv7.Parse(req.OrganizationID)
	if err != nil {
		response.BadRequest(c, "Invalid organization ID format")
		return
	}

	acc, err := h.usecase.CreateManualAccount(
		c.Request.Context(),
		organizationID,
		req.Name,
		req.BankName,
		req.CurrencyCode,
		systemUserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, accounterrors.ErrInvalidOrganizationID):
			response.BadRequest(c, "Invalid organization ID")
		case errors.Is(err, accounterrors.ErrInvalidAccountName):
			response.BadRequest(c, "Invalid account name")
		case errors.Is(err, accounterrors.ErrInvalidBankName):
			response.BadRequest(c, "Invalid bank name")
		default:
			response.InternalError(c, "Failed to create bank account")
		}
		return
	}

	response.Created(c, dto.ToBankAccountResponse(acc))
}

// ConnectProvider connects a provider-based bank account.
// @Summary Connect provider account
// @Description Connect a bank account from external provider (Monobank, Privat24, etc.)
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param request body dto.ConnectProviderAccountRequest true "Provider account connection request"
// @Success 201 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/connect [post]
func (h *BankAccountHandler) ConnectProvider(c *gin.Context) {
	var req dto.ConnectProviderAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, err := uuidv7.Parse(req.OrganizationID)
	if err != nil {
		response.BadRequest(c, "Invalid organization ID format")
		return
	}

	acc, err := h.usecase.ConnectProviderAccount(
		c.Request.Context(),
		organizationID,
		req.Name,
		req.BankName,
		aggregate.BankProvider(req.Provider),
		req.ProviderAccountID,
		systemUserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, accounterrors.ErrInvalidProvider):
			response.BadRequest(c, "Invalid provider")
		case errors.Is(err, accounterrors.ErrInvalidOrganizationID):
			response.BadRequest(c, "Invalid organization ID")
		case errors.Is(err, accounterrors.ErrInvalidAccountName):
			response.BadRequest(c, "Invalid account name")
		default:
			response.InternalError(c, "Failed to connect provider account")
		}
		return
	}

	response.Created(c, dto.ToBankAccountResponse(acc))
}

// GetByID retrieves a bank account by ID.
// @Summary Get bank account by ID
// @Description Get bank account details by ID
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id} [get]
func (h *BankAccountHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, accounterrors.ErrBankAccountNotFound) {
			response.NotFound(c, "Bank account not found")
			return
		}
		response.InternalError(c, "Failed to get bank account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// List retrieves bank accounts for an organization.
// @Summary List bank accounts
// @Description Get paginated list of bank accounts for organization
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param organization_id query string true "Organization ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.BankAccountListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts [get]
func (h *BankAccountHandler) List(c *gin.Context) {
	organizationID, err := uuidv7.Parse(c.Query("organization_id"))
	if err != nil {
		response.BadRequest(c, "Invalid organization ID format")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	accounts, err := h.usecase.ListAccounts(c.Request.Context(), organizationID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list accounts")
		return
	}

	total, err := h.usecase.CountAccounts(c.Request.Context(), organizationID)
	if err != nil {
		response.InternalError(c, "Failed to count accounts")
		return
	}

	response.Success(c, dto.ToBankAccountListResponse(accounts, total, limit, offset))
}

// UpdateDetails updates bank account details.
// @Summary Update account details
// @Description Update bank account name, bank name, IBAN, or account number
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Param request body dto.UpdateAccountDetailsRequest true "Account details update request"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id}/details [put]
func (h *BankAccountHandler) UpdateDetails(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	var req dto.UpdateAccountDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateAccountDetails(
		c.Request.Context(),
		id,
		req.Name,
		req.BankName,
		req.IBAN,
		req.AccountNumber,
		systemUserID,
	); err != nil {
		switch {
		case errors.Is(err, accounterrors.ErrBankAccountNotFound):
			response.NotFound(c, "Bank account not found")
		case errors.Is(err, accounterrors.ErrInvalidAccountName):
			response.BadRequest(c, "Invalid account name")
		case errors.Is(err, accounterrors.ErrInvalidBankName):
			response.BadRequest(c, "Invalid bank name")
		default:
			response.InternalError(c, "Failed to update account details")
		}
		return
	}

	// Reload account to get updated state
	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// UpdateBalance updates bank account balance.
// @Summary Update account balance
// @Description Update the current balance of a bank account
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Param request body dto.UpdateBalanceRequest true "Balance update request"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id}/balance [put]
func (h *BankAccountHandler) UpdateBalance(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	var req dto.UpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateBalance(c.Request.Context(), id, req.BalanceCents, systemUserID); err != nil {
		if errors.Is(err, accounterrors.ErrBankAccountNotFound) {
			response.NotFound(c, "Bank account not found")
			return
		}
		response.InternalError(c, "Failed to update balance")
		return
	}

	// Reload account to get updated state
	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// RecordSync records a successful sync operation.
// @Summary Record sync
// @Description Record a successful synchronization with external provider
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Param request body dto.UpdateBalanceRequest true "Balance from sync"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id}/sync [post]
func (h *BankAccountHandler) RecordSync(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	var req dto.UpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.RecordSync(c.Request.Context(), id, req.BalanceCents, systemUserID); err != nil {
		switch {
		case errors.Is(err, accounterrors.ErrBankAccountNotFound):
			response.NotFound(c, "Bank account not found")
		case errors.Is(err, accounterrors.ErrManualAccountCannotSync):
			response.BadRequest(c, "Manual accounts cannot be synced")
		default:
			response.InternalError(c, "Failed to record sync")
		}
		return
	}

	// Reload account to get updated state
	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// Activate activates a bank account.
// @Summary Activate account
// @Description Activate an inactive bank account
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id}/activate [post]
func (h *BankAccountHandler) Activate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	if err := h.usecase.ActivateAccount(c.Request.Context(), id, systemUserID); err != nil {
		if errors.Is(err, accounterrors.ErrBankAccountNotFound) {
			response.NotFound(c, "Bank account not found")
			return
		}
		response.InternalError(c, "Failed to activate account")
		return
	}

	// Reload account to get updated state
	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// Deactivate deactivates a bank account.
// @Summary Deactivate account
// @Description Deactivate an active bank account
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id}/deactivate [post]
func (h *BankAccountHandler) Deactivate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	if err := h.usecase.DeactivateAccount(c.Request.Context(), id, systemUserID); err != nil {
		if errors.Is(err, accounterrors.ErrBankAccountNotFound) {
			response.NotFound(c, "Bank account not found")
			return
		}
		response.InternalError(c, "Failed to deactivate account")
		return
	}

	// Reload account to get updated state
	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// Archive archives a bank account.
// @Summary Archive account
// @Description Archive a bank account (no further transactions)
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankAccountResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id}/archive [post]
func (h *BankAccountHandler) Archive(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	if err := h.usecase.ArchiveAccount(c.Request.Context(), id, systemUserID); err != nil {
		if errors.Is(err, accounterrors.ErrBankAccountNotFound) {
			response.NotFound(c, "Bank account not found")
			return
		}
		response.InternalError(c, "Failed to archive account")
		return
	}

	// Reload account to get updated state
	acc, err := h.usecase.GetAccount(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload account")
		return
	}

	response.Success(c, dto.ToBankAccountResponse(acc))
}

// Delete deletes a bank account.
// @Summary Delete account
// @Description Delete a bank account (soft delete)
// @Tags Bank Accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/accounts/{id} [delete]
func (h *BankAccountHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	if err := h.usecase.DeleteAccount(c.Request.Context(), id); err != nil {
		if errors.Is(err, accounterrors.ErrBankAccountNotFound) {
			response.NotFound(c, "Bank account not found")
			return
		}
		response.InternalError(c, "Failed to delete account")
		return
	}

	c.Status(http.StatusNoContent)
}
