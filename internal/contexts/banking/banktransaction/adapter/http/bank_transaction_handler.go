package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	txerrors "github.com/basilex/promenade/internal/contexts/banking/banktransaction"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/dto"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// System user ID for API operations (TODO: replace with actual authenticated user)
var systemUserID = uuidv7.MustParse("00000000-0000-0000-0000-000000000000")

// BankTransactionHandler handles HTTP requests for bank transaction operations.
type BankTransactionHandler struct {
	usecase usecase.IBankTransactionUseCase
}

// NewBankTransactionHandler creates a new BankTransactionHandler.
func NewBankTransactionHandler(uc usecase.IBankTransactionUseCase) *BankTransactionHandler {
	return &BankTransactionHandler{
		usecase: uc,
	}
}

// Record records a new bank transaction.
// @Summary Record bank transaction
// @Description Record a new bank transaction (manual or from sync)
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param request body dto.RecordTransactionRequest true "Transaction record request"
// @Success 201 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions [post]
func (h *BankTransactionHandler) Record(c *gin.Context) {
	var req dto.RecordTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accountID, err := uuidv7.Parse(req.AccountID)
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
		return
	}

	var tx *aggregate.BankTransaction
	if req.ExternalID != "" {
		tx, err = h.usecase.RecordTransactionWithExternalID(
			c.Request.Context(),
			accountID,
			req.ExternalID,
			aggregate.TransactionDirection(req.Direction),
			req.AmountCents,
			req.CurrencyCode,
			req.TransactionAt,
			req.Description,
			systemUserID,
		)
		if err != nil {
			if errors.Is(err, txerrors.ErrExternalIDExists) {
				response.Conflict(c, "Transaction with this external ID already exists")
				return
			}
		}
	} else {
		tx, err = h.usecase.RecordTransaction(
			c.Request.Context(),
			accountID,
			aggregate.TransactionDirection(req.Direction),
			req.AmountCents,
			req.CurrencyCode,
			req.TransactionAt,
			req.Description,
			systemUserID,
		)
	}

	if err != nil {
		switch {
		case errors.Is(err, txerrors.ErrInvalidBankAccountID):
			response.BadRequest(c, "Invalid bank account ID")
		case errors.Is(err, txerrors.ErrInvalidDirection):
			response.BadRequest(c, "Invalid transaction direction")
		case errors.Is(err, txerrors.ErrInvalidAmount):
			response.BadRequest(c, "Invalid transaction amount")
		default:
			response.InternalError(c, "Failed to record transaction")
		}
		return
	}

	// Set counterparty if provided
	if req.CounterpartyName != "" {
		if err := h.usecase.SetCounterparty(c.Request.Context(), tx.GetID(), req.CounterpartyName, req.CounterpartyIBAN, systemUserID); err != nil {
			// Non-fatal, just log
			_ = err // Error intentionally ignored - counterparty update is non-critical
		}
	}

	// Reload transaction to get updated state
	tx, err = h.usecase.GetTransaction(c.Request.Context(), tx.GetID())
	if err != nil {
		response.InternalError(c, "Failed to reload transaction")
		return
	}

	response.Created(c, dto.ToBankTransactionResponse(tx))
}

// GetByID retrieves a bank transaction by ID.
// @Summary Get transaction by ID
// @Description Get bank transaction details by ID
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id} [get]
func (h *BankTransactionHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	tx, err := h.usecase.GetTransaction(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, txerrors.ErrBankTransactionNotFound) {
			response.NotFound(c, "Transaction not found")
			return
		}
		response.InternalError(c, "Failed to get transaction")
		return
	}

	response.Success(c, dto.ToBankTransactionResponse(tx))
}

// ListByAccount retrieves transactions for a bank account.
// @Summary List transactions by account
// @Description Get paginated list of transactions for a bank account
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param account_id query string true "Bank Account ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.BankTransactionListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions [get]
func (h *BankTransactionHandler) ListByAccount(c *gin.Context) {
	accountID, err := uuidv7.Parse(c.Query("account_id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
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

	transactions, err := h.usecase.ListTransactionsByAccount(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list transactions")
		return
	}

	total, err := h.usecase.CountTransactionsByAccount(c.Request.Context(), accountID)
	if err != nil {
		response.InternalError(c, "Failed to count transactions")
		return
	}

	response.Success(c, dto.ToBankTransactionListResponse(transactions, total, limit, offset))
}

// ListUnmatched retrieves unmatched transactions for reconciliation.
// @Summary List unmatched transactions
// @Description Get paginated list of unmatched transactions for a bank account
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param account_id query string true "Bank Account ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=dto.BankTransactionListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/unmatched [get]
func (h *BankTransactionHandler) ListUnmatched(c *gin.Context) {
	accountID, err := uuidv7.Parse(c.Query("account_id"))
	if err != nil {
		response.BadRequest(c, "Invalid account ID format")
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

	transactions, err := h.usecase.ListUnmatchedTransactions(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list unmatched transactions")
		return
	}

	// For unmatched list, we don't have accurate total count
	total := len(transactions)

	response.Success(c, dto.ToBankTransactionListResponse(transactions, total, limit, offset))
}

// Book marks a transaction as booked.
// @Summary Book transaction
// @Description Mark a pending transaction as booked
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id}/book [post]
func (h *BankTransactionHandler) Book(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	if err := h.usecase.BookTransaction(c.Request.Context(), id, systemUserID); err != nil {
		if errors.Is(err, txerrors.ErrBankTransactionNotFound) {
			response.NotFound(c, "Transaction not found")
			return
		}
		response.InternalError(c, "Failed to book transaction")
		return
	}

	// Reload transaction to get updated state
	tx, err := h.usecase.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload transaction")
		return
	}

	response.Success(c, dto.ToBankTransactionResponse(tx))
}

// Cancel cancels a transaction.
// @Summary Cancel transaction
// @Description Cancel a pending transaction
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id}/cancel [post]
func (h *BankTransactionHandler) Cancel(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	if err := h.usecase.CancelTransaction(c.Request.Context(), id, systemUserID); err != nil {
		if errors.Is(err, txerrors.ErrBankTransactionNotFound) {
			response.NotFound(c, "Transaction not found")
			return
		}
		response.InternalError(c, "Failed to cancel transaction")
		return
	}

	// Reload transaction to get updated state
	tx, err := h.usecase.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload transaction")
		return
	}

	response.Success(c, dto.ToBankTransactionResponse(tx))
}

// Match matches a transaction to an entity.
// @Summary Match transaction
// @Description Match a bank transaction to an invoice, order, or payment
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Param request body dto.MatchTransactionRequest true "Match request"
// @Success 200 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id}/match [post]
func (h *BankTransactionHandler) Match(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	var req dto.MatchTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	entityID, err := uuidv7.Parse(req.EntityID)
	if err != nil {
		response.BadRequest(c, "Invalid entity ID format")
		return
	}

	switch req.EntityType {
	case "invoice":
		err = h.usecase.MatchToInvoice(c.Request.Context(), id, entityID, systemUserID)
	case "order":
		err = h.usecase.MatchToOrder(c.Request.Context(), id, entityID, systemUserID)
	case "payment":
		err = h.usecase.MatchToPayment(c.Request.Context(), id, entityID, systemUserID)
	default:
		response.BadRequest(c, "Invalid entity type (must be invoice, order, or payment)")
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, txerrors.ErrBankTransactionNotFound):
			response.NotFound(c, "Transaction not found")
		case errors.Is(err, txerrors.ErrAlreadyMatched):
			response.BadRequest(c, "Transaction is already matched")
		case errors.Is(err, txerrors.ErrCannotMatchPendingTx):
			response.BadRequest(c, "Cannot match pending transaction")
		case errors.Is(err, txerrors.ErrInvalidMatchedEntity):
			response.BadRequest(c, "Invalid entity type")
		default:
			response.InternalError(c, "Failed to match transaction")
		}
		return
	}

	// Reload transaction to get updated state
	tx, err := h.usecase.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload transaction")
		return
	}

	response.Success(c, dto.ToBankTransactionResponse(tx))
}

// Unmatch unmatches a transaction.
// @Summary Unmatch transaction
// @Description Remove matching from a bank transaction
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Success 200 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id}/unmatch [post]
func (h *BankTransactionHandler) Unmatch(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	if err := h.usecase.UnmatchTransaction(c.Request.Context(), id, systemUserID); err != nil {
		if errors.Is(err, txerrors.ErrBankTransactionNotFound) {
			response.NotFound(c, "Transaction not found")
			return
		}
		response.InternalError(c, "Failed to unmatch transaction")
		return
	}

	// Reload transaction to get updated state
	tx, err := h.usecase.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload transaction")
		return
	}

	response.Success(c, dto.ToBankTransactionResponse(tx))
}

// SetCounterparty sets or updates counterparty information.
// @Summary Set counterparty
// @Description Set or update counterparty name and IBAN
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Param request body dto.SetCounterpartyRequest true "Counterparty request"
// @Success 200 {object} response.Response{data=dto.BankTransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id}/counterparty [put]
func (h *BankTransactionHandler) SetCounterparty(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	var req dto.SetCounterpartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.SetCounterparty(c.Request.Context(), id, req.Name, req.IBAN, systemUserID); err != nil {
		switch {
		case errors.Is(err, txerrors.ErrBankTransactionNotFound):
			response.NotFound(c, "Transaction not found")
		case errors.Is(err, txerrors.ErrInvalidCounterparty):
			response.BadRequest(c, "Invalid counterparty information")
		default:
			response.InternalError(c, "Failed to set counterparty")
		}
		return
	}

	// Reload transaction to get updated state
	tx, err := h.usecase.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to reload transaction")
		return
	}

	response.Success(c, dto.ToBankTransactionResponse(tx))
}

// Delete deletes a bank transaction.
// @Summary Delete transaction
// @Description Delete a bank transaction (soft delete)
// @Tags Bank Transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /banking/transactions/{id} [delete]
func (h *BankTransactionHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID format")
		return
	}

	if err := h.usecase.DeleteTransaction(c.Request.Context(), id); err != nil {
		if errors.Is(err, txerrors.ErrBankTransactionNotFound) {
			response.NotFound(c, "Transaction not found")
			return
		}
		response.InternalError(c, "Failed to delete transaction")
		return
	}

	c.Status(http.StatusNoContent)
}
