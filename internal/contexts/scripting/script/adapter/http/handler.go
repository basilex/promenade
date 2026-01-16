package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	//nolint:staticcheck // dot import improves handler readability by removing script. prefix from 100+ locations
	. "github.com/basilex/promenade/internal/contexts/scripting/script"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ScriptHandler handles HTTP requests for script operations
type ScriptHandler struct {
	useCase IScriptUseCase
}

// NewScriptHandler creates a new script handler
func NewScriptHandler(useCase IScriptUseCase) *ScriptHandler {
	return &ScriptHandler{
		useCase: useCase,
	}
}

// ============================================================================
// Script Execution Endpoints
// ============================================================================

// ExecuteScript executes a script by name
// @Summary Execute script
// @Description Execute LUA script by name with parameters
// @Tags scripts
// @Accept json
// @Produce json
// @Param name path string true "Script name"
// @Param body body ExecuteScriptRequest true "Execution parameters"
// @Success 200 {object} ExecutionResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/name/{name}/execute [post]
func (h *ScriptHandler) ExecuteScript(c *gin.Context) {
	scriptName := c.Param("name")
	if scriptName == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "script name is required")
		return
	}

	var req ExecuteScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get user ID from JWT context (middleware will be added later)
	executedBy := uuidv7.New() // TODO: Extract from JWT context

	result, err := h.useCase.ExecuteScript(c.Request.Context(), scriptName, req.Parameters, executedBy)
	if err != nil {
		if errors.Is(err, ErrScriptNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
			return
		}
		if errors.Is(err, ErrScriptNotActive) {
			response.ErrorResponse(c, http.StatusBadRequest, "SCRIPT_NOT_ACTIVE", "Script is not active")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "EXECUTION_ERROR", "Failed to execute script")
		return
	}

	// Get execution details from result
	if execution, ok := result.(*ScriptExecution); ok {
		response.Success(c, ToExecutionResponse(execution))
	} else {
		response.Success(c, gin.H{"result": result})
	}
}

// ValidateScript validates script syntax
// @Summary Validate script syntax
// @Description Validate LUA script syntax without executing
// @Tags scripts
// @Accept json
// @Produce json
// @Param body body ValidateScriptRequest true "Script code"
// @Success 200 {object} ValidationResponse
// @Failure 400 {object} response.Response
// @Router /scripts/validate [post]
func (h *ScriptHandler) ValidateScript(c *gin.Context) {
	var req ValidateScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err := h.useCase.ValidateScript(c.Request.Context(), req.Code)
	if err != nil {
		response.Success(c, ValidationResponse{
			Valid:   false,
			Message: err.Error(),
		})
		return
	}

	response.Success(c, ValidationResponse{
		Valid:   true,
		Message: "Script syntax is valid",
	})
}

// ============================================================================
// Script CRUD Endpoints
// ============================================================================

// CreateScript creates a new script
// @Summary Create script
// @Description Create a new LUA script
// @Tags scripts
// @Accept json
// @Produce json
// @Param body body CreateScriptRequest true "Script data"
// @Success 201 {object} ScriptResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts [post]
func (h *ScriptHandler) CreateScript(c *gin.Context) {
	var req CreateScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get user ID from JWT context
	createdBy := uuidv7.New() // TODO: Extract from JWT context

	scr, err := h.useCase.CreateScript(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.Code,
		createdBy,
	)
	if err != nil {
		// Validation errors
		if errors.Is(err, ErrScriptNameEmpty) {
			response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Script name is required")
			return
		}
		if errors.Is(err, ErrScriptCodeEmpty) {
			response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Script code is required")
			return
		}
		if errors.Is(err, ErrScriptSyntaxInvalid) {
			response.ErrorResponse(c, http.StatusBadRequest, "SYNTAX_ERROR", "Script syntax is invalid")
			return
		}
		// System errors
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create script")
		return
	}

	response.Created(c, ToScriptResponse(scr))
}

// GetScript retrieves a script by ID
// @Summary Get script by ID
// @Description Get script details by UUID
// @Tags scripts
// @Produce json
// @Param id path string true "Script ID (UUID)"
// @Success 200 {object} ScriptResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /scripts/{id} [get]
func (h *ScriptHandler) GetScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	scr, err := h.useCase.GetScript(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
		return
	}

	response.Success(c, ToScriptResponse(scr))
}

// GetScriptByName retrieves a script by name
// @Summary Get script by name
// @Description Get script details by name
// @Tags scripts
// @Produce json
// @Param name path string true "Script name"
// @Success 200 {object} ScriptResponse
// @Failure 404 {object} response.Response
// @Router /scripts/name/{name} [get]
func (h *ScriptHandler) GetScriptByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "script name is required")
		return
	}

	scr, err := h.useCase.GetScriptByName(c.Request.Context(), name)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
		return
	}

	response.Success(c, ToScriptResponse(scr))
}

// UpdateScript updates an existing script
// @Summary Update script
// @Description Update script code, description, or metadata
// @Tags scripts
// @Accept json
// @Produce json
// @Param id path string true "Script ID (UUID)"
// @Param body body UpdateScriptRequest true "Update data"
// @Success 200 {object} ScriptResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/{id} [put]
func (h *ScriptHandler) UpdateScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	var req UpdateScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	var scr *Script

	// Update code if provided
	if req.Code != nil {
		err = h.useCase.UpdateScript(c.Request.Context(), id, *req.Code)
		if err != nil {
			if errors.Is(err, ErrScriptNotFound) {
				response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
				return
			}
			if errors.Is(err, ErrScriptCodeEmpty) {
				response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Script code cannot be empty")
				return
			}
			if errors.Is(err, ErrScriptSyntaxInvalid) {
				response.ErrorResponse(c, http.StatusBadRequest, "SYNTAX_ERROR", "Script syntax is invalid")
				return
			}
			response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update script")
			return
		}
	}

	// Fetch updated script
	scr, err = h.useCase.GetScript(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
		return
	}

	// TODO: Update metadata and description if provided
	// Currently only code update is supported via UpdateScript method

	response.Success(c, ToScriptResponse(scr))
}

// DeleteScript deletes a script
// @Summary Delete script
// @Description Delete a script by ID
// @Tags scripts
// @Param id path string true "Script ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/{id} [delete]
func (h *ScriptHandler) DeleteScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	if err := h.useCase.DeleteScript(c.Request.Context(), id); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete script")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListScripts lists all scripts with filters
// @Summary List scripts
// @Description List scripts with optional status filter
// @Tags scripts
// @Produce json
// @Param status query string false "Filter by status (draft, active, inactive, archived)"
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} ScriptListResponse
// @Failure 400 {object} response.Response
// @Router /scripts [get]
func (h *ScriptHandler) ListScripts(c *gin.Context) {
	statusStr := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var status *ScriptStatus
	if statusStr != "" {
		s := ScriptStatus(statusStr)
		status = &s
	}

	var scripts []*Script
	var total int

	if status != nil {
		scripts, total, err = h.useCase.ListScripts(c.Request.Context(), *status, limit, offset)
	} else {
		scripts, total, err = h.useCase.ListAllScripts(c.Request.Context(), limit, offset)
	}

	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_ERROR", "Failed to list scripts")
		return
	}

	response.Success(c, ToScriptListResponse(scripts, total))
}

// ============================================================================
// Script Status Management Endpoints
// ============================================================================

// ActivateScript activates a script
// @Summary Activate script
// @Description Change script status to active
// @Tags scripts
// @Param id path string true "Script ID (UUID)"
// @Success 200 {object} ScriptResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/{id}/activate [post]
func (h *ScriptHandler) ActivateScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	if err := h.useCase.ActivateScript(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrScriptNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
			return
		}
		if errors.Is(err, ErrScriptCannotActivateArchived) {
			response.ErrorResponse(c, http.StatusBadRequest, "CANNOT_ACTIVATE_ARCHIVED", "Cannot activate archived script")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "ACTIVATION_ERROR", "Failed to activate script")
		return
	}

	// Fetch updated script
	scr, err := h.useCase.GetScript(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
		return
	}

	response.Success(c, ToScriptResponse(scr))
}

// DeactivateScript deactivates a script
// @Summary Deactivate script
// @Description Change script status to inactive
// @Tags scripts
// @Param id path string true "Script ID (UUID)"
// @Success 200 {object} ScriptResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/{id}/deactivate [post]
func (h *ScriptHandler) DeactivateScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	if err := h.useCase.DeactivateScript(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrScriptNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
			return
		}
		if errors.Is(err, ErrScriptCannotDeactivateArchived) {
			response.ErrorResponse(c, http.StatusBadRequest, "CANNOT_DEACTIVATE_ARCHIVED", "Cannot deactivate archived script")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DEACTIVATION_ERROR", "Failed to deactivate script")
		return
	}

	// Fetch updated script
	scr, err := h.useCase.GetScript(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
		return
	}

	response.Success(c, ToScriptResponse(scr))
}

// ArchiveScript archives a script
// @Summary Archive script
// @Description Archive a script (soft delete)
// @Tags scripts
// @Param id path string true "Script ID (UUID)"
// @Success 200 {object} ScriptResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/{id}/archive [post]
func (h *ScriptHandler) ArchiveScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	if err := h.useCase.ArchiveScript(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrScriptNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "ARCHIVE_ERROR", "Failed to archive script")
		return
	}

	// Fetch updated script
	scr, err := h.useCase.GetScript(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
		return
	}

	response.Success(c, ToScriptResponse(scr))
}

// ============================================================================
// Execution History Endpoints
// ============================================================================

// GetExecutionHistory retrieves execution history for a script
// @Summary Get execution history
// @Description Get execution history for a specific script
// @Tags scripts
// @Produce json
// @Param id path string true "Script ID (UUID)"
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} ExecutionHistoryResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /scripts/{id}/executions [get]
func (h *ScriptHandler) GetExecutionHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	executions, total, err := h.useCase.GetExecutionHistory(c.Request.Context(), id, limit, offset)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "QUERY_ERROR", "Failed to fetch execution history")
		return
	}

	response.Success(c, ToExecutionHistoryResponse(executions, total))
}

// ListScriptVersions retrieves version history for a script
// @Summary List script versions
// @Description List version history for a script
// @Tags scripts
// @Produce json
// @Param id path string true "Script ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]ScriptVersionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scripts/{id}/versions [get]
func (h *ScriptHandler) ListScriptVersions(c *gin.Context) {
	scriptID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid script ID format")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	versions, total, err := h.useCase.ListScriptVersions(c.Request.Context(), scriptID, pageSize, (page-1)*pageSize)
	if err != nil {
		if errors.Is(err, ErrScriptNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "SCRIPT_NOT_FOUND", "Script not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list script versions")
		return
	}

	response.SuccessWithPagination(c, ToScriptVersionListResponse(versions), int64(total), page, pageSize)
}

// GetRecentExecutions retrieves recent executions across all scripts
// @Summary Get recent executions
// @Description Get recent executions across all scripts
// @Tags scripts
// @Produce json
// @Param limit query int false "Limit results (default: 20)"
// @Success 200 {object} ExecutionHistoryResponse
// @Failure 400 {object} response.Response
// @Router /scripts/executions/recent [get]
func (h *ScriptHandler) GetRecentExecutions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	executions, err := h.useCase.GetRecentExecutions(c.Request.Context(), limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "QUERY_ERROR", "Failed to fetch recent executions")
		return
	}

	response.Success(c, ToExecutionHistoryResponse(executions, len(executions)))
}

// GetExecutionDetails retrieves details of a single execution
// @Summary Get execution details
// @Description Get detailed information about a specific execution
// @Tags scripts
// @Produce json
// @Param execution_id path string true "Execution ID (UUID)"
// @Success 200 {object} ExecutionResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /scripts/executions/{execution_id} [get]
func (h *ScriptHandler) GetExecutionDetails(c *gin.Context) {
	idStr := c.Param("execution_id")
	id, err := ParseUUID(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "invalid UUID format")
		return
	}

	execution, err := h.useCase.GetExecutionDetails(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "EXECUTION_NOT_FOUND", "Execution not found")
		return
	}

	response.Success(c, ToExecutionResponse(execution))
}
