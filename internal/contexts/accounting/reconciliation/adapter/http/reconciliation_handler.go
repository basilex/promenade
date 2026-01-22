package http

import (
	"errors"
	"strconv"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/dto"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

type ReconciliationHandler struct {
	uc usecase.IReconciliationUseCase
}

func NewReconciliationHandler(uc usecase.IReconciliationUseCase) *ReconciliationHandler {
	return &ReconciliationHandler{uc: uc}
}

func (h *ReconciliationHandler) RegisterRoutes(router *gin.RouterGroup) {
	reconciliations := router.Group("/bank-reconciliations")
	{
		reconciliations.POST("", h.CreateReconciliation)
		reconciliations.GET("/:id", h.GetReconciliation)
		reconciliations.POST("/:id/items", h.AddReconciliationItem)
		reconciliations.DELETE("/:id/items/:itemId", h.RemoveReconciliationItem)
		reconciliations.POST("/:id/items/:itemId/match", h.MarkItemMatched)
		reconciliations.POST("/:id/complete", h.CompleteReconciliation)
		reconciliations.POST("/:id/reopen", h.ReopenReconciliation)
		reconciliations.DELETE("/:id", h.DeleteReconciliation)
		reconciliations.GET("/bank-account/:bankAccountId", h.ListReconciliationsByBankAccount)
		reconciliations.GET("", h.ListReconciliations)
		reconciliations.GET("/status/:status", h.ListReconciliationsByStatus)
	}
}

// CreateReconciliation creates a new bank reconciliation
// @Summary Create a new bank reconciliation
// @Tags bank-reconciliations
// @Accept json
// @Produce json
// @Param request body dto.CreateReconciliationRequest true "Create bank reconciliation request"
// @Success 201 {object} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations [post]
func (h *ReconciliationHandler) CreateReconciliation(c *gin.Context) {
	var req dto.CreateReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	bankAccountID, err := uuidv7.Parse(req.BankAccountID)
	if err != nil {
		response.BadRequest(c, "invalid bank_account_id")
		return
	}

	accountID, err := uuidv7.Parse(req.AccountID)
	if err != nil {
		response.BadRequest(c, "invalid account_id")
		return
	}

	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))
	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	br, err := h.uc.CreateReconciliation(
		c.Request.Context(),
		organizationID,
		bankAccountID,
		accountID,
		req.ReconciliationDate,
		req.StatementDate,
		req.BankStatementBalanceCents,
		req.BookBalanceCents,
		req.CurrencyCode,
		userUUID,
	)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Created(c, dto.ToReconciliationResponse(br))
}

// GetReconciliation gets a bank reconciliation by ID
// @Summary Get a bank reconciliation
// @Tags bank-reconciliations
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Success 200 {object} dto.ReconciliationResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id} [get]
func (h *ReconciliationHandler) GetReconciliation(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	br, err := h.uc.GetReconciliationByID(c.Request.Context(), id)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Success(c, dto.ToReconciliationResponse(br))
}

// AddReconciliationItem adds an item to the bank reconciliation
// @Summary Add item to bank reconciliation
// @Tags bank-reconciliations
// @Accept json
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Param request body dto.AddReconciliationItemRequest true "Add reconciliation item request"
// @Success 200 {object} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id}/items [post]
func (h *ReconciliationHandler) AddReconciliationItem(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	var req dto.AddReconciliationItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	transactionType := reconciliation.TransactionType(req.TransactionType)
	if !isValidTransactionType(transactionType) {
		response.BadRequest(c, "invalid transaction type")
		return
	}

	var transactionID *uuidv7.UUID
	if req.TransactionID != nil {
		parsed, err := uuidv7.Parse(*req.TransactionID)
		if err != nil {
			response.BadRequest(c, "invalid transaction_id")
			return
		}
		transactionID = &parsed
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	br, err := h.uc.AddReconciliationItem(
		c.Request.Context(),
		id,
		transactionType,
		transactionID,
		req.TransactionDate,
		req.Description,
		req.AmountCents,
		req.Notes,
		userUUID,
	)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Success(c, dto.ToReconciliationResponse(br))
}

// RemoveReconciliationItem removes an item from the bank reconciliation
// @Summary Remove item from bank reconciliation
// @Tags bank-reconciliations
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Param itemId path string true "Item ID"
// @Success 200 {object} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id}/items/{itemId} [delete]
func (h *ReconciliationHandler) RemoveReconciliationItem(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	itemID, err := uuidv7.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	br, err := h.uc.RemoveReconciliationItem(c.Request.Context(), id, itemID, userUUID)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Success(c, dto.ToReconciliationResponse(br))
}

// MarkItemMatched marks a reconciliation item as matched
// @Summary Mark reconciliation item as matched
// @Tags bank-reconciliations
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Param itemId path string true "Item ID"
// @Success 200 {object} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id}/items/{itemId}/match [post]
func (h *ReconciliationHandler) MarkItemMatched(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	itemID, err := uuidv7.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "invalid item ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	br, err := h.uc.MarkItemMatched(c.Request.Context(), id, itemID, userUUID)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Success(c, dto.ToReconciliationResponse(br))
}

// CompleteReconciliation completes a bank reconciliation
// @Summary Complete bank reconciliation
// @Tags bank-reconciliations
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Success 200 {object} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id}/complete [post]
func (h *ReconciliationHandler) CompleteReconciliation(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	br, err := h.uc.CompleteReconciliation(c.Request.Context(), id, userUUID)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Success(c, dto.ToReconciliationResponse(br))
}

// ReopenReconciliation reopens a completed bank reconciliation
// @Summary Reopen bank reconciliation
// @Tags bank-reconciliations
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Success 200 {object} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id}/reopen [post]
func (h *ReconciliationHandler) ReopenReconciliation(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	br, err := h.uc.ReopenReconciliation(c.Request.Context(), id, userUUID)
	if err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.Success(c, dto.ToReconciliationResponse(br))
}

// DeleteReconciliation deletes a bank reconciliation
// @Summary Delete a bank reconciliation
// @Tags bank-reconciliations
// @Produce json
// @Param id path string true "Bank Reconciliation ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/{id} [delete]
func (h *ReconciliationHandler) DeleteReconciliation(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid bank reconciliation ID")
		return
	}

	if err := h.uc.DeleteReconciliation(c.Request.Context(), id); err != nil {
		handleReconciliationError(c, err)
		return
	}

	response.SuccessWithMessage(c, "bank reconciliation deleted successfully")
}

// ListReconciliationsByBankAccount lists all bank reconciliations for a bank account
// @Summary List bank reconciliations by bank account
// @Tags bank-reconciliations
// @Produce json
// @Param bankAccountId path string true "Bank Account ID"
// @Success 200 {array} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/bank-account/{bankAccountId} [get]
func (h *ReconciliationHandler) ListReconciliationsByBankAccount(c *gin.Context) {
	bankAccountID, err := uuidv7.Parse(c.Param("bankAccountId"))
	if err != nil {
		response.BadRequest(c, "invalid bank account ID")
		return
	}

	reconciliations, err := h.uc.ListReconciliationsByBankAccount(c.Request.Context(), bankAccountID)
	if err != nil {
		response.InternalError(c, "failed to list bank reconciliations")
		return
	}

	response.Success(c, dto.ToReconciliationResponseList(reconciliations))
}

// ListReconciliations lists all bank reconciliations for an organization
// @Summary List bank reconciliations
// @Tags bank-reconciliations
// @Produce json
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} dto.ReconciliationResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations [get]
func (h *ReconciliationHandler) ListReconciliations(c *gin.Context) {
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

	reconciliations, err := h.uc.ListReconciliationsByOrganization(c.Request.Context(), organizationID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list bank reconciliations")
		return
	}

	response.Success(c, dto.ToReconciliationResponseList(reconciliations))
}

// ListReconciliationsByStatus lists bank reconciliations by status
// @Summary List bank reconciliations by status
// @Tags bank-reconciliations
// @Produce json
// @Param status path string true "Reconciliation status"
// @Success 200 {array} dto.ReconciliationResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/bank-reconciliations/status/{status} [get]
func (h *ReconciliationHandler) ListReconciliationsByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	status := reconciliation.Status(statusStr)
	if !isValidReconciliationStatus(status) {
		response.BadRequest(c, "invalid reconciliation status")
		return
	}

	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	reconciliations, err := h.uc.ListReconciliationsByStatus(c.Request.Context(), organizationID, status)
	if err != nil {
		response.InternalError(c, "failed to list bank reconciliations by status")
		return
	}

	response.Success(c, dto.ToReconciliationResponseList(reconciliations))
}

func handleReconciliationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, reconciliation.ErrReconciliationNotFound):
		response.NotFound(c, "bank reconciliation not found")
	case errors.Is(err, reconciliation.ErrDescriptionRequired),
		errors.Is(err, reconciliation.ErrCannotModifyCompleted),
		errors.Is(err, reconciliation.ErrAlreadyCompleted),
		errors.Is(err, reconciliation.ErrHasUnmatchedItems),
		errors.Is(err, reconciliation.ErrNotCompleted),
		errors.Is(err, reconciliation.ErrCannotReopenApproved),
		errors.Is(err, reconciliation.ErrAlreadyInProgress),
		errors.Is(err, reconciliation.ErrItemNotFound):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "operation failed")
	}
}

func isValidTransactionType(t reconciliation.TransactionType) bool {
	switch t {
	case reconciliation.TransactionTypeBankTransaction,
		reconciliation.TransactionTypeJournalEntry,
		reconciliation.TransactionTypeOutstanding:
		return true
	default:
		return false
	}
}

func isValidReconciliationStatus(s reconciliation.Status) bool {
	switch s {
	case reconciliation.StatusInProgress,
		reconciliation.StatusCompleted,
		reconciliation.StatusApproved:
		return true
	default:
		return false
	}
}
