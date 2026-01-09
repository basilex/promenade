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
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create subscription")
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
	if errors.Is(err, subscription.ErrSubscriptionNotFound) {
		response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
		return
	}
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to get subscription")
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
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", "Failed to retrieve subscription")
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
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete subscription")
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
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list subscriptions")
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
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "ACTIVATE_FAILED", "Failed to activate subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", "Failed to retrieve subscription")
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
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "PAUSE_FAILED", "Failed to pause subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", "Failed to retrieve subscription")
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
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "CANCEL_FAILED", "Failed to cancel subscription")
		return
	}

	sub, err := h.usecase.GetSubscription(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", "Failed to retrieve subscription")
		return
	}

	response.Success(c, ToSubscriptionResponse(sub))
}
