package http

import (
    "errors"
    "strconv"

    "github.com/basilex/promenade/internal/contexts/accounting/budget"
    "github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/budget/dto"
    "github.com/basilex/promenade/internal/contexts/accounting/budget/usecase"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/gin-gonic/gin"
)

type BudgetHandler struct {
    uc usecase.IBudgetUseCase
}

func NewBudgetHandler(uc usecase.IBudgetUseCase) *BudgetHandler {
    return &BudgetHandler{uc: uc}
}

func (h *BudgetHandler) RegisterRoutes(router *gin.RouterGroup) {
    budgets := router.Group("/budgets")
    {
        budgets.POST("", h.CreateBudget)
        budgets.GET("/:id", h.GetBudget)
        budgets.GET("/name/:name", h.GetBudgetByName)
        budgets.POST("/:id/lines", h.AddLine)
        budgets.PUT("/:id/lines/:lineId", h.UpdateLine)
        budgets.DELETE("/:id/lines/:lineId", h.RemoveLine)
        budgets.POST("/:id/approve", h.ApproveBudget)
        budgets.POST("/:id/activate", h.ActivateBudget)
        budgets.POST("/:id/close", h.CloseBudget)
        budgets.DELETE("/:id", h.DeleteBudget)
        budgets.GET("", h.ListBudgets)
        budgets.GET("/status/:status", h.ListBudgetsByStatus)
        budgets.GET("/active", h.ListActiveBudgets)
    }
}

// CreateBudget creates a new budget
// @Summary Create a new budget
// @Tags budgets
// @Accept json
// @Produce json
// @Param request body dto.CreateBudgetRequest true "Create budget request"
// @Success 201 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets [post]
func (h *BudgetHandler) CreateBudget(c *gin.Context) {
    var req dto.CreateBudgetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))
    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.CreateBudget(
        c.Request.Context(),
        organizationID,
        req.Name,
        req.FiscalYear,
        userUUID,
    )
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Created(c, dto.ToBudgetResponse(b))
}

// GetBudget gets a budget by ID
// @Summary Get a budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id} [get]
func (h *BudgetHandler) GetBudget(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    b, err := h.uc.GetBudgetByID(c.Request.Context(), id)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// GetBudgetByName gets a budget by name
// @Summary Get a budget by name
// @Tags budgets
// @Produce json
// @Param name path string true "Budget name"
// @Success 200 {object} dto.BudgetResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/name/{name} [get]
func (h *BudgetHandler) GetBudgetByName(c *gin.Context) {
    name := c.Param("name")
    if name == "" {
        response.BadRequest(c, "name is required")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    b, err := h.uc.GetBudgetByName(c.Request.Context(), organizationID, name)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// AddLine adds a line to the budget
// @Summary Add line to budget
// @Tags budgets
// @Accept json
// @Produce json
// @Param id path string true "Budget ID"
// @Param request body dto.AddLineRequest true "Add line request"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id}/lines [post]
func (h *BudgetHandler) AddLine(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    var req dto.AddLineRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    accountID, err := uuidv7.Parse(req.AccountID)
    if err != nil {
        response.BadRequest(c, "invalid account_id")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.AddLine(
        c.Request.Context(),
        id,
        accountID,
        req.BudgetAmount,
        req.Description,
        userUUID,
    )
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// UpdateLine updates a budget line
// @Summary Update budget line
// @Tags budgets
// @Accept json
// @Produce json
// @Param id path string true "Budget ID"
// @Param lineId path string true "Line ID"
// @Param request body dto.UpdateLineRequest true "Update line request"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id}/lines/{lineId} [put]
func (h *BudgetHandler) UpdateLine(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    lineID, err := uuidv7.Parse(c.Param("lineId"))
    if err != nil {
        response.BadRequest(c, "invalid line ID")
        return
    }

    var req dto.UpdateLineRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.UpdateLine(c.Request.Context(), id, lineID, req.BudgetAmount, userUUID)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// RemoveLine removes a line from the budget
// @Summary Remove line from budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Param lineId path string true "Line ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id}/lines/{lineId} [delete]
func (h *BudgetHandler) RemoveLine(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    lineID, err := uuidv7.Parse(c.Param("lineId"))
    if err != nil {
        response.BadRequest(c, "invalid line ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.RemoveLine(c.Request.Context(), id, lineID, userUUID)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// ApproveBudget approves a budget
// @Summary Approve budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id}/approve [post]
func (h *BudgetHandler) ApproveBudget(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.ApproveBudget(c.Request.Context(), id, userUUID)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// ActivateBudget activates a budget
// @Summary Activate budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id}/activate [post]
func (h *BudgetHandler) ActivateBudget(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.ActivateBudget(c.Request.Context(), id, userUUID)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// CloseBudget closes a budget
// @Summary Close budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id}/close [post]
func (h *BudgetHandler) CloseBudget(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    b, err := h.uc.CloseBudget(c.Request.Context(), id, userUUID)
    if err != nil {
        handleBudgetError(c, err)
        return
    }

    response.Success(c, dto.ToBudgetResponse(b))
}

// DeleteBudget deletes a budget
// @Summary Delete a budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/{id} [delete]
func (h *BudgetHandler) DeleteBudget(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid budget ID")
        return
    }

    if err := h.uc.DeleteBudget(c.Request.Context(), id); err != nil {
        handleBudgetError(c, err)
        return
    }

    response.SuccessWithMessage(c, "budget deleted successfully")
}

// ListBudgets lists all budgets for an organization
// @Summary List budgets
// @Tags budgets
// @Produce json
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} dto.BudgetResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets [get]
func (h *BudgetHandler) ListBudgets(c *gin.Context) {
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

    budgets, err := h.uc.ListBudgetsByOrganization(c.Request.Context(), organizationID, limit, offset)
    if err != nil {
        response.InternalError(c, "failed to list budgets")
        return
    }

    response.Success(c, dto.ToBudgetResponseList(budgets))
}

// ListBudgetsByStatus lists budgets by status
// @Summary List budgets by status
// @Tags budgets
// @Produce json
// @Param status path string true "Budget status"
// @Success 200 {array} dto.BudgetResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/status/{status} [get]
func (h *BudgetHandler) ListBudgetsByStatus(c *gin.Context) {
    statusStr := c.Param("status")
    status := aggregate.BudgetStatus(statusStr)
    if !isValidBudgetStatus(status) {
        response.BadRequest(c, "invalid budget status")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    budgets, err := h.uc.ListBudgetsByStatus(c.Request.Context(), organizationID, status)
    if err != nil {
        response.InternalError(c, "failed to list budgets by status")
        return
    }

    response.Success(c, dto.ToBudgetResponseList(budgets))
}

// ListActiveBudgets lists all active budgets
// @Summary List active budgets
// @Tags budgets
// @Produce json
// @Success 200 {array} dto.BudgetResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/budgets/active [get]
func (h *BudgetHandler) ListActiveBudgets(c *gin.Context) {
    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    budgets, err := h.uc.ListActiveBudgets(c.Request.Context(), organizationID)
    if err != nil {
        response.InternalError(c, "failed to list active budgets")
        return
    }

    response.Success(c, dto.ToBudgetResponseList(budgets))
}

func handleBudgetError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, budget.ErrBudgetNotFound):
        response.NotFound(c, "budget not found")
    case errors.Is(err, budget.ErrBudgetNameEmpty),
        errors.Is(err, budget.ErrInvalidBudgetPeriod),
        errors.Is(err, budget.ErrInvalidAmount),
        errors.Is(err, budget.ErrLineAccountRequired),
        errors.Is(err, budget.ErrInvalidFiscalYear),
        errors.Is(err, budget.ErrAccountRequired),
        errors.Is(err, budget.ErrInvalidBudgetAmount),
        errors.Is(err, budget.ErrAccountAlreadyExists),
        errors.Is(err, budget.ErrBudgetAlreadyApproved),
        errors.Is(err, budget.ErrBudgetNotApproved),
        errors.Is(err, budget.ErrCannotModifyActive),
        errors.Is(err, budget.ErrCannotDeleteApproved),
        errors.Is(err, budget.ErrLineNotFound),
        errors.Is(err, budget.ErrCannotModifyApprovedBudget),
        errors.Is(err, budget.ErrCannotModifyActiveBudget),
        errors.Is(err, budget.ErrBudgetLineNotFound),
        errors.Is(err, budget.ErrBudgetNotActive),
        errors.Is(err, budget.ErrBudgetHasNoLines):
        response.BadRequest(c, err.Error())
    default:
        response.InternalError(c, "operation failed")
    }
}

func isValidBudgetStatus(s aggregate.BudgetStatus) bool {
    switch s {
    case aggregate.BudgetStatusDraft,
        aggregate.BudgetStatusApproved,
        aggregate.BudgetStatusActive,
        aggregate.BudgetStatusClosed:
        return true
    default:
        return false
    }
}