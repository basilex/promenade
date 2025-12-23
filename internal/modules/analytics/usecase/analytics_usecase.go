package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/internal/modules/analytics/domain/repository"
)

type AnalyticsUseCase struct {
	metricRepo repository.MetricRepository
	logger     *slog.Logger
}

func NewAnalyticsUseCase(metricRepo repository.MetricRepository, logger *slog.Logger) *AnalyticsUseCase {
	return &AnalyticsUseCase{
		metricRepo: metricRepo,
		logger:     logger,
	}
}

type CollectMetricInput struct {
	Module   string
	Scope    entity.MetricScope
	ScopeID  string
	Name     string
	Type     entity.MetricType
	Value    float64
	Metadata map[string]interface{}
}

func (uc *AnalyticsUseCase) CollectMetric(ctx context.Context, input *CollectMetricInput) (*entity.Metric, error) {
	if input.Module == "" {
		return nil, entity.ErrInvalidModule
	}
	if input.Name == "" {
		return nil, entity.ErrInvalidMetricName
	}

	metric := &entity.Metric{
		Module:    input.Module,
		Scope:     input.Scope,
		ScopeID:   input.ScopeID,
		Name:      input.Name,
		Type:      input.Type,
		Value:     input.Value,
		Metadata:  input.Metadata,
		CreatedAt: time.Now(),
	}

	if err := metric.Validate(); err != nil {
		return nil, err
	}

	if err := uc.metricRepo.Store(ctx, metric); err != nil {
		uc.logger.Error("Failed to store metric", "error", err)
		return nil, fmt.Errorf("failed to collect metric: %w", err)
	}

	uc.logger.Info("Metric collected", "id", metric.ID, "module", metric.Module, "name", metric.Name)
	return metric, nil
}

func (uc *AnalyticsUseCase) GetMetric(ctx context.Context, id string) (*entity.Metric, error) {
	if id == "" {
		return nil, fmt.Errorf("metric ID is required")
	}
	return uc.metricRepo.GetByID(ctx, id)
}

func (uc *AnalyticsUseCase) ListMetricsByScope(ctx context.Context, scope entity.MetricScope, scopeID string, limit, offset int) ([]*entity.Metric, int64, error) {
	if scopeID == "" {
		return nil, 0, fmt.Errorf("scope ID is required")
	}
	return uc.metricRepo.ListByScope(ctx, string(scope), scopeID, limit, offset)
}

func (uc *AnalyticsUseCase) ListMetricsByModule(ctx context.Context, module string, limit, offset int) ([]*entity.Metric, int64, error) {
	if module == "" {
		return nil, 0, fmt.Errorf("module name is required")
	}
	return uc.metricRepo.ListByModule(ctx, module, limit, offset)
}

func (uc *AnalyticsUseCase) GetLatestMetric(ctx context.Context, module string, scope entity.MetricScope, scopeID, name string) (*entity.Metric, error) {
	if module == "" || scopeID == "" || name == "" {
		return nil, fmt.Errorf("module, scope ID, and metric name are required")
	}
	return uc.metricRepo.GetLatestByName(ctx, module, string(scope), scopeID, name)
}
