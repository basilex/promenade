package http

import (
    "errors"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/journalentry"
    "github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/journalentry/dto"
    "github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/gin-gonic/gin"
)

type JournalEntryHandler struct {
    uc usecase.IJournalEntryUseCase
}

func NewJournalEntryHandler(uc usecase.IJournalEntryUseCase) *JournalEntryHandler {
    return &JournalEntryHandler{uc: uc}
}

func (h *JournalEntryHandler) RegisterRoutes(router *gin.RouterGroup) {
    entries := router.Group("/journal-entries")
    {
        entries.POST("", h.CreateJournalEntry)
        entries.GET("/:id", h.GetJournalEntry)
        entries.PUT("/:id/description", h.UpdateDescription)
        entries.POST("/:id/lines", h.AddLine)
        entries.DELETE("/:id/lines/:lineId", h.RemoveLine)
        entries.POST("/:id/post", h.PostEntry)
        entries.POST("/:id/reverse", h.ReverseEntry)
        entries.DELETE("/:id", h.DeleteJournalEntry)
        entries.GET("", h.ListJournalEntries)
        entries.GET("/period/:periodId", h.ListJournalEntriesByPeriod)
    }
}

// CreateJournalEntry creates a new journal entry
// @Summary Create a new journal entry
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param request body dto.CreateJournalEntryRequest true "Create journal entry request"
// @Success 201 {object} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries [post]
func (h *JournalEntryHandler) CreateJournalEntry(c *gin.Context) {
    var req dto.CreateJournalEntryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    sourceType := aggregate.SourceType(req.SourceType)
    if !isValidSourceType(sourceType) {
        response.BadRequest(c, "invalid source type")
        return
    }

    entryDate, err := time.Parse("2006-01-02", req.EntryDate)
    if err != nil {
        response.BadRequest(c, "invalid entry_date format, expected YYYY-MM-DD")
        return
    }

    var sourceID *uuidv7.UUID
    if req.SourceID != nil && *req.SourceID != "" {
        parsed, err := uuidv7.Parse(*req.SourceID)
        if err != nil {
            response.BadRequest(c, "invalid source_id")
            return
        }
        sourceID = &parsed
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))
    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    je, err := h.uc.CreateJournalEntry(
        c.Request.Context(),
        organizationID,
        entryDate,
        req.Description,
        sourceType,
        sourceID,
        userUUID,
    )
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Created(c, dto.ToJournalEntryResponse(je))
}

// GetJournalEntry gets a journal entry by ID
// @Summary Get a journal entry
// @Tags journal-entries
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Success 200 {object} dto.JournalEntryResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id} [get]
func (h *JournalEntryHandler) GetJournalEntry(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
        return
    }

    je, err := h.uc.GetJournalEntryByID(c.Request.Context(), id)
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Success(c, dto.ToJournalEntryResponse(je))
}

// UpdateDescription updates the journal entry description
// @Summary Update journal entry description
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Param request body dto.UpdateDescriptionRequest true "Update description request"
// @Success 200 {object} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id}/description [put]
func (h *JournalEntryHandler) UpdateDescription(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
        return
    }

    var req dto.UpdateDescriptionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    je, err := h.uc.UpdateDescription(c.Request.Context(), id, req.Description, userUUID)
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Success(c, dto.ToJournalEntryResponse(je))
}

// AddLine adds a line to the journal entry
// @Summary Add line to journal entry
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Param request body dto.AddLineRequest true "Add line request"
// @Success 200 {object} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id}/lines [post]
func (h *JournalEntryHandler) AddLine(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
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

    je, err := h.uc.AddLine(
        c.Request.Context(),
        id,
        accountID,
        req.DebitCents,
        req.CreditCents,
        "UAH",
        req.Description,
        userUUID,
    )
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Success(c, dto.ToJournalEntryResponse(je))
}

// RemoveLine removes a line from the journal entry
// @Summary Remove line from journal entry
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Param lineId path string true "Line ID"
// @Success 200 {object} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id}/lines/{lineId} [delete]
func (h *JournalEntryHandler) RemoveLine(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
        return
    }

    lineID, err := uuidv7.Parse(c.Param("lineId"))
    if err != nil {
        response.BadRequest(c, "invalid line ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    je, err := h.uc.RemoveLine(c.Request.Context(), id, lineID, userUUID)
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Success(c, dto.ToJournalEntryResponse(je))
}

// PostEntry posts the journal entry to the ledger
// @Summary Post journal entry
// @Tags journal-entries
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Success 200 {object} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id}/post [post]
func (h *JournalEntryHandler) PostEntry(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    je, err := h.uc.PostEntry(c.Request.Context(), id, userUUID)
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Success(c, dto.ToJournalEntryResponse(je))
}

// ReverseEntry reverses a posted journal entry
// @Summary Reverse journal entry
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Param request body dto.ReverseEntryRequest true "Reverse entry request"
// @Success 200 {object} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id}/reverse [post]
func (h *JournalEntryHandler) ReverseEntry(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
        return
    }

    var req dto.ReverseEntryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    userID, _ := c.Get("user_id")
    userUUID, _ := uuidv7.Parse(userID.(string))

    je, err := h.uc.ReverseEntry(c.Request.Context(), id, userUUID, req.ReverseDescription)
    if err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.Success(c, dto.ToJournalEntryResponse(je))
}

// DeleteJournalEntry deletes a journal entry
// @Summary Delete a journal entry
// @Tags journal-entries
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/{id} [delete]
func (h *JournalEntryHandler) DeleteJournalEntry(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "invalid journal entry ID")
        return
    }

    if err := h.uc.DeleteJournalEntry(c.Request.Context(), id); err != nil {
        handleJournalEntryError(c, err)
        return
    }

    response.SuccessWithMessage(c, "journal entry deleted successfully")
}

// ListJournalEntries lists journal entries for an organization
// @Summary List journal entries
// @Tags journal-entries
// @Produce json
// @Param status query string false "Filter by status (draft, posted, reversed)"
// @Success 200 {array} dto.JournalEntryResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries [get]
func (h *JournalEntryHandler) ListJournalEntries(c *gin.Context) {
    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    statusStr := c.Query("status")
    var status aggregate.EntryStatus
    if statusStr != "" {
        status = aggregate.EntryStatus(statusStr)
        if !isValidEntryStatus(status) {
            response.BadRequest(c, "invalid status")
            return
        }
    }

    entries, err := h.uc.ListJournalEntriesByOrganization(c.Request.Context(), organizationID, status)
    if err != nil {
        response.InternalError(c, "failed to list journal entries")
        return
    }

    response.Success(c, dto.ToJournalEntryResponseList(entries))
}

// ListJournalEntriesByPeriod lists journal entries for a fiscal period
// @Summary List journal entries by period
// @Tags journal-entries
// @Produce json
// @Param periodId path string true "Fiscal Period ID"
// @Success 200 {array} dto.JournalEntryResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/journal-entries/period/{periodId} [get]
func (h *JournalEntryHandler) ListJournalEntriesByPeriod(c *gin.Context) {
    periodID, err := uuidv7.Parse(c.Param("periodId"))
    if err != nil {
        response.BadRequest(c, "invalid period ID")
        return
    }

    orgID, _ := c.Get("organization_id")
    organizationID, _ := uuidv7.Parse(orgID.(string))

    entries, err := h.uc.ListJournalEntriesByPeriod(c.Request.Context(), organizationID, periodID)
    if err != nil {
        response.InternalError(c, "failed to list journal entries by period")
        return
    }

    response.Success(c, dto.ToJournalEntryResponseList(entries))
}

func handleJournalEntryError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, journalentry.ErrJournalEntryNotFound):
        response.NotFound(c, "journal entry not found")
    case errors.Is(err, journalentry.ErrJournalEntryDescriptionEmpty),
        errors.Is(err, journalentry.ErrJournalEntryInvalidDate),
        errors.Is(err, journalentry.ErrJournalEntryNoLines),
        errors.Is(err, journalentry.ErrJournalEntryInvalidStatus),
        errors.Is(err, journalentry.ErrJournalEntryLineInvalidAccount),
        errors.Is(err, journalentry.ErrJournalEntryLineInvalidAmount),
        errors.Is(err, journalentry.ErrJournalEntryLineBothSides),
        errors.Is(err, journalentry.ErrJournalEntryUnbalanced),
        errors.Is(err, journalentry.ErrJournalEntryAlreadyPosted),
        errors.Is(err, journalentry.ErrJournalEntryNotPosted),
        errors.Is(err, journalentry.ErrJournalEntryAlreadyReversed),
        errors.Is(err, journalentry.ErrJournalEntryCannotPostDraft):
        response.BadRequest(c, err.Error())
    default:
        response.InternalError(c, "operation failed")
    }
}

func isValidSourceType(t aggregate.SourceType) bool {
    switch t {
    case aggregate.SourceTypeManual,
        aggregate.SourceTypeBankTransaction,
        aggregate.SourceTypeInvoice,
        aggregate.SourceTypePayment,
        aggregate.SourceTypeReceipt:
        return true
    default:
        return false
    }
}

func isValidEntryStatus(s aggregate.EntryStatus) bool {
    switch s {
    case aggregate.EntryStatusDraft,
        aggregate.EntryStatusPosted,
        aggregate.EntryStatusReversed:
        return true
    default:
        return false
    }
}