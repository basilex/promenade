package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	subscriptionerrors "github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/aggregate"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/dto"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// SubscriptionHandler handles HTTP requests
type SubscriptionHandler struct {
	u usecase.ISubscriptionUseCase
}

// NewSubscriptionHandler creates handler
func NewSubscriptionHandler(u usecase.ISubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{u: u}
}

// handleError discriminates domain errors and returns appropriate HTTP response
func (h *SubscriptionHandler) handleError(c *gin.Context, err error, defaultCode string, defaultMessage string) {
	// 404 Not Found
	if errors.Is(err, subscriptionerrors.ErrSubscriptionNotFound) {
		response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
		return
	}

	// 400 Bad Request - Validation Errors
	if errors.Is(err, subscriptionerrors.ErrCustomerIDRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "CUSTOMER_ID_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrPlanIDRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "PLAN_ID_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrCurrencyRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "CURRENCY_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrAmountMustBePositive) {
		response.ErrorResponse(c, http.StatusBadRequest, "AMOUNT_MUST_BE_POSITIVE", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrInvalidMoney) {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_MONEY", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrStartDateRequired) {
		response.ErrorResponse(c, http.StatusBadRequest, "START_DATE_REQUIRED", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrRenewalDateInvalid) {
		response.ErrorResponse(c, http.StatusBadRequest, "RENEWAL_DATE_INVALID", err.Error())
		return
	}

	// 409 Conflict - State Machine Errors
	if errors.Is(err, subscriptionerrors.ErrCannotActivate) {
		response.ErrorResponse(c, http.StatusConflict, "CANNOT_ACTIVATE", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrCanOnlyPauseActive) {
		response.ErrorResponse(c, http.StatusConflict, "CAN_ONLY_PAUSE_ACTIVE", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrCanOnlyResumePaused) {
		response.ErrorResponse(c, http.StatusConflict, "CAN_ONLY_RESUME_PAUSED", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrAlreadyInTerminalStatus) {
		response.ErrorResponse(c, http.StatusConflict, "ALREADY_IN_TERMINAL_STATUS", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrCanOnlyRenewActive) {
		response.ErrorResponse(c, http.StatusConflict, "CAN_ONLY_RENEW_ACTIVE", err.Error())
		return
	}
	if errors.Is(err, subscriptionerrors.ErrAlreadyExpired) {
		response.ErrorResponse(c, http.StatusConflict, "ALREADY_EXPIRED", err.Error())
		return
	}

	// 500 Internal Server Error - Technical Errors (hide details)
	if errors.Is(err, subscriptionerrors.ErrQueryFailed) ||
		errors.Is(err, subscriptionerrors.ErrCreateFailed) ||
		errors.Is(err, subscriptionerrors.ErrUpdateFailed) ||
		errors.Is(err, subscriptionerrors.ErrDeleteFailed) {
		response.ErrorResponse(c, http.StatusInternalServerError, defaultCode, defaultMessage)
		return
	}

	// Default: 500 Internal Server Error
	response.ErrorResponse(c, http.StatusInternalServerError, defaultCode, defaultMessage)
}

// Create creates subscription
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	customerID, err := uuidv7.Parse(req.CustomerID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	billingPeriod := aggregate.BillingPeriod(req.BillingPeriod)

	sub, err := h.u.CreateSubscription(c.Request.Context(), customerID, req.PlanID, billingPeriod, req.Currency, req.Amount, req.TrialDays)
	if err != nil {
		h.handleError(c, err, "CREATE_FAILED", "Failed to create subscription")
		return
	}

	response.Created(c, dto.ToSubscriptionResponse(sub))
}

// GetByID retrieves subscription
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	sub, err := h.u.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "GET_FAILED", "Failed to get subscription")
		return
	}

	response.Success(c, dto.ToSubscriptionResponse(sub))
}

// Update updates subscription
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	var req dto.UpdateSubscriptionRequest
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

	if err := h.u.UpdateSubscription(c.Request.Context(), id, planID, amount); err != nil {
		h.handleError(c, err, "UPDATE_FAILED", "Failed to update subscription")
		return
	}

	sub, err := h.u.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, dto.ToSubscriptionResponse(sub))
}

// Delete soft-deletes subscription
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	if err := h.u.DeleteSubscription(c.Request.Context(), id); err != nil {
		h.handleError(c, err, "DELETE_FAILED", "Failed to delete subscription")
		return
	}

	response.Success(c, gin.H{"message": "Subscription deleted successfully"})
}

// List retrieves subscriptions
func (h *SubscriptionHandler) List(c *gin.Context) {
	var req dto.ListSubscriptionsRequest
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

	var subscriptions []*aggregate.Subscription
	var err error

	if req.CustomerID != "" {
		customerID, parseErr := uuidv7.Parse(req.CustomerID)
		if parseErr != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", parseErr.Error())
			return
		}
		subscriptions, err = h.u.ListByCustomer(c.Request.Context(), customerID)
	} else if req.Status != "" {
		status := aggregate.SubscriptionStatus(req.Status)
		subscriptions, err = h.u.ListByStatus(c.Request.Context(), status)
	} else {
		subscriptions, err = h.u.ListSubscriptions(c.Request.Context(), req.Page, req.PageSize)
	}

	if err != nil {
		h.handleError(c, err, "LIST_FAILED", "Failed to list subscriptions")
		return
	}

	result := make([]dto.SubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		result[i] = dto.ToSubscriptionResponse(sub)
	}

	response.Success(c, result)
}

// Activate activates subscription
func (h *SubscriptionHandler) Activate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	if err := h.u.ActivateSubscription(c.Request.Context(), id); err != nil {
		h.handleError(c, err, "ACTIVATE_FAILED", "Failed to activate subscription")
		return
	}

	sub, err := h.u.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, dto.ToSubscriptionResponse(sub))
}

// Pause pauses subscription
func (h *SubscriptionHandler) Pause(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	if err := h.u.PauseSubscription(c.Request.Context(), id); err != nil {
		h.handleError(c, err, "PAUSE_FAILED", "Failed to pause subscription")
		return
	}

	sub, err := h.u.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, dto.ToSubscriptionResponse(sub))
}

// Cancel cancels subscription
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID format")
		return
	}

	reason := "User requested cancellation"
	effectiveDate := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

	if err := h.u.CancelSubscription(c.Request.Context(), id, reason, effectiveDate); err != nil {
		h.handleError(c, err, "CANCEL_FAILED", "Failed to cancel subscription")
		return
	}

	sub, err := h.u.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, dto.ToSubscriptionResponse(sub))
}
