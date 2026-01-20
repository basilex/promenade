package http

import (
    "errors"
    "strconv"

    "github.com/basilex/promenade/internal/contexts/accounting/costcenter"
    "github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/costcenter/dto"
    "github.com/basilex/promenade/internal/contexts/accounting/costcenter/usecase"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/gin-gonic/gin"
)

type CostCenterHandler struct {
    uc usecase.ICostCenterUseCase
}

func NewCostCenterHandler(uc usecase.ICostCenterUseCase) *CostCenterHandler {
    return &CostCenterHandler{uc: uc}
}

func (h *CostCenterHandler) RegisterRoutes(router *gin.RouterGroup) {
    costCenters := router.Group("/cost-centers")
    {
        costCenters.POST("", h.CreateCostCenter)
        costCenters.GET("/:id", h.GetCostCenter)
        costCenters.GET("/code/:code", h.GetCostCenterByCode)
        costCenters.PUT("/:id/parent", h.SetParent)
        costCenters.PUT("/:id/manager", h.SetManager)
        costCenters.PUT("/:id/activate", h.ActivateCostCenter)
        costCenters.PUT("/:id/deactivate", h.DeactivateCostCenter)
        costCenters.DELETE("/:id", h.DeleteCostCenter)
        costCenters.GET("", h.ListCostCenters)
        costCenters.GET("/children/:parentId", h.ListChildren)
        costCenters.GET("/type/:type", h.ListCostCentersByType)
        costCenters.GET("/active", h.ListActiveCostCenters)
    }
}

// CreateCostCenter creates a new cost center
// @Summary Create a new cost center
// @Tags cost-centers
// @Accept json
// @Produce json
// @Param request body dto.CreateCostCenterRequest true "Create cost center request"
// @Success 201 {object} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers [post]
func (h *CostCenterHandler) CreateCostCenter(c *gin.Context) {
    var req dto.CreateCostCenterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    centerType := aggregate.CenterType(req.CenterType)
    if !isValidCenterType(centerType) {
        response.BadRequest(c, "invalid center type")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))
    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    cc, err := h.uc.CreateCostCenter(
        c.Request.Context(),
        organizationID,
        req.Code,
        req.Name,
        centerType,
        req.ParentID,
        userUUID,
    )
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Created(c, dto.ToCostCenterResponse(cc))
}

// GetCostCenter gets a cost center by ID
// @Summary Get a cost center
// @Tags cost-centers
// @Produce json
// @Param id path string true "Cost Center ID"
// @Success 200 {object} dto.CostCenterResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/{id} [get]
func (h *CostCenterHandler) GetCostCenter(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid cost center ID")
        return
    }

    cc, err := h.uc.GetCostCenterByID(c.Request.Context(), id)
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Success(c, dto.ToCostCenterResponse(cc))
}

// GetCostCenterByCode gets a cost center by code
// @Summary Get a cost center by code
// @Tags cost-centers
// @Produce json
// @Param code path string true "Cost center code"
// @Success 200 {object} dto.CostCenterResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/code/{code} [get]
func (h *CostCenterHandler) GetCostCenterByCode(c *gin.Context) {
    code := c.Param("code")
    if code == "" {
        response.BadRequest(c, "code is required")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    cc, err := h.uc.GetCostCenterByCode(c.Request.Context(), organizationID, code)
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Success(c, dto.ToCostCenterResponse(cc))
}

// SetParent sets the parent of a cost center
// @Summary Set parent cost center
// @Tags cost-centers
// @Accept json
// @Produce json
// @Param id path string true "Cost Center ID"
// @Param request body dto.SetParentRequest true "Set parent request"
// @Success 200 {object} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/{id}/parent [put]
func (h *CostCenterHandler) SetParent(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid cost center ID")
        return
    }

    var req dto.SetParentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    cc, err := h.uc.SetParent(c.Request.Context(), id, req.ParentID, userUUID)
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Success(c, dto.ToCostCenterResponse(cc))
}

// SetManager sets the manager of a cost center
// @Summary Set manager for cost center
// @Tags cost-centers
// @Accept json
// @Produce json
// @Param id path string true "Cost Center ID"
// @Param request body dto.SetManagerRequest true "Set manager request"
// @Success 200 {object} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/{id}/manager [put]
func (h *CostCenterHandler) SetManager(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid cost center ID")
        return
    }

    var req dto.SetManagerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    cc, err := h.uc.SetManager(c.Request.Context(), id, req.ManagerID, userUUID)
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Success(c, dto.ToCostCenterResponse(cc))
}

// ActivateCostCenter activates a cost center
// @Summary Activate a cost center
// @Tags cost-centers
// @Produce json
// @Param id path string true "Cost Center ID"
// @Success 200 {object} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/{id}/activate [put]
func (h *CostCenterHandler) ActivateCostCenter(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid cost center ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    cc, err := h.uc.ActivateCostCenter(c.Request.Context(), id, userUUID)
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Success(c, dto.ToCostCenterResponse(cc))
}

// DeactivateCostCenter deactivates a cost center
// @Summary Deactivate a cost center
// @Tags cost-centers
// @Produce json
// @Param id path string true "Cost Center ID"
// @Success 200 {object} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/{id}/deactivate [put]
func (h *CostCenterHandler) DeactivateCostCenter(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid cost center ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    cc, err := h.uc.DeactivateCostCenter(c.Request.Context(), id, userUUID)
    if err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.Success(c, dto.ToCostCenterResponse(cc))
}

// DeleteCostCenter deletes a cost center
// @Summary Delete a cost center
// @Tags cost-centers
// @Produce json
// @Param id path string true "Cost Center ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/{id} [delete]
func (h *CostCenterHandler) DeleteCostCenter(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid cost center ID")
        return
    }

    if err := h.uc.DeleteCostCenter(c.Request.Context(), id); err != nil {
        handleCostCenterError(c, err)
        return
    }

    response.SuccessWithMessage(c, "cost center deleted successfully")
}

// ListCostCenters lists all cost centers for an organization
// @Summary List cost centers
// @Tags cost-centers
// @Produce json
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} dto.CostCenterResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers [get]
func (h *CostCenterHandler) ListCostCenters(c *gin.Context) {
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

    costCenters, err := h.uc.ListCostCentersByOrganization(c.Request.Context(), organizationID, limit, offset)
    if err != nil {
        response.InternalError(c, "failed to list cost centers")
        return
    }

    response.Success(c, dto.ToCostCenterResponseList(costCenters))
}

// ListChildren lists child cost centers
// @Summary List child cost centers
// @Tags cost-centers
// @Produce json
// @Param parentId path string true "Parent cost center ID"
// @Success 200 {array} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/children/{parentId} [get]
func (h *CostCenterHandler) ListChildren(c *gin.Context) {
    parentID, err := uuidv7.Parse(c.Param("parentId"))
    if err != nil {
        response.BadRequest(c, "invalid parent ID")
        return
    }

    children, err := h.uc.ListChildCostCenters(c.Request.Context(), parentID)
    if err != nil {
        response.InternalError(c, "failed to list child cost centers")
        return
    }

    response.Success(c, dto.ToCostCenterResponseList(children))
}

// ListCostCentersByType lists cost centers by type
// @Summary List cost centers by type
// @Tags cost-centers
// @Produce json
// @Param type path string true "Center type"
// @Success 200 {array} dto.CostCenterResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/type/{type} [get]
func (h *CostCenterHandler) ListCostCentersByType(c *gin.Context) {
    typeStr := c.Param("type")
    centerType := aggregate.CenterType(typeStr)
    if !isValidCenterType(centerType) {
        response.BadRequest(c, "invalid center type")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    costCenters, err := h.uc.ListCostCentersByType(c.Request.Context(), organizationID, centerType)
    if err != nil {
        response.InternalError(c, "failed to list cost centers by type")
        return
    }

    response.Success(c, dto.ToCostCenterResponseList(costCenters))
}

// ListActiveCostCenters lists all active cost centers
// @Summary List active cost centers
// @Tags cost-centers
// @Produce json
// @Success 200 {array} dto.CostCenterResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/cost-centers/active [get]
func (h *CostCenterHandler) ListActiveCostCenters(c *gin.Context) {
    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    costCenters, err := h.uc.ListActiveCostCenters(c.Request.Context(), organizationID)
    if err != nil {
        response.InternalError(c, "failed to list active cost centers")
        return
    }

    response.Success(c, dto.ToCostCenterResponseList(costCenters))
}

func handleCostCenterError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, costcenter.ErrCostCenterNotFound):
        response.NotFound(c, "cost center not found")
    case errors.Is(err, costcenter.ErrCodeRequired),
        errors.Is(err, costcenter.ErrNameRequired),
        errors.Is(err, costcenter.ErrInvalidCenterType),
        errors.Is(err, costcenter.ErrCannotBeOwnParent),
        errors.Is(err, costcenter.ErrCodeAlreadyExists),
        errors.Is(err, costcenter.ErrHasChildren):
        response.BadRequest(c, err.Error())
    default:
        response.InternalError(c, "operation failed")
    }
}

func isValidCenterType(t aggregate.CenterType) bool {
    switch t {
    case aggregate.CenterTypeCost,
        aggregate.CenterTypeProfit,
        aggregate.CenterTypeInvestment:
        return true
    default:
        return false
    }
}