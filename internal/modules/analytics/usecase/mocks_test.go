package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
)

// mockMetricRepository implements IMetricRepository for testing
type mockMetricRepository struct {
	mock.Mock
}

func (m *mockMetricRepository) Store(ctx context.Context, metric *entity.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *mockMetricRepository) GetByID(ctx context.Context, id string) (*entity.Metric, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Metric), args.Error(1)
}

func (m *mockMetricRepository) ListByScope(ctx context.Context, scope, scopeID string, limit, offset int) ([]*entity.Metric, int64, error) {
	args := m.Called(ctx, scope, scopeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Metric), args.Get(1).(int64), args.Error(2)
}

func (m *mockMetricRepository) ListByModule(ctx context.Context, module string, limit, offset int) ([]*entity.Metric, int64, error) {
	args := m.Called(ctx, module, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Metric), args.Get(1).(int64), args.Error(2)
}

func (m *mockMetricRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	args := m.Called(ctx, days)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockMetricRepository) GetLatestByName(ctx context.Context, module, scope, scopeID, name string) (*entity.Metric, error) {
	args := m.Called(ctx, module, scope, scopeID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Metric), args.Error(1)
}
