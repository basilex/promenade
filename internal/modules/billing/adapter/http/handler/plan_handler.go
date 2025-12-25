package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/billing/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PlanHandler struct {
	planUC usecase.IPlanUseCase
}

func NewPlanHandler(planUC usecase.IPlanUseCase) *PlanHandler {
	return &PlanHandler{planUC: planUC}
}

func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req dto.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	plan, err := h.planUC.CreatePlan(c.Request.Context(), req.Name, req.Slug, req.Description, req.Currency, req.Price, entity.PlanInterval(req.Interval), req.Features)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToPlanResponse(plan))
}

func (h *PlanHandler) GetPlan(c *gin.Context) {
	idStr := c.Param("id")
	planID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", err)
		return
	}

	plan, err := h.planUC.GetPlan(c.Request.Context(), planID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "PLAN_NOT_FOUND", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPlanResponse(plan))
}

func (h *PlanHandler) ListPlans(c *gin.Context) {
	plans, err := h.planUC.ListPlans(c.Request.Context(), nil, 100, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPlanListResponse(plans))
}

func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	idStr := c.Param("id")
	planID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", err)
		return
	}

	var req dto.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	updates := make(map[string]any)
	plan, err := h.planUC.UpdatePlan(c.Request.Context(), planID, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPlanResponse(plan))
}

func (h *PlanHandler) ActivatePlan(c *gin.Context) {
	idStr := c.Param("id")
	planID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", err)
		return
	}

	if err := h.planUC.ActivatePlan(c.Request.Context(), planID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PlanHandler) DeactivatePlan(c *gin.Context) {
	idStr := c.Param("id")
	planID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", err)
		return
	}

	if err := h.planUC.DeactivatePlan(c.Request.Context(), planID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PlanHandler) DeletePlan(c *gin.Context) {
	idStr := c.Param("id")
	planID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", err)
		return
	}

	if err := h.planUC.DeletePlan(c.Request.Context(), planID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}
