package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/pkg/response"
)

// SubscriptionHandler handles HTTP requests
type SubscriptionHandler struct {
	usecase subscription.IUseCase
}

// NewSubscriptionHandler creates handler
func NewSubscriptionHandler(usecase subscription.IUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{usecase: usecase}
}

// handleError discriminates domain errors and returns appropriate HTTP response
func (h *SubscriptionHandler) handleError(c *gin.Context, err error, defaultCode string, defaultMessage string) {
	// 404 Not Found
	if errors.Is(err, subscription.ErrSubscriptionNotFound) {
		response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
		return
	}

	// 400 Bad Request - Validation Errors
	if errors.Is(err, subscription.ErrCustomerIDRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "CUSTOMER_ID_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrPlanIDRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "PLAN_ID_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrCurrencyRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "CURRENCY_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrAmountMustBePositive) {
		response.ErrorResponse(c, http.StatusBadRequest, "AMOUNT_MUST_BE_POSITIVE", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrInvalidMoney) {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_MONEY", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrStartDateRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "START_DATE_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrRenewalDateInvalid) {
		response.ErrorResponse(c, http.StatusBadRequest, "RENEWAL_DATE_INVALID", err.Error())
		return
	}

	// 409 Conflict - State Machine Errors
	if errors.Is(err, subscription.ErrCannotActivate) {
		response.ErrorResponse(c, http.StatusConflict, "CANNOT_ACTIVATE", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrCanOnlyPauseActive) {
		response.ErrorResponse(c, http.StatusConflict, "CAN_ONLY_PAUSE_ACTIVE", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrCanOnlyResumePaused) {
		response.ErrorResponse(c, http.StatusConflict, "CAN_ONLY_RESUME_PAUSED", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrAlreadyInTerminalStatus) {
		response.ErrorResponse(c, http.StatusConflict, "ALREADY_IN_TERMINAL_STATUS", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrCanOnlyRenewActive) {
		response.ErrorResponse(c, http.StatusConflict, "CAN_ONLY_RENEW_ACTIVE", err.Error())
		return
	}
	if errors.Is(err, subscription.ErrAlreadyExpired) {
		response.ErrorResponse(c, http.StatusConflict, "ALREADY_EXPIRED", err.Error())
		return
	}

	// 500 Internal Server Error - Technical Errors (hide details)
	if errors.Is(err, subscription.ErrQueryFailed) ||
		errors.Is(err, subscription.ErrCreateFailed) ||
		errors.Is(err, subscription.ErrUpdateFailed) ||
		errors.Is(err, subscription.ErrDeleteFailed) {
		response.ErrorResponse(c, http.StatusInternalServerError, defaultCode, defaultMessage)
		return
	}

	// Default: 500 Internal Server Error
	response.ErrorResponse(c, http.StatusInternalServerError, defaultCode, defaultMessage)
}

// Create creates subscription
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	customerID, err := ParseCustomerID(req.CustomerID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	billingPeriod := ParseBillingPeriod(req.BillingPeriod)

	sub, err := h.usecase.CreateSubscription(c.Request.Context(), customerID, req.PlanID, billingPeriod, req.Currency, req.Amount, req.TrialDays)
	if err != nil {
		h.handleError(c, err, "CREATE_FAILED", "Failed to create subscription")
		return
	}

	response.Created(c, ToSubscriptionResponse(sub))
}

// GetByID retrieves subscription
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id, err := ParseCustomerID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "GET_FAILED", "Failed to get subscription")
		return
	}

	response.Success(c, ToSubscriptionResponse(sub))
}

// Update updates subscription
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, err := ParseCustomerID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	planID := ""
	if req.PlanID != nil {
		planID = *req.PlanID
	}

	amount := int64(0)
	if req.Amount != nil {
		amount = *req.Amount
	}

	if err := h.usecase.UpdateSubscription(c.Request.Context(), id, planID, amount); err != nil {
		h.handleError(c, err, "UPDATE_FAILED", "Failed to update subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, ToSubscriptionResponse(sub))
}

// Delete soft-deletes subscription
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, err := ParseCustomerID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	if err := h.usecase.DeleteSubscription(c.Request.Context(), id); err != nil {
		h.handleError(c, err, "DELETE_FAILED", "Failed to delete subscription")
		return
	}

	response.Success(c, gin.H{"message": "Subscription deleted successfully"})
}

// List retrieves subscriptions
func (h *SubscriptionHandler) List(c *gin.Context) {
	var req ListSubscriptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	var subscriptions []*subscription.Subscription
	var err error

	if req.CustomerID != "" {
		customerID, parseErr := ParseCustomerID(req.CustomerID)
		if parseErr != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", parseErr.Error())
			return
		}
		subscriptions, err = h.usecase.ListByCustomer(c.Request.Context(), customerID)
	} else if req.Status != "" {
		status := subscription.SubscriptionStatus(req.Status)
		subscriptions, err = h.usecase.ListByStatus(c.Request.Context(), status)
	} else {
		subscriptions, err = h.usecase.ListSubscriptions(c.Request.Context(), req.Page, req.PageSize)
	}

	if err != nil {
		h.handleError(c, err, "LIST_FAILED", "Failed to list subscriptions")
		return
	}

	result := make([]SubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		result[i] = ToSubscriptionResponse(sub)
	}

	response.Success(c, result)
}

// Activate activates subscription
func (h *SubscriptionHandler) Activate(c *gin.Context) {
	id, err := ParseCustomerID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	if err := h.usecase.ActivateSubscription(c.Request.Context(), id); err != nil {
		h.handleError(c, err, "ACTIVATE_FAILED", "Failed to activate subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, ToSubscriptionResponse(sub))
}

// Pause pauses subscription
func (h *SubscriptionHandler) Pause(c *gin.Context) {
	id, err := ParseCustomerID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	if err := h.usecase.PauseSubscription(c.Request.Context(), id); err != nil {
		h.handleError(c, err, "PAUSE_FAILED", "Failed to pause subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, ToSubscriptionResponse(sub))
}

// Cancel cancels subscription
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	id, err := ParseCustomerID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	reason := "User requested cancellation"
	effectiveDate := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

	if err := h.usecase.CancelSubscription(c.Request.Context(), id, reason, effectiveDate); err != nil {
		h.handleError(c, err, "CANCEL_FAILED", "Failed to cancel subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, ToSubscriptionResponse(sub))
}
