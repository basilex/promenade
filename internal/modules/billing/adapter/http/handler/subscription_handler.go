package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/billing/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/billing/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type SubscriptionHandler struct {
	subscriptionUC usecase.ISubscriptionUseCase
}

func NewSubscriptionHandler(subscriptionUC usecase.ISubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionUC: subscriptionUC}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", nil)
		return
	}

	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	planID, _ := uuidv7.Parse(req.PlanID)
	subscription, err := h.subscriptionUC.CreateSubscription(c.Request.Context(), userID.(uuidv7.UUID), planID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToSubscriptionResponse(subscription))
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idStr := c.Param("id")
	subscriptionID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_SUBSCRIPTION_ID", err)
		return
	}

	subscription, err := h.subscriptionUC.GetSubscription(c.Request.Context(), subscriptionID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToSubscriptionResponse(subscription))
}

func (h *SubscriptionHandler) GetMySubscriptions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", nil)
		return
	}

	subscriptions, err := h.subscriptionUC.GetUserSubscriptions(c.Request.Context(), userID.(uuidv7.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToSubscriptionListResponse(subscriptions))
}

func (h *SubscriptionHandler) GetMyActiveSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", nil)
		return
	}

	subscription, err := h.subscriptionUC.GetActiveSubscription(c.Request.Context(), userID.(uuidv7.UUID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "NO_ACTIVE_SUBSCRIPTION", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToSubscriptionResponse(subscription))
}

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	subscriptions, err := h.subscriptionUC.ListSubscriptions(c.Request.Context(), nil, 100, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToSubscriptionListResponse(subscriptions))
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", nil)
		return
	}

	idStr := c.Param("id")
	subscriptionID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_SUBSCRIPTION_ID", err)
		return
	}

	if err := h.subscriptionUC.CancelSubscription(c.Request.Context(), userID.(uuidv7.UUID), subscriptionID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SubscriptionHandler) UpgradeSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", nil)
		return
	}

	idStr := c.Param("id")
	subscriptionID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_SUBSCRIPTION_ID", err)
		return
	}

	var req dto.UpgradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	newPlanID, _ := uuidv7.Parse(req.NewPlanID)
	subscription, err := h.subscriptionUC.UpgradeSubscription(c.Request.Context(), userID.(uuidv7.UUID), subscriptionID, newPlanID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToSubscriptionResponse(subscription))
}
