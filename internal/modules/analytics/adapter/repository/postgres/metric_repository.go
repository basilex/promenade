package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/internal/modules/analytics/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type IMetricRepository struct {
	db *sqlx.DB
}

func NewMetricRepository(db *sqlx.DB) repository.IMetricRepository {
	return &IMetricRepository{db: db}
}

// metricDB is an internal type for scanning from database
type metricDB struct {
	ID        string          `db:"id"`
	IModule    string          `db:"module"`
	Scope     string          `db:"scope"`
	ScopeID   string          `db:"scope_id"`
	Name      string          `db:"name"`
	Type      string          `db:"type"`
	Value     float64         `db:"value"`
	Metadata  json.RawMessage `db:"metadata"`
	CreatedAt time.Time       `db:"created_at"`
}

func (m *metricDB) toEntity() (*entity.Metric, error) {
	var metadata map[string]interface{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		metadata = make(map[string]interface{})
	}

	return &entity.Metric{
		ID:        m.ID,
		IModule:    m.IModule,
		Scope:     entity.MetricScope(m.Scope),
		ScopeID:   m.ScopeID,
		Name:      m.Name,
		Type:      entity.MetricType(m.Type),
		Value:     m.Value,
		Metadata:  metadata,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *IMetricRepository) Store(ctx context.Context, metric *entity.Metric) error {
	if metric.ID == "" {
		metric.ID = uuidv7.New().String()
	}

	// Convert metadata map to JSON
	metadataJSON, err := json.Marshal(metric.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO analytics_metrics (
			id, module, scope, scope_id, name, type, value, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err = r.db.ExecContext(ctx, query,
		metric.ID, metric.IModule, metric.Scope, metric.ScopeID,
		metric.Name, metric.Type, metric.Value, metadataJSON, metric.CreatedAt)

	return err
}

func (r *IMetricRepository) GetByID(ctx context.Context, id string) (*entity.Metric, error) {
	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics WHERE id = $1`

	var metricDB metricDB
	err := r.db.GetContext(ctx, &metricDB, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metric not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	return metricDB.toEntity()
}

func (r *IMetricRepository) ListByScope(ctx context.Context, scope, scopeID string, limit, offset int) ([]*entity.Metric, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM analytics_metrics WHERE scope = $1 AND scope_id = $2", scope, scopeID); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics WHERE scope = $1 AND scope_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`

	var metricsDB []metricDB
	if err := r.db.SelectContext(ctx, &metricsDB, query, scope, scopeID, limit, offset); err != nil {
		return nil, 0, err
	}

	metrics := make([]*entity.Metric, 0, len(metricsDB))
	for _, mdb := range metricsDB {
		m, err := mdb.toEntity()
		if err != nil {
			return nil, 0, err
		}
		metrics = append(metrics, m)
	}

	return metrics, total, nil
}

func (r *IMetricRepository) ListByModule(ctx context.Context, module string, limit, offset int) ([]*entity.Metric, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM analytics_metrics WHERE module = $1", module); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics WHERE module = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var metricsDB []metricDB
	if err := r.db.SelectContext(ctx, &metricsDB, query, module, limit, offset); err != nil {
		return nil, 0, err
	}

	metrics := make([]*entity.Metric, 0, len(metricsDB))
	for _, mdb := range metricsDB {
		m, err := mdb.toEntity()
		if err != nil {
			return nil, 0, err
		}
		metrics = append(metrics, m)
	}

	return metrics, total, nil
}

func (r *IMetricRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM analytics_metrics WHERE created_at < NOW() - INTERVAL '1 day' * $1", days)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return count, nil
}

func (r *IMetricRepository) GetLatestByName(ctx context.Context, module, scope, scopeID, name string) (*entity.Metric, error) {
	query := `SELECT id, module, scope, scope_id, name, type, value, metadata, created_at
		FROM analytics_metrics
		WHERE module = $1 AND scope = $2 AND scope_id = $3 AND name = $4
		ORDER BY created_at DESC LIMIT 1`

	var metricDB metricDB
	err := r.db.GetContext(ctx, &metricDB, query, module, scope, scopeID, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metric not found")
	}
	if err != nil {
		return nil, err
	}

	return metricDB.toEntity()
}
