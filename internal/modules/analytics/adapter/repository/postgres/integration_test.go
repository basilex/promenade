//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/analytics/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestMetricRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)

	repo := postgres.NewMetricRepository(testDB.DB)
	ctx := testDB.GetContext()

	t.Run("Store and GetByID", func(t *testing.T) {
		metric := &entity.Metric{
			ID:        uuidv7.New().String(),
			IModule:    "posts",
			Scope:     entity.MetricScopePost,
			ScopeID:   uuidv7.New().String(),
			Name:      "post_views",
			Type:      entity.MetricTypeCounter,
			Value:     100.0,
			Metadata:  map[string]interface{}{"source": "web"},
			CreatedAt: time.Now(),
		}

		err := repo.Store(ctx, metric)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, metric.ID)
		require.NoError(t, err)
		assert.Equal(t, metric.IModule, retrieved.IModule)
		assert.Equal(t, metric.Name, retrieved.Name)
		assert.Equal(t, metric.Value, retrieved.Value)
		assert.Equal(t, metric.Scope, retrieved.Scope)
	})

	t.Run("List by module", func(t *testing.T) {
		// Create multiple metrics
		for i := 0; i < 3; i++ {
			metric := &entity.Metric{
				ID:        uuidv7.New().String(),
				IModule:    "profiles",
				Scope:     entity.MetricScopeUser,
				ScopeID:   uuidv7.New().String(),
				Name:      "profile_updates",
				Type:      entity.MetricTypeCounter,
				Value:     float64(i + 1),
				CreatedAt: time.Now(),
			}
			require.NoError(t, repo.Store(ctx, metric))
		}

		metrics, total, err := repo.ListByModule(ctx, "profiles", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(metrics), 3)
		assert.GreaterOrEqual(t, total, int64(3))
	})

	t.Run("List by scope", func(t *testing.T) {
		scopeID := uuidv7.New().String()
		// Create metrics with same scope
		for i := 0; i < 5; i++ {
			metric := &entity.Metric{
				ID:        uuidv7.New().String(),
				IModule:    "analytics",
				Scope:     entity.MetricScopeSystem,
				ScopeID:   scopeID,
				Name:      "api_calls",
				Type:      entity.MetricTypeCounter,
				Value:     10.0,
				CreatedAt: time.Now(),
			}
			require.NoError(t, repo.Store(ctx, metric))
		}

		metrics, total, err := repo.ListByScope(ctx, string(entity.MetricScopeSystem), scopeID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(metrics), 5)
		assert.GreaterOrEqual(t, total, int64(5))
	})

	t.Run("Delete old metrics", func(t *testing.T) {
		// Create an old metric (3 days ago, so it will be deleted by "older than 2 days")
		oldTime := time.Now().Add(-72 * time.Hour)
		oldMetric := &entity.Metric{
			ID:        uuidv7.New().String(),
			IModule:    "test",
			Scope:     entity.MetricScopeSystem,
			ScopeID:   "test",
			Name:      "old_metric",
			Type:      entity.MetricTypeCounter,
			Value:     1.0,
			CreatedAt: oldTime,
		}
		require.NoError(t, repo.Store(ctx, oldMetric))

		// Delete metrics older than 2 days (48 hours)
		deletedCount, err := repo.DeleteOlderThan(ctx, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deletedCount, int64(1))

		// Verify old metric is deleted
		_, err = repo.GetByID(ctx, oldMetric.ID)
		assert.Error(t, err)
	})

	t.Run("Get latest by name", func(t *testing.T) {
		module := "test_latest_module"
		scopeID := "test_scope"
		metricName := "test_metric"
		
		// Create multiple metrics with same name
		for i := 0; i < 3; i++ {
			metric := &entity.Metric{
				ID:        uuidv7.New().String(),
				IModule:    module,
				Scope:     entity.MetricScopeSystem,
				ScopeID:   scopeID,
				Name:      metricName,
				Type:      entity.MetricTypeGauge,
				Value:     float64(i),
				CreatedAt: time.Now(),
			}
			require.NoError(t, repo.Store(ctx, metric))
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		latest, err := repo.GetLatestByName(ctx, module, string(entity.MetricScopeSystem), scopeID, metricName)
		require.NoError(t, err)
		assert.NotNil(t, latest)
		assert.Equal(t, metricName, latest.Name)
	})
}
