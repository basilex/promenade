package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/modules/workflows/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/workflows/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowInstanceHandler handles workflow instance HTTP requests
type WorkflowInstanceHandler struct {
	instanceUC usecase.IWorkflowInstanceUseCase
}

// NewWorkflowInstanceHandler creates a new workflow instance handler
func NewWorkflowInstanceHandler(instanceUC usecase.IWorkflowInstanceUseCase) *WorkflowInstanceHandler {
	return &WorkflowInstanceHandler{
		instanceUC: instanceUC,
	}
}

// StartWorkflow godoc
// @Summary Start a new workflow instance
// @Description Create and start a new workflow instance from a definition
// @Tags workflows
// @Accept json
// @Produce json
// @Param instance body dto.StartInstanceRequest true "Workflow instance data"
// @Success 201 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances [post]
func (h *WorkflowInstanceHandler) StartWorkflow(c *gin.Context) {
	startedBy, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	var req dto.StartInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Parse definition ID
	definitionID, err := uuidv7.Parse(req.DefinitionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow definition ID", err)
		return
	}

	// Parse optional assignedTo
	var assignedTo *uuidv7.UUID
	if req.AssignedTo != nil {
		parsed, err := uuidv7.Parse(*req.AssignedTo)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid assignedTo user ID", err)
			return
		}
		assignedTo = &parsed
	}

	// Start workflow instance
	instance, err := h.instanceUC.StartWorkflow(
		c.Request.Context(),
		definitionID,
		req.ExternalReference,
		req.Priority,
		req.Context,
		assignedTo,
		startedBy,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNotFound) {
			response.Error(c, http.StatusNotFound, "workflow definition not found", err)
			return
		}
		if errors.Is(err, usecase.ErrWorkflowNotActive) {
			response.Error(c, http.StatusBadRequest, "workflow definition is not active", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to start workflow instance", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToWorkflowInstanceResponse(instance))
}

// GetInstance godoc
// @Summary Get workflow instance by ID
// @Description Retrieve a workflow instance by its ID
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow instance ID"
// @Success 200 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances/{id} [get]
func (h *WorkflowInstanceHandler) GetInstance(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow instance ID", err)
		return
	}

	instance, err := h.instanceUC.GetInstance(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrInstanceNotFound) {
			response.Error(c, http.StatusNotFound, "workflow instance not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get workflow instance", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowInstanceResponse(instance))
}

// ListInstances godoc
// @Summary List workflow instances
// @Description List workflow instances with optional filters
// @Tags workflows
// @Accept json
// @Produce json
// @Param definition_id query string false "Filter by workflow definition ID"
// @Param status query string false "Filter by status"
// @Param assigned_to query string false "Filter by assigned user ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]dto.WorkflowInstanceResponse}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances [get]
func (h *WorkflowInstanceHandler) ListInstances(c *gin.Context) {
	// Parse filters
	var definitionID *uuidv7.UUID
	if defID := c.Query("definition_id"); defID != "" {
		parsed, err := uuidv7.Parse(defID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid definition_id", err)
			return
		}
		definitionID = &parsed
	}

	status := c.Query("status")

	var assignedTo *uuidv7.UUID
	if assignedToStr := c.Query("assigned_to"); assignedToStr != "" {
		parsed, err := uuidv7.Parse(assignedToStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid assigned_to", err)
			return
		}
		assignedTo = &parsed
	}

	// Get pagination params
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)
	offset := (page - 1) * pageSize

	// List instances
	instances, err := h.instanceUC.ListInstances(
		c.Request.Context(),
		definitionID,
		status,
		assignedTo,
		pageSize,
		offset,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list workflow instances", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowInstanceListResponse(instances))
}

// PauseInstance godoc
// @Summary Pause a workflow instance
// @Description Pause a running workflow instance
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow instance ID"
// @Success 200 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances/{id}/pause [post]
func (h *WorkflowInstanceHandler) PauseInstance(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow instance ID", err)
		return
	}

	instance, err := h.instanceUC.PauseInstance(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrInstanceNotFound) {
			response.Error(c, http.StatusNotFound, "workflow instance not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to pause workflow instance", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowInstanceResponse(instance))
}

// ResumeInstance godoc
// @Summary Resume a paused workflow instance
// @Description Resume a paused workflow instance
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow instance ID"
// @Success 200 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances/{id}/resume [post]
func (h *WorkflowInstanceHandler) ResumeInstance(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow instance ID", err)
		return
	}

	instance, err := h.instanceUC.ResumeInstance(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrInstanceNotFound) {
			response.Error(c, http.StatusNotFound, "workflow instance not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to resume workflow instance", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowInstanceResponse(instance))
}

// CancelInstance godoc
// @Summary Cancel a workflow instance
// @Description Cancel a workflow instance with optional reason
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow instance ID"
// @Param reason query string false "Cancellation reason"
// @Success 200 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances/{id}/cancel [post]
func (h *WorkflowInstanceHandler) CancelInstance(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow instance ID", err)
		return
	}

	reason := c.Query("reason")

	instance, err := h.instanceUC.CancelInstance(c.Request.Context(), id, reason)
	if err != nil {
		if errors.Is(err, usecase.ErrInstanceNotFound) {
			response.Error(c, http.StatusNotFound, "workflow instance not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to cancel workflow instance", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowInstanceResponse(instance))
}

// UpdateInstance godoc
// @Summary Update a workflow instance
// @Description Update priority or assignee of a workflow instance
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow instance ID"
// @Param update body dto.UpdateInstanceRequest true "Update data"
// @Success 200 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/instances/{id} [put]
func (h *WorkflowInstanceHandler) UpdateInstance(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow instance ID", err)
		return
	}

	var req dto.UpdateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Parse AssignedTo UUID if provided
	var assignedTo *uuidv7.UUID
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		parsedID, err := uuidv7.Parse(*req.AssignedTo)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid assigned_to UUID", err)
			return
		}
		assignedTo = &parsedID
	}

	// Update the instance
	updated, err := h.instanceUC.UpdateInstance(c.Request.Context(), id, req.Priority, assignedTo)
	if err != nil {
		if errors.Is(err, usecase.ErrInstanceNotFound) {
			response.Error(c, http.StatusNotFound, "workflow instance not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update workflow instance", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowInstanceResponse(updated))
}
