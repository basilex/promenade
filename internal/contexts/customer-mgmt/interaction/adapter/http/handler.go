package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// InteractionHandler handles HTTP requests for interaction operations
type InteractionHandler struct {
	usecase interaction.IUseCase
}

// NewInteractionHandler creates a new interaction handler
func NewInteractionHandler(usecase interaction.IUseCase) *InteractionHandler {
	return &InteractionHandler{
		usecase: usecase,
	}
}

// Create godoc
// @Summary Create a new interaction
// @Description Create a new customer/company interaction
// @Tags interactions
// @Accept json
// @Produce json
// @Param request body CreateInteractionRequest true "Interaction data"
// @Success 201 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions [post]
func (h *InteractionHandler) Create(c *gin.Context) {
	var req CreateInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse UUIDs
	customerID, err := parseUUID(req.CustomerID)
	if err != nil {
		response.BadRequest(c, "invalid customer_id")
		return
	}

	var companyID *uuidv7.UUID
	if req.CompanyID != nil && *req.CompanyID != "" {
		id, err := parseUUID(*req.CompanyID)
		if err != nil {
			response.BadRequest(c, "invalid company_id")
			return
		}
		companyID = &id
	}

	createdBy, err := parseUUID(req.CreatedBy)
	if err != nil {
		response.BadRequest(c, "invalid created_by")
		return
	}

	startedAt, err := parseTime(req.StartedAt)
	if err != nil {
		response.BadRequest(c, "invalid started_at format")
		return
	}

	// Create interaction
	inter, err := h.usecase.CreateInteraction(
		c.Request.Context(),
		customerID,
		companyID,
		req.Type,
		req.Direction,
		req.Subject,
		req.Description,
		createdBy,
		startedAt,
	)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create interaction")
		return
	}

	response.Created(c, ToInteractionResponse(inter))
}

// List godoc
// @Summary List interactions
// @Description List all interactions with pagination
// @Tags interactions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]InteractionResponse}
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions [get]
func (h *InteractionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	interactions, total, err := h.usecase.ListByCustomer(c.Request.Context(), uuidv7.Nil, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list interactions")
		return
	}

	// Convert to response DTOs
	respData := make([]*InteractionResponse, len(interactions))
	for i, inter := range interactions {
		respData[i] = ToInteractionResponse(inter)
	}

	response.SuccessWithPagination(c, respData, total, page, pageSize)
}

// GetByID godoc
// @Summary Get interaction by ID
// @Description Get a specific interaction by its ID
// @Tags interactions
// @Produce json
// @Param id path string true "Interaction ID"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id} [get]
func (h *InteractionHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	inter, err := h.usecase.GetInteraction(c.Request.Context(), id)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve interaction")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// ListByCustomer godoc
// @Summary List interactions by customer
// @Description List all interactions for a specific customer
// @Tags interactions
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/customer/{customer_id} [get]
func (h *InteractionHandler) ListByCustomer(c *gin.Context) {
	customerID, err := parseUUID(c.Param("customer_id"))
	if err != nil {
		response.BadRequest(c, "invalid customer_id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	interactions, total, err := h.usecase.ListByCustomer(c.Request.Context(), customerID, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list customer interactions")
		return
	}

	// Convert to response DTOs
	respData := make([]*InteractionResponse, len(interactions))
	for i, inter := range interactions {
		respData[i] = ToInteractionResponse(inter)
	}

	response.SuccessWithPagination(c, respData, total, page, pageSize)
}

// ListByCompany godoc
// @Summary List interactions by company
// @Description List all interactions for a specific company
// @Tags interactions
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/company/{company_id} [get]
func (h *InteractionHandler) ListByCompany(c *gin.Context) {
	companyID, err := parseUUID(c.Param("company_id"))
	if err != nil {
		response.BadRequest(c, "invalid company_id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	interactions, total, err := h.usecase.ListByCompany(c.Request.Context(), companyID, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list company interactions")
		return
	}

	// Convert to response DTOs
	respData := make([]*InteractionResponse, len(interactions))
	for i, inter := range interactions {
		respData[i] = ToInteractionResponse(inter)
	}

	response.SuccessWithPagination(c, respData, total, page, pageSize)
}

// ListByType godoc
// @Summary List interactions by type
// @Description List all interactions of a specific type
// @Tags interactions
// @Produce json
// @Param type path string true "Interaction Type (call, email, meeting, note, sms, chat)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/type/{type} [get]
func (h *InteractionHandler) ListByType(c *gin.Context) {
	interactionType := c.Param("type")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	interactions, total, err := h.usecase.ListByType(c.Request.Context(), interactionType, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list interactions by type")
		return
	}

	// Convert to response DTOs
	respData := make([]*InteractionResponse, len(interactions))
	for i, inter := range interactions {
		respData[i] = ToInteractionResponse(inter)
	}

	response.SuccessWithPagination(c, respData, total, page, pageSize)
}

// ListPendingFollowUps godoc
// @Summary List pending follow-ups
// @Description List all interactions with pending follow-ups
// @Tags interactions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]InteractionResponse}
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/pending-followups [get]
func (h *InteractionHandler) ListPendingFollowUps(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	interactions, total, err := h.usecase.ListPendingFollowUps(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list pending follow-ups")
		return
	}

	// Convert to response DTOs
	respData := make([]*InteractionResponse, len(interactions))
	for i, inter := range interactions {
		respData[i] = ToInteractionResponse(inter)
	}

	response.SuccessWithPagination(c, respData, total, page, pageSize)
}

// UpdateContent godoc
// @Summary Update interaction content
// @Description Update subject and description of an interaction
// @Tags interactions
// @Accept json
// @Produce json
// @Param id path string true "Interaction ID"
// @Param request body UpdateContentRequest true "Content data"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id}/content [put]
func (h *InteractionHandler) UpdateContent(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	var req UpdateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inter, err := h.usecase.UpdateContent(c.Request.Context(), id, req.Subject, req.Description)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update interaction content")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// SetOutcome godoc
// @Summary Set interaction outcome
// @Description Set the outcome of an interaction
// @Tags interactions
// @Accept json
// @Produce json
// @Param id path string true "Interaction ID"
// @Param request body SetOutcomeRequest true "Outcome data"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id}/outcome [put]
func (h *InteractionHandler) SetOutcome(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	var req SetOutcomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inter, err := h.usecase.SetOutcome(c.Request.Context(), id, req.Outcome)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set interaction outcome")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// EndInteraction godoc
// @Summary End an interaction
// @Description Mark an interaction as ended with end time
// @Tags interactions
// @Accept json
// @Produce json
// @Param id path string true "Interaction ID"
// @Param request body EndInteractionRequest true "End time data"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id}/end [put]
func (h *InteractionHandler) EndInteraction(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	var req EndInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	endedAt, err := parseTime(req.EndedAt)
	if err != nil {
		response.BadRequest(c, "invalid ended_at format")
		return
	}

	inter, err := h.usecase.EndInteraction(c.Request.Context(), id, endedAt)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		if err == interaction.ErrInteractionAlreadyEnded {
			response.BadRequest(c, "Interaction already ended")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to end interaction")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// SetFollowUp godoc
// @Summary Set follow-up requirements
// @Description Configure follow-up requirements for an interaction
// @Tags interactions
// @Accept json
// @Produce json
// @Param id path string true "Interaction ID"
// @Param request body SetFollowUpRequest true "Follow-up data"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id}/followup [put]
func (h *InteractionHandler) SetFollowUp(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	var req SetFollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var followUpDate *time.Time
	if req.FollowUpDate != nil && *req.FollowUpDate != "" {
		date, err := parseOptionalTime(req.FollowUpDate)
		if err != nil {
			response.BadRequest(c, "invalid follow_up_date format")
			return
		}
		followUpDate = date
	}

	inter, err := h.usecase.SetFollowUp(c.Request.Context(), id, req.Required, followUpDate, req.Notes)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set follow-up")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// AddAttendee godoc
// @Summary Add an attendee
// @Description Add a participant to an interaction
// @Tags interactions
// @Accept json
// @Produce json
// @Param id path string true "Interaction ID"
// @Param request body AddAttendeeRequest true "Attendee data"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id}/attendees [post]
func (h *InteractionHandler) AddAttendee(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	var req AddAttendeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	attendeeID, err := parseUUID(req.AttendeeID)
	if err != nil {
		response.BadRequest(c, "invalid attendee_id")
		return
	}

	inter, err := h.usecase.AddAttendee(c.Request.Context(), id, attendeeID)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add attendee")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// RemoveAttendee godoc
// @Summary Remove an attendee
// @Description Remove a participant from an interaction
// @Tags interactions
// @Accept json
// @Produce json
// @Param id path string true "Interaction ID"
// @Param attendee_id path string true "Attendee ID"
// @Success 200 {object} response.Response{data=InteractionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id}/attendees/{attendee_id} [delete]
func (h *InteractionHandler) RemoveAttendee(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	attendeeID, err := parseUUID(c.Param("attendee_id"))
	if err != nil {
		response.BadRequest(c, "invalid attendee_id")
		return
	}

	inter, err := h.usecase.RemoveAttendee(c.Request.Context(), id, attendeeID)
	if err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove attendee")
		return
	}

	response.Success(c, ToInteractionResponse(inter))
}

// Delete godoc
// @Summary Delete an interaction
// @Description Soft delete an interaction
// @Tags interactions
// @Produce json
// @Param id path string true "Interaction ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/interactions/{id} [delete]
func (h *InteractionHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid interaction ID")
		return
	}

	if err := h.usecase.DeleteInteraction(c.Request.Context(), id); err != nil {
		if err == interaction.ErrInteractionNotFound {
			response.NotFound(c, "Interaction not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete interaction")
		return
	}

	response.SuccessWithMessage(c, "Interaction deleted successfully")
}
