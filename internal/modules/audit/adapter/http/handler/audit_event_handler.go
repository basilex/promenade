package handler

import (
	"net/http"
	"time"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/internal/modules/audit/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

// AuditEventHandler handles audit event HTTP requests
type AuditEventHandler struct {
	useCase usecase.AuditEventUseCase
}

// NewAuditEventHandler creates a new audit event handler
func NewAuditEventHandler(useCase usecase.AuditEventUseCase) *AuditEventHandler {
	return &AuditEventHandler{
		useCase: useCase,
	}
}

// CreateAuditEvent godoc
// @Summary Create audit event
// @Description Create a new immutable audit event
// @Tags audit
// @Accept json
// @Produce json
// @Param request body CreateAuditEventRequest true "Audit event data"
// @Success 201 {object} response.SuccessResponse{data=AuditEventResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/audit/events [post]
// @Security BearerAuth
func (h *AuditEventHandler) CreateAuditEvent(c *gin.Context) {
	var req CreateAuditEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	data := &entity.AuditEventCreate{
		UserID:     req.UserID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		OldData:    req.OldData,
		NewData:    req.NewData,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		RequestID:  req.RequestID,
		Metadata:   req.Metadata,
	}

	event, err := h.useCase.CreateAuditEvent(c.Request.Context(), data)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err)
		return
	}

	resp := h.toResponse(event)
	response.Success(c, http.StatusCreated, resp)
}

// GetAuditEvent godoc
// @Summary Get audit event
// @Description Get audit event by ID
// @Tags audit
// @Produce json
// @Param id path string true "Audit Event ID (UUID)"
// @Success 200 {object} response.SuccessResponse{data=AuditEventResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/audit/events/{id} [get]
// @Security BearerAuth
func (h *AuditEventHandler) GetAuditEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", err)
		return
	}

	event, err := h.useCase.GetAuditEvent(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", err)
		return
	}

	resp := h.toResponse(event)
	response.Success(c, http.StatusOK, resp)
}

// ListAuditEvents godoc
// @Summary List audit events
// @Description List audit events with filters and pagination
// @Tags audit
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param entity_type query string false "Filter by entity type"
// @Param entity_id query string false "Filter by entity ID"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.SuccessResponse{data=[]AuditEventResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/audit/events [get]
// @Security BearerAuth
func (h *AuditEventHandler) ListAuditEvents(c *gin.Context) {
	var req ListAuditEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	filters := repository.AuditEventFilters{
		UserID:     req.UserID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	}

	params := pagination.Params{
		Page:     req.Page,
		PageSize: req.PageSize,
		Limit:    req.PageSize,
	}

	events, metadata, err := h.useCase.ListAuditEvents(c.Request.Context(), filters, params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", err)
		return
	}

	respList := make([]AuditEventResponse, len(events))
	for i, event := range events {
		respList[i] = *h.toResponse(event)
	}

	result := gin.H{
		"events":   respList,
		"metadata": metadata,
	}
	response.Success(c, http.StatusOK, result)
}

// GetAuditEventsByEntity godoc
// @Summary Get audit events by entity
// @Description Get all audit events for a specific entity
// @Tags audit
// @Produce json
// @Param entity_type path string true "Entity type"
// @Param entity_id path string true "Entity ID (UUID)"
// @Success 200 {object} response.SuccessResponse{data=[]AuditEventResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/audit/events/entity/{entity_type}/{entity_id} [get]
// @Security BearerAuth
func (h *AuditEventHandler) GetAuditEventsByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")

	entityID, err := uuidv7.Parse(entityIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", err)
		return
	}

	events, err := h.useCase.GetAuditEventsByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", err)
		return
	}

	respList := make([]AuditEventResponse, len(events))
	for i, event := range events {
		respList[i] = *h.toResponse(event)
	}

	response.Success(c, http.StatusOK, respList)
}

// VerifyAuditEvent godoc
// @Summary Verify audit event signature
// @Description Verify HMAC-SHA256 signature of audit event
// @Tags audit
// @Produce json
// @Param id path string true "Audit Event ID (UUID)"
// @Success 200 {object} response.SuccessResponse{data=VerifyAuditEventResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/audit/events/{id}/verify [get]
// @Security BearerAuth
func (h *AuditEventHandler) VerifyAuditEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", err)
		return
	}

	valid, err := h.useCase.VerifyAuditEvent(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", err)
		return
	}

	resp := VerifyAuditEventResponse{
		EventID:  id,
		Valid:    valid,
		Verified: time.Now(),
	}

	response.Success(c, http.StatusOK, resp)
}

// toResponse converts entity to response DTO
func (h *AuditEventHandler) toResponse(event *entity.AuditEvent) *AuditEventResponse {
	return &AuditEventResponse{
		ID:         event.ID,
		UserID:     event.UserID,
		Action:     event.Action,
		EntityType: event.EntityType,
		EntityID:   event.EntityID,
		OldData:    event.OldData,
		NewData:    event.NewData,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		RequestID:  event.RequestID,
		Metadata:   event.Metadata,
		Signature:  event.Signature,
		CreatedAt:  event.CreatedAt,
	}
}
