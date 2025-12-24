package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
)

type IMetricRepository interface {
	Store(ctx context.Context, metric *entity.Metric) error
	GetByID(ctx context.Context, id string) (*entity.Metric, error)
	ListByScope(ctx context.Context, scope, scopeID string, limit, offset int) ([]*entity.Metric, int64, error)
	ListByModule(ctx context.Context, module string, limit, offset int) ([]*entity.Metric, int64, error)
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
	GetLatestByName(ctx context.Context, module, scope, scopeID, name string) (*entity.Metric, error)
}
