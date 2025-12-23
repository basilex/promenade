package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/internal/modules/analytics/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MetricRepository struct {
	db *sqlx.DB
}

func NewMetricRepository(db *sqlx.DB) repository.MetricRepository {
	return &MetricRepository{db: db}
}

func (r *MetricRepository) Store(ctx context.Context, metric *entity.Metric) error {
	if metric.ID == "" {
		metric.ID = uuidv7.New().String()
	}

	query := `
		INSERT INTO analytics_metrics (
			id, module, scope, scope_id, name, type, value, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.db.ExecContext(ctx, query,
		metric.ID, metric.Module, metric.Scope, metric.ScopeID,
		metric.Name, metric.Type, metric.Value, metric.Metadata, metric.CreatedAt)

	return err
}

func (r *MetricRepository) GetByID(ctx context.Context, id string) (*entity.Metric, error) {
	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics WHERE id = $1`

	var metric entity.Metric
	err := r.db.GetContext(ctx, &metric, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metric not found: %s", id)
	}
	return &metric, err
}

func (r *MetricRepository) ListByScope(ctx context.Context, scope, scopeID string, limit, offset int) ([]*entity.Metric, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM analytics_metrics WHERE scope = $1 AND scope_id = $2", scope, scopeID); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics WHERE scope = $1 AND scope_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`

	var metrics []*entity.Metric
	if err := r.db.SelectContext(ctx, &metrics, query, scope, scopeID, limit, offset); err != nil {
		return nil, 0, err
	}

	return metrics, total, nil
}

func (r *MetricRepository) ListByModule(ctx context.Context, module string, limit, offset int) ([]*entity.Metric, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM analytics_metrics WHERE module = $1", module); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics WHERE module = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var metrics []*entity.Metric
	if err := r.db.SelectContext(ctx, &metrics, query, module, limit, offset); err != nil {
		return nil, 0, err
	}

	return metrics, total, nil
}

func (r *MetricRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM analytics_metrics WHERE created_at < NOW() - INTERVAL '1 day' * $1", days)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return count, nil
}

func (r *MetricRepository) GetLatestByName(ctx context.Context, module, scope, scopeID, name string) (*entity.Metric, error) {
	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics
		WHERE module = $1 AND scope = $2 AND scope_id = $3 AND name = $4
		ORDER BY created_at DESC LIMIT 1`

	var metric entity.Metric
	err := r.db.GetContext(ctx, &metric, query, module, scope, scopeID, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metric not found")
	}
	return &metric, err
}
