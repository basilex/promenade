package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/modules/workflows/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/internal/modules/workflows/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowDefinitionHandler handles workflow definition HTTP requests
type WorkflowDefinitionHandler struct {
	definitionUC usecase.IWorkflowDefinitionUseCase
}

// NewWorkflowDefinitionHandler creates a new workflow definition handler
func NewWorkflowDefinitionHandler(definitionUC usecase.IWorkflowDefinitionUseCase) *WorkflowDefinitionHandler {
	return &WorkflowDefinitionHandler{
		definitionUC: definitionUC,
	}
}

// CreateDefinition godoc
// @Summary Create a new workflow definition
// @Description Create a new workflow definition with schema and metadata
// @Tags workflows
// @Accept json
// @Produce json
// @Param definition body dto.CreateDefinitionRequest true "Workflow definition data"
// @Success 201 {object} response.Response{data=dto.WorkflowDefinitionResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions [post]
func (h *WorkflowDefinitionHandler) CreateDefinition(c *gin.Context) {
	createdBy, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	var req dto.CreateDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Create workflow definition
	definition, err := h.definitionUC.Create(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.Version,
		req.Schema,
		req.Tags,
		req.Metadata,
		createdBy,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNameAlreadyExists) {
			response.Error(c, http.StatusConflict, "workflow name already exists", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create workflow definition", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToWorkflowDefinitionResponse(definition))
}

// GetDefinition godoc
// @Summary Get workflow definition by ID
// @Description Retrieve a workflow definition by its ID
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow definition ID"
// @Success 200 {object} response.Response{data=dto.WorkflowDefinitionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions/{id} [get]
func (h *WorkflowDefinitionHandler) GetDefinition(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow definition ID", err)
		return
	}

	definition, err := h.definitionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNotFound) {
			response.Error(c, http.StatusNotFound, "workflow definition not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get workflow definition", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowDefinitionResponse(definition))
}

// GetDefinitionByName godoc
// @Summary Get workflow definition by name
// @Description Retrieve the latest active workflow definition by name
// @Tags workflows
// @Accept json
// @Produce json
// @Param name path string true "Workflow name"
// @Success 200 {object} response.Response{data=dto.WorkflowDefinitionResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions/name/{name} [get]
func (h *WorkflowDefinitionHandler) GetDefinitionByName(c *gin.Context) {
	name := c.Param("name")

	definition, err := h.definitionUC.GetByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNotFound) {
			response.Error(c, http.StatusNotFound, "workflow definition not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get workflow definition", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowDefinitionResponse(definition))
}

// ActivateDefinition godoc
// @Summary Activate workflow definition
// @Description Activate a workflow definition to make it available for use
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow definition ID"
// @Success 200 {object} response.Response{data=dto.WorkflowDefinitionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions/{id}/activate [post]
func (h *WorkflowDefinitionHandler) ActivateDefinition(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow definition ID", err)
		return
	}

	definition, err := h.definitionUC.Activate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNotFound) {
			response.Error(c, http.StatusNotFound, "workflow definition not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to activate workflow definition", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowDefinitionResponse(definition))
}

// ListDefinitions godoc
// @Summary List workflow definitions
// @Description List workflow definitions with pagination and optional filters
// @Tags workflows
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (draft, active, deprecated, archived)"
// @Param category query string false "Filter by category"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 20, max 100)"
// @Success 200 {object} response.PaginatedResponse{data=[]dto.WorkflowDefinitionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions [get]
func (h *WorkflowDefinitionHandler) ListDefinitions(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)
	offset := (page - 1) * pageSize

	// Get filter parameters
	statusStr := c.Query("status")
	category := c.Query("category")

	// Parse status filter
	var status *entity.WorkflowDefinitionStatus
	if statusStr != "" {
		s := entity.WorkflowDefinitionStatus(statusStr)
		status = &s
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	// List definitions
	definitions, total, err := h.definitionUC.List(c.Request.Context(), status, categoryPtr, &userID, pageSize, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list workflow definitions", err)
		return
	}

	// Convert to DTOs
	dtoDefinitions := make([]*dto.WorkflowDefinitionResponse, len(definitions))
	for i, def := range definitions {
		dtoDefinitions[i] = dto.ToWorkflowDefinitionResponse(def)
	}

	response.SuccessWithPagination(c, http.StatusOK, dtoDefinitions, int(total), page, pageSize)
}

// UpdateDefinition godoc
// @Summary Update a workflow definition
// @Description Update an existing workflow definition (only in draft status)
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow definition ID"
// @Param definition body dto.CreateDefinitionRequest true "Updated definition data"
// @Success 200 {object} response.Response{data=dto.WorkflowDefinitionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions/{id} [put]
func (h *WorkflowDefinitionHandler) UpdateDefinition(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow definition ID", err)
		return
	}

	var req dto.CreateDefinitionRequest // Reuse create request DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Get existing definition to check ownership
	existing, err := h.definitionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNotFound) {
			response.Error(c, http.StatusNotFound, "workflow definition not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get workflow definition", err)
		return
	}

	// Check ownership
	if existing.CreatedBy != userID {
		response.Error(c, http.StatusForbidden, "you can only update your own definitions", nil)
		return
	}

	// Check if definition is in draft status
	if existing.Status != "draft" {
		response.Error(c, http.StatusConflict, "can only update definitions in draft status", nil)
		return
	}

	// Update the definition
	updated, err := h.definitionUC.Update(
		c.Request.Context(),
		id,
		&req.Name,
		&req.Description,
		&req.Schema,
		req.Tags,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update workflow definition", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToWorkflowDefinitionResponse(updated))
}

// DeleteDefinition godoc
// @Summary Delete a workflow definition
// @Description Soft delete a workflow definition (only in draft status)
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow definition ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /workflows/definitions/{id} [delete]
func (h *WorkflowDefinitionHandler) DeleteDefinition(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workflow definition ID", err)
		return
	}

	// Get existing definition to check ownership and status
	existing, err := h.definitionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrWorkflowNotFound) {
			response.Error(c, http.StatusNotFound, "workflow definition not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get workflow definition", err)
		return
	}

	// Check ownership
	if existing.CreatedBy != userID {
		response.Error(c, http.StatusForbidden, "you can only delete your own definitions", nil)
		return
	}

	// Check if definition is in draft status
	if existing.Status != "draft" {
		response.Error(c, http.StatusConflict, "can only delete definitions in draft status", nil)
		return
	}

	// Delete the definition
	if err := h.definitionUC.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete workflow definition", err)
		return
	}

	c.Status(http.StatusNoContent)
}
