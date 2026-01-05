package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// DealHandler handles HTTP requests for deals
type DealHandler struct {
	dealUC deal.IUseCase
}

// NewDealHandler creates a new deal handler
func NewDealHandler(dealUC deal.IUseCase) *DealHandler {
	return &DealHandler{
		dealUC: dealUC,
	}
}

// Create godoc
// @Summary Create a new deal
// @Description Create a new deal in the sales pipeline
// @Tags Deals
// @Accept json
// @Produce json
// @Param body body CreateDealRequest true "Deal creation data"
// @Success 201 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals [post]
// @Security Bearer
func (h *DealHandler) Create(c *gin.Context) {
	var req CreateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Parse customer ID (required)
	customerID, err := parseCustomerID(req.CustomerID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", err.Error())
		return
	}

	// Parse optional fields
	assignedTo, err := parseAssignedTo(req.AssignedTo)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ASSIGNED_TO", err.Error())
		return
	}

	// Create deal
	var assignedToID uuidv7.UUID
	if assignedTo != nil {
		assignedToID = *assignedTo
	}
	d, err := h.dealUC.CreateDeal(
		c.Request.Context(),
		req.Name,
		customerID,
		assignedToID,
		req.Value,
		req.Currency,
		req.ExpectedCloseDate,
	)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_DEAL_FAILED", err.Error())
		return
	}

	// Set optional fields
	if req.Source != nil {
		if _, err := h.dealUC.SetDealSource(c.Request.Context(), d.ID, deal.DealSource(*req.Source)); err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "SET_SOURCE_FAILED", err.Error())
			return
		}
	}

	response.Created(c, ToDealResponse(d))
}

// GetByID godoc
// @Summary Get deal by ID
// @Description Get a single deal by ID
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Success 200 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id} [get]
// @Security Bearer
func (h *DealHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	d, err := h.dealUC.GetDeal(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_DEAL_FAILED", err.Error())
		return
	}

	response.Success(c, ToDealResponse(d))
}

// UpdateBasicInfo godoc
// @Summary Update deal basic info
// @Description Update deal name and expected close date
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param body body UpdateDealBasicInfoRequest true "Update data"
// @Success 200 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id}/basic-info [put]
// @Security Bearer
func (h *DealHandler) UpdateBasicInfo(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	var req UpdateDealBasicInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Dereference pointers or use empty strings
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	description := "" // No description field in DTO, so empty
	
	d, err := h.dealUC.UpdateDealBasicInfo(c.Request.Context(), id, name, description)
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_DEAL_FAILED", err.Error())
		return
	}

	response.Success(c, ToDealResponse(d))
}

// UpdateValue godoc
// @Summary Update deal value
// @Description Update deal value and currency
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param body body UpdateDealValueRequest true "Update data"
// @Success 200 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id}/value [put]
// @Security Bearer
func (h *DealHandler) UpdateValue(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	var req UpdateDealValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	d, err := h.dealUC.UpdateDealValue(c.Request.Context(), id, req.Value, req.Currency)
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_DEAL_VALUE_FAILED", err.Error())
		return
	}

	response.Success(c, ToDealResponse(d))
}

// MoveToStage godoc
// @Summary Move deal to stage
// @Description Move deal to a new stage in the pipeline
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param body body MoveDealToStageRequest true "Stage data"
// @Success 200 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id}/stage [put]
// @Security Bearer
func (h *DealHandler) MoveToStage(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	var req MoveDealToStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	d, err := h.dealUC.MoveDealToStage(c.Request.Context(), id, deal.DealStage(req.Stage))
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "MOVE_DEAL_FAILED", err.Error())
		return
	}

	response.Success(c, ToDealResponse(d))
}

// MarkAsWon godoc
// @Summary Mark deal as won
// @Description Mark a deal as won with close date
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param body body MarkDealAsWonRequest true "Close data"
// @Success 200 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id}/won [post]
// @Security Bearer
func (h *DealHandler) MarkAsWon(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	var req MarkDealAsWonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	d, err := h.dealUC.MarkDealAsWon(c.Request.Context(), id, req.CloseDate)
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "MARK_WON_FAILED", err.Error())
		return
	}

	response.Success(c, ToDealResponse(d))
}

// MarkAsLost godoc
// @Summary Mark deal as lost
// @Description Mark a deal as lost with close date and reason
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param body body MarkDealAsLostRequest true "Close data"
// @Success 200 {object} DealResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id}/lost [post]
// @Security Bearer
func (h *DealHandler) MarkAsLost(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	var req MarkDealAsLostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	reason := ""
	if req.LostReason != nil {
		reason = *req.LostReason
	}
	d, err := h.dealUC.MarkDealAsLost(c.Request.Context(), id, reason)
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "MARK_LOST_FAILED", err.Error())
		return
	}

	response.Success(c, ToDealResponse(d))
}

// Delete godoc
// @Summary Delete deal
// @Description Soft delete a deal
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/{id} [delete]
// @Security Bearer
func (h *DealHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid deal ID format")
		return
	}

	err = h.dealUC.DeleteDeal(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, deal.ErrDealNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "DEAL_NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_DEAL_FAILED", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary List deals
// @Description Get a paginated list of all deals
// @Tags Deals
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]DealResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals [get]
// @Security Bearer
func (h *DealHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	deals, total, err := h.dealUC.ListDeals(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_DEALS_FAILED", err.Error())
		return
	}

	response.SuccessWithPagination(c, ToDealListResponse(deals), total, page, pageSize)
}

// ListByStage godoc
// @Summary List deals by stage
// @Description Get a paginated list of deals in a specific stage
// @Tags Deals
// @Accept json
// @Produce json
// @Param stage path string true "Deal stage" Enums(lead, qualified, proposal, negotiation, closed_won, closed_lost)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]DealResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/stage/{stage} [get]
// @Security Bearer
func (h *DealHandler) ListByStage(c *gin.Context) {
	stage := deal.DealStage(c.Param("stage"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	deals, total, err := h.dealUC.ListDealsByStage(c.Request.Context(), stage, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_DEALS_FAILED", err.Error())
		return
	}

	response.SuccessWithPagination(c, ToDealListResponse(deals), total, page, pageSize)
}

// GetPipelineStats godoc
// @Summary Get pipeline statistics
// @Description Get count of deals by stage
// @Tags Deals
// @Accept json
// @Produce json
// @Success 200 {object} PipelineStatsResponse
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/stats/pipeline [get]
// @Security Bearer
func (h *DealHandler) GetPipelineStats(c *gin.Context) {
	stats, err := h.dealUC.GetPipelineStats(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_STATS_FAILED", err.Error())
		return
	}

	response.Success(c, ToPipelineStatsResponse(stats))
}

// GetWonDeals godoc
// @Summary Get won deals statistics
// @Description Get count and total value of won deals
// @Tags Deals
// @Accept json
// @Produce json
// @Success 200 {object} WonDealsResponse
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/deals/stats/won [get]
// @Security Bearer
func (h *DealHandler) GetWonDeals(c *gin.Context) {
	count, totalValue, err := h.dealUC.GetWonDeals(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_WON_DEALS_FAILED", err.Error())
		return
	}

	response.Success(c, &WonDealsResponse{
		Count:      count,
		TotalValue: totalValue,
	})
}
