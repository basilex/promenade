package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/analytics/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/internal/modules/analytics/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/response"
)

type MetricsHandler struct {
	analyticsUC *usecase.AnalyticsUseCase
	logger      *slog.Logger
}

func NewMetricsHandler(analyticsUC *usecase.AnalyticsUseCase, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		analyticsUC: analyticsUC,
		logger:      logger,
	}
}

func (h *MetricsHandler) CollectMetric(c *gin.Context) {
	var req dto.CollectMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	input := &usecase.CollectMetricInput{
		Module:   req.Module,
		Scope:    entity.MetricScope(req.Scope),
		ScopeID:  req.ScopeID,
		Name:     req.Name,
		Type:     entity.MetricType(req.Type),
		Value:    req.Value,
		Metadata: req.Metadata,
	}

	metric, err := h.analyticsUC.CollectMetric(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to collect metric", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to collect metric", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToMetricResponse(metric))
}

func (h *MetricsHandler) GetMetric(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "metric ID is required", errors.New("metric ID is required"))
		return
	}

	metric, err := h.analyticsUC.GetMetric(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "metric not found", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToMetricResponse(metric))
}

func (h *MetricsHandler) ListMetrics(c *gin.Context) {
	module := c.Query("module")
	scope := c.Query("scope")
	scopeID := c.Query("scope_id")

	// Parse pagination params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var metrics []*entity.Metric
	var total int64
	var err error

	if scope != "" {
		if scopeID == "" {
			response.Error(c, http.StatusBadRequest, "scope_id is required when scope is provided", errors.New("scope_id is required"))
			return
		}
		metrics, total, err = h.analyticsUC.ListMetricsByScope(c.Request.Context(), entity.MetricScope(scope), scopeID, limit, offset)
	} else if module != "" {
		metrics, total, err = h.analyticsUC.ListMetricsByModule(c.Request.Context(), module, limit, offset)
	} else {
		response.Error(c, http.StatusBadRequest, "either module or scope parameter is required", errors.New("either module or scope parameter is required"))
		return
	}

	if err != nil {
		h.logger.Error("Failed to list metrics", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to list metrics", err)
		return
	}

	metadata := pagination.NewMetadata(int(total), limit, offset)
	response.Success(c, http.StatusOK, dto.ToMetricListResponse(metrics, metadata))
}

func (h *MetricsHandler) GetLatestMetric(c *gin.Context) {
	module := c.Query("module")
	scope := c.Query("scope")
	scopeID := c.Query("scope_id")
	name := c.Query("name")

	if module == "" || scope == "" || scopeID == "" || name == "" {
		response.Error(c, http.StatusBadRequest, "module, scope, scope_id, and name are required", errors.New("missing required parameters"))
		return
	}

	metric, err := h.analyticsUC.GetLatestMetric(c.Request.Context(), module, entity.MetricScope(scope), scopeID, name)
	if err != nil {
		response.Error(c, http.StatusNotFound, "metric not found", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToMetricResponse(metric))
}
