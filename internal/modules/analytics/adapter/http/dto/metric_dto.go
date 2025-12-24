package dto

import (
	"time"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
)

type CollectMetricRequest struct {
	IModule   string                 `json:"module" binding:"required"`
	Scope    string                 `json:"scope" binding:"required,oneof=user post comment system"`
	ScopeID  string                 `json:"scope_id" binding:"required"`
	Name     string                 `json:"name" binding:"required"`
	Type     string                 `json:"type" binding:"required,oneof=counter gauge histogram"`
	Value    float64                `json:"value" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

type MetricResponse struct {
	ID        string                 `json:"id"`
	IModule    string                 `json:"module"`
	Scope     string                 `json:"scope"`
	ScopeID   string                 `json:"scope_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Value     float64                `json:"value"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type MetricListResponse struct {
	Metrics    []*MetricResponse `json:"metrics"`
	Pagination map[string]any    `json:"pagination,omitempty"`
}

func ToMetricResponse(metric *entity.Metric) *MetricResponse {
	return &MetricResponse{
		ID:        metric.ID,
		IModule:    metric.IModule,
		Scope:     string(metric.Scope),
		ScopeID:   metric.ScopeID,
		Name:      metric.Name,
		Type:      string(metric.Type),
		Value:     metric.Value,
		Metadata:  metric.Metadata,
		CreatedAt: metric.CreatedAt,
	}
}

func ToMetricResponseList(metrics []*entity.Metric) []*MetricResponse {
	responses := make([]*MetricResponse, len(metrics))
	for i, metric := range metrics {
		responses[i] = ToMetricResponse(metric)
	}
	return responses
}

func ToMetricListResponse(metrics []*entity.Metric, meta *pagination.Metadata) *MetricListResponse {
	var paginationData map[string]any
	if meta != nil {
		paginationData = map[string]any{
			"total":        meta.Total,
			"limit":        meta.Limit,
			"offset":       meta.Offset,
			"total_pages":  meta.TotalPages,
			"current_page": meta.CurrentPage,
			"has_next":     meta.HasNext,
			"has_prev":     meta.HasPrev,
		}
	}

	return &MetricListResponse{
		Metrics:    ToMetricResponseList(metrics),
		Pagination: paginationData,
	}
}
