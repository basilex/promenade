package handler

import (
	"fmt"
	"net/http"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/infrastructure/scheduler"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AdminPurgeHandler handles admin purge operations
type AdminPurgeHandler struct {
	purgeUseCase usecase.PurgeUseCase
	scheduler    *scheduler.Scheduler
}

// NewAdminPurgeHandler creates a new admin purge handler
func NewAdminPurgeHandler(purgeUseCase usecase.PurgeUseCase, scheduler *scheduler.Scheduler) *AdminPurgeHandler {
	return &AdminPurgeHandler{
		purgeUseCase: purgeUseCase,
		scheduler:    scheduler,
	}
}

// TriggerPurge manually triggers a purge operation
// @Summary Manually trigger purge operation
// @Description Triggers a purge operation for soft-deleted records. Can be run in dry-run mode to preview results.
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body dto.PurgeRequest true "Purge request"
// @Success 200 {object} response.Response{data=dto.PurgeResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/purge/trigger [post]
func (h *AdminPurgeHandler) TriggerPurge(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	log.Info("Manual purge triggered",
		"entity_name", req.EntityName,
		"dry_run", req.DryRun,
	)

	if req.EntityName == "" {
		// Purge all entities
		summary, err := h.purgeUseCase.PurgeAll(ctx, req.DryRun)
		if err != nil {
			log.Error("Failed to purge all entities", "error", err)
			response.Error(c, http.StatusInternalServerError, "failed to purge records", err)
			return
		}

		response.Success(c, http.StatusOK, dto.ToPurgeResponse(summary))
		return
	}

	// Purge specific entity
	result, err := h.purgeUseCase.PurgeEntity(ctx, req.EntityName, req.DryRun)
	if err != nil {
		log.Error("Failed to purge entity",
			"entity_name", req.EntityName,
			"error", err,
		)
		response.Error(c, http.StatusInternalServerError, "failed to purge records", err)
		return
	}

	// Convert single result to summary format
	summary := &dto.PurgeResponse{
		Success:            true,
		Message:            fmt.Sprintf("Successfully purged %s", req.EntityName),
		TotalRecordsPurged: result.RecordsPurged,
		Results: []dto.PurgeResultResponse{
			{
				EntityName:    result.EntityName,
				RecordsPurged: result.RecordsPurged,
				Duration:      result.Duration.String(),
				Error:         result.Error,
			},
		},
		Duration:  result.Duration.String(),
		DryRun:    result.DryRun,
		Timestamp: result.Timestamp,
	}

	response.Success(c, http.StatusOK, summary)
}

// GetRetentionPolicies returns all configured retention policies
// @Summary Get retention policies
// @Description Returns all configured retention policies for soft-deleted records
// @Tags Admin
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.RetentionPolicyResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/purge/policies [get]
func (h *AdminPurgeHandler) GetRetentionPolicies(c *gin.Context) {
	ctx := c.Request.Context()

	policies, err := h.purgeUseCase.GetRetentionPolicies(ctx)
	if err != nil {
		logger.FromContext(ctx).Error("Failed to get retention policies", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to get retention policies", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRetentionPolicyListResponse(policies))
}

// PreviewPurge returns a count of records that would be purged
// @Summary Preview purge operation
// @Description Returns the count of records that would be purged for a specific entity without actually deleting them
// @Tags Admin
// @Produce json
// @Param entity_name path string true "Entity name" Enums(user_posts, post_comments)
// @Success 200 {object} response.Response{data=dto.PurgePreviewResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/purge/preview/{entity_name} [get]
func (h *AdminPurgeHandler) PreviewPurge(c *gin.Context) {
	ctx := c.Request.Context()
	entityName := c.Param("entity_name")

	if entityName != "user_posts" && entityName != "post_comments" {
		response.Error(c, http.StatusBadRequest, "invalid entity name", nil)
		return
	}

	policy, err := h.purgeUseCase.GetRetentionPolicy(ctx, entityName)
	if err != nil {
		logger.FromContext(ctx).Error("Failed to get retention policy",
			"entity_name", entityName,
			"error", err,
		)
		response.Error(c, http.StatusInternalServerError, "failed to get retention policy", err)
		return
	}

	count, err := h.purgeUseCase.PreviewPurge(ctx, entityName)
	if err != nil {
		logger.FromContext(ctx).Error("Failed to preview purge",
			"entity_name", entityName,
			"error", err,
		)
		response.Error(c, http.StatusInternalServerError, "failed to preview purge", err)
		return
	}

	previewResponse := &dto.PurgePreviewResponse{
		EntityName: entityName,
		Count:      count,
		CutoffDate: policy.GetCutoffDate().Format("2006-01-02 15:04:05"),
	}

	response.Success(c, http.StatusOK, previewResponse)
}

// GetSchedulerStatus returns the current status of the purge scheduler
// @Summary Get scheduler status
// @Description Returns the current status of the automatic purge scheduler
// @Tags Admin
// @Produce json
// @Success 200 {object} response.Response{data=dto.SchedulerStatusResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/purge/scheduler/status [get]
func (h *AdminPurgeHandler) GetSchedulerStatus(c *gin.Context) {
	status := h.scheduler.GetStatus()

	response.Success(c, http.StatusOK, dto.SchedulerStatusResponse{
		Enabled:       status.Enabled,
		Schedule:      status.Schedule,
		DryRun:        status.DryRun,
		IsRunning:     status.IsRunning,
		LastRunTime:   status.LastRunTime,
		LastRunStatus: status.LastRunStatus,
	})
}
