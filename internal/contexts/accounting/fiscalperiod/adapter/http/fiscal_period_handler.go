package http

import (
    "errors"
    "strconv"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/dto"
    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/usecase"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/gin-gonic/gin"
)

type FiscalPeriodHandler struct {
    uc usecase.IFiscalPeriodUseCase
}

func NewFiscalPeriodHandler(uc usecase.IFiscalPeriodUseCase) *FiscalPeriodHandler {
    return &FiscalPeriodHandler{uc: uc}
}

func (h *FiscalPeriodHandler) RegisterRoutes(router *gin.RouterGroup) {
    periods := router.Group("/fiscal-periods")
    {
        periods.POST("", h.CreateFiscalPeriod)
        periods.GET("/:id", h.GetFiscalPeriod)
        periods.GET("/current", h.GetCurrentPeriod)
        periods.PUT("/:id/close", h.ClosePeriod)
        periods.PUT("/:id/reopen", h.ReopenPeriod)
        periods.PUT("/:id/lock", h.LockPeriod)
        periods.DELETE("/:id", h.DeleteFiscalPeriod)
        periods.GET("", h.ListFiscalPeriods)
        periods.GET("/year/:year", h.ListFiscalPeriodsByYear)
        periods.GET("/open", h.ListOpenPeriods)
    }
}

// CreateFiscalPeriod creates a new fiscal period
// @Summary Create a new fiscal period
// @Tags fiscal-periods
// @Accept json
// @Produce json
// @Param request body dto.CreateFiscalPeriodRequest true "Create fiscal period request"
// @Success 201 {object} dto.FiscalPeriodResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods [post]
func (h *FiscalPeriodHandler) CreateFiscalPeriod(c *gin.Context) {
    var req dto.CreateFiscalPeriodRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    periodType := aggregate.PeriodType(req.PeriodType)
    if !isValidPeriodType(periodType) {
        response.BadRequest(c, "invalid period type")
        return
    }

    startDate, err := time.Parse("2006-01-02", req.StartDate)
    if err != nil {
        response.BadRequest(c, "invalid start_date format, expected YYYY-MM-DD")
        return
    }

    endDate, err := time.Parse("2006-01-02", req.EndDate)
    if err != nil {
        response.BadRequest(c, "invalid end_date format, expected YYYY-MM-DD")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))
    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    fp, err := h.uc.CreateFiscalPeriod(
        c.Request.Context(),
        organizationID,
        req.Code,
        req.Name,
        periodType,
        startDate,
        endDate,
        userUUID,
    )
    if err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.Created(c, dto.ToFiscalPeriodResponse(fp))
}

// GetFiscalPeriod gets a fiscal period by ID
// @Summary Get a fiscal period
// @Tags fiscal-periods
// @Produce json
// @Param id path string true "Fiscal Period ID"
// @Success 200 {object} dto.FiscalPeriodResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/{id} [get]
func (h *FiscalPeriodHandler) GetFiscalPeriod(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid fiscal period ID")
        return
    }

    fp, err := h.uc.GetFiscalPeriodByID(c.Request.Context(), id)
    if err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponse(fp))
}

// GetCurrentPeriod gets the current fiscal period for a date
// @Summary Get current fiscal period
// @Tags fiscal-periods
// @Produce json
// @Param date query string false "Date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} dto.FiscalPeriodResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/current [get]
func (h *FiscalPeriodHandler) GetCurrentPeriod(c *gin.Context) {
    dateStr := c.Query("date")
    if dateStr == "" {
        dateStr = time.Now().Format("2006-01-02")
    }

    _, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        response.BadRequest(c, "invalid date format, expected YYYY-MM-DD")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    fp, err := h.uc.GetCurrentPeriod(c.Request.Context(), organizationID, dateStr)
    if err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponse(fp))
}

// ClosePeriod closes a fiscal period
// @Summary Close a fiscal period
// @Tags fiscal-periods
// @Produce json
// @Param id path string true "Fiscal Period ID"
// @Success 200 {object} dto.FiscalPeriodResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/{id}/close [put]
func (h *FiscalPeriodHandler) ClosePeriod(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid fiscal period ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    fp, err := h.uc.ClosePeriod(c.Request.Context(), id, userUUID)
    if err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponse(fp))
}

// ReopenPeriod reopens a fiscal period
// @Summary Reopen a fiscal period
// @Tags fiscal-periods
// @Produce json
// @Param id path string true "Fiscal Period ID"
// @Success 200 {object} dto.FiscalPeriodResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/{id}/reopen [put]
func (h *FiscalPeriodHandler) ReopenPeriod(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid fiscal period ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    fp, err := h.uc.ReopenPeriod(c.Request.Context(), id, userUUID)
    if err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponse(fp))
}

// LockPeriod locks a fiscal period
// @Summary Lock a fiscal period
// @Tags fiscal-periods
// @Produce json
// @Param id path string true "Fiscal Period ID"
// @Success 200 {object} dto.FiscalPeriodResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/{id}/lock [put]
func (h *FiscalPeriodHandler) LockPeriod(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid fiscal period ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    fp, err := h.uc.LockPeriod(c.Request.Context(), id, userUUID)
    if err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponse(fp))
}

// DeleteFiscalPeriod deletes a fiscal period
// @Summary Delete a fiscal period
// @Tags fiscal-periods
// @Produce json
// @Param id path string true "Fiscal Period ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/{id} [delete]
func (h *FiscalPeriodHandler) DeleteFiscalPeriod(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid fiscal period ID")
        return
    }

    if err := h.uc.DeleteFiscalPeriod(c.Request.Context(), id); err != nil {
        handleFiscalPeriodError(c, err)
        return
    }

    response.SuccessWithMessage(c, "fiscal period deleted successfully")
}

// ListFiscalPeriods lists all fiscal periods for an organization
// @Summary List fiscal periods
// @Tags fiscal-periods
// @Produce json
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} dto.FiscalPeriodResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods [get]
func (h *FiscalPeriodHandler) ListFiscalPeriods(c *gin.Context) {
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

    periods, err := h.uc.ListFiscalPeriodsByOrganization(c.Request.Context(), organizationID, limit, offset)
    if err != nil {
        response.InternalError(c, "failed to list fiscal periods")
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponseList(periods))
}

// ListFiscalPeriodsByYear lists fiscal periods for a specific year
// @Summary List fiscal periods by year
// @Tags fiscal-periods
// @Produce json
// @Param year path int true "Fiscal year"
// @Success 200 {array} dto.FiscalPeriodResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/year/{year} [get]
func (h *FiscalPeriodHandler) ListFiscalPeriodsByYear(c *gin.Context) {
    yearStr := c.Param("year")
    year, err := strconv.Atoi(yearStr)
    if err != nil || year < 1900 || year > 2100 {
        response.BadRequest(c, "invalid year")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    periods, err := h.uc.ListFiscalPeriodsByYear(c.Request.Context(), organizationID, year)
    if err != nil {
        response.InternalError(c, "failed to list fiscal periods by year")
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponseList(periods))
}

// ListOpenPeriods lists all open fiscal periods
// @Summary List open fiscal periods
// @Tags fiscal-periods
// @Produce json
// @Success 200 {array} dto.FiscalPeriodResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/fiscal-periods/open [get]
func (h *FiscalPeriodHandler) ListOpenPeriods(c *gin.Context) {
    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    periods, err := h.uc.ListOpenPeriods(c.Request.Context(), organizationID)
    if err != nil {
        response.InternalError(c, "failed to list open periods")
        return
    }

    response.Success(c, dto.ToFiscalPeriodResponseList(periods))
}

func handleFiscalPeriodError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, fiscalperiod.ErrPeriodNotFound):
        response.NotFound(c, "fiscal period not found")
    case errors.Is(err, fiscalperiod.ErrPeriodCodeEmpty),
        errors.Is(err, fiscalperiod.ErrPeriodNameEmpty),
        errors.Is(err, fiscalperiod.ErrInvalidPeriodType),
        errors.Is(err, fiscalperiod.ErrInvalidDateRange),
        errors.Is(err, fiscalperiod.ErrPeriodAlreadyClosed),
        errors.Is(err, fiscalperiod.ErrPeriodAlreadyLocked),
        errors.Is(err, fiscalperiod.ErrPeriodNotClosed),
        errors.Is(err, fiscalperiod.ErrCannotReopenLocked),
        errors.Is(err, fiscalperiod.ErrPeriodOverlap):
        response.BadRequest(c, err.Error())
    default:
        response.InternalError(c, "operation failed")
    }
}

func isValidPeriodType(t aggregate.PeriodType) bool {
    switch t {
    case aggregate.PeriodTypeMonth,
        aggregate.PeriodTypeQuarter,
        aggregate.PeriodTypeYear:
        return true
    default:
        return false
    }
}