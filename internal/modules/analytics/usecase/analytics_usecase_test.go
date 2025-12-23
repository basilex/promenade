package usecase

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/analytics/domain/entity"
	"github.com/basilex/promenade/internal/modules/analytics/usecase/mocks"
)

func TestAnalyticsUseCase_CollectMetric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name    string
		input   *CollectMetricInput
		mockFn  func(*mocks.MockMetricRepository)
		wantErr error
	}{
		{
			name: "successfully collect counter metric",
			input: &CollectMetricInput{
				Module:  "posts",
				Scope:   entity.MetricScopePost,
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "post_views",
				Type:    entity.MetricTypeCounter,
				Value:   1,
			},
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("Store", mock.Anything, mock.MatchedBy(func(m *entity.Metric) bool {
					return m.Module == "posts" && m.Name == "post_views" && m.Type == entity.MetricTypeCounter && m.Value == 1
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "successfully collect gauge metric",
			input: &CollectMetricInput{
				Module:  "system",
				Scope:   entity.MetricScopeSystem,
				ScopeID: "system",
				Name:    "memory_usage",
				Type:    entity.MetricTypeGauge,
				Value:   75.5,
			},
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("Store", mock.Anything, mock.MatchedBy(func(m *entity.Metric) bool {
					return m.Module == "system" && m.Name == "memory_usage" && m.Type == entity.MetricTypeGauge && m.Value == 75.5
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "successfully collect histogram metric with metadata",
			input: &CollectMetricInput{
				Module:  "api",
				Scope:   entity.MetricScopeSystem,
				ScopeID: "api",
				Name:    "request_duration",
				Type:    entity.MetricTypeHistogram,
				Value:   250.5,
				Metadata: map[string]interface{}{
					"endpoint": "/api/v1/posts",
					"method":   "GET",
					"status":   200,
				},
			},
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("Store", mock.Anything, mock.MatchedBy(func(m *entity.Metric) bool {
					return m.Module == "api" && m.Name == "request_duration" && m.Metadata != nil && len(m.Metadata) == 3
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "empty module name",
			input: &CollectMetricInput{
				Module:  "",
				Scope:   entity.MetricScopeUser,
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "login_count",
				Type:    entity.MetricTypeCounter,
				Value:   1,
			},
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: entity.ErrInvalidModule,
		},
		{
			name: "empty metric name",
			input: &CollectMetricInput{
				Module:  "users",
				Scope:   entity.MetricScopeUser,
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "",
				Type:    entity.MetricTypeCounter,
				Value:   1,
			},
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: entity.ErrInvalidMetricName,
		},
		{
			name: "invalid metric type",
			input: &CollectMetricInput{
				Module:  "users",
				Scope:   entity.MetricScopeUser,
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "login_count",
				Type:    "invalid",
				Value:   1,
			},
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: entity.ErrInvalidMetricType,
		},
		{
			name: "invalid metric scope",
			input: &CollectMetricInput{
				Module:  "users",
				Scope:   "invalid",
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "login_count",
				Type:    entity.MetricTypeCounter,
				Value:   1,
			},
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: entity.ErrInvalidMetricScope,
		},
		{
			name: "repository store error",
			input: &CollectMetricInput{
				Module:  "posts",
				Scope:   entity.MetricScopePost,
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "post_likes",
				Type:    entity.MetricTypeCounter,
				Value:   1,
			},
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("Store", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			wantErr: errors.New("failed to collect metric"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMetricRepository()
			tt.mockFn(repo)

			uc := NewAnalyticsUseCase(repo, logger)
			metric, err := uc.CollectMetric(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, metric)
				if tt.wantErr == entity.ErrInvalidModule || tt.wantErr == entity.ErrInvalidMetricName || tt.wantErr == entity.ErrInvalidMetricType || tt.wantErr == entity.ErrInvalidMetricScope {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metric)
				assert.Equal(t, tt.input.Module, metric.Module)
				assert.Equal(t, tt.input.Name, metric.Name)
				assert.Equal(t, tt.input.Type, metric.Type)
				assert.Equal(t, tt.input.Value, metric.Value)
				assert.Equal(t, tt.input.Scope, metric.Scope)
				assert.Equal(t, tt.input.ScopeID, metric.ScopeID)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestAnalyticsUseCase_GetMetric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.MockMetricRepository)
		wantErr bool
	}{
		{
			name: "successfully get metric",
			id:   "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("GetByID", mock.Anything, "01936d6a-8f7c-7890-a1b2-c3d4e5f67890").Return(&entity.Metric{
					ID:        "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
					Module:    "posts",
					Name:      "post_views",
					Type:      entity.MetricTypeCounter,
					Value:     100,
					CreatedAt: time.Now(),
				}, nil)
			},
			wantErr: false,
		},
		{
			name:    "empty metric ID",
			id:      "",
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: true,
		},
		{
			name: "metric not found",
			id:   "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("GetByID", mock.Anything, "01936d6a-8f7c-7890-a1b2-c3d4e5f67890").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMetricRepository()
			tt.mockFn(repo)

			uc := NewAnalyticsUseCase(repo, logger)
			metric, err := uc.GetMetric(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metric)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metric)
				assert.Equal(t, tt.id, metric.ID)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestAnalyticsUseCase_ListMetricsByScope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name    string
		scope   entity.MetricScope
		scopeID string
		limit   int
		offset  int
		mockFn  func(*mocks.MockMetricRepository)
		wantErr bool
	}{
		{
			name:    "successfully list metrics by user scope",
			scope:   entity.MetricScopeUser,
			scopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			limit:   20,
			offset:  0,
			mockFn: func(repo *mocks.MockMetricRepository) {
				metrics := []*entity.Metric{
					{
						ID:      "metric-1",
						Module:  "users",
						Scope:   entity.MetricScopeUser,
						ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
						Name:    "profile_views",
						Type:    entity.MetricTypeCounter,
						Value:   10,
					},
					{
						ID:      "metric-2",
						Module:  "users",
						Scope:   entity.MetricScopeUser,
						ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
						Name:    "profile_updates",
						Type:    entity.MetricTypeCounter,
						Value:   5,
					},
				}
				repo.On("ListByScope", mock.Anything, "user", "01936d6a-8f7c-7890-a1b2-c3d4e5f67890", 20, 0).Return(metrics, int64(2), nil)
			},
			wantErr: false,
		},
		{
			name:    "empty scope ID",
			scope:   entity.MetricScopeUser,
			scopeID: "",
			limit:   20,
			offset:  0,
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: true,
		},
		{
			name:    "repository error",
			scope:   entity.MetricScopePost,
			scopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			limit:   20,
			offset:  0,
			mockFn: func(repo *mocks.MockMetricRepository) {
				repo.On("ListByScope", mock.Anything, "post", "01936d6a-8f7c-7890-a1b2-c3d4e5f67890", 20, 0).Return(nil, int64(0), errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMetricRepository()
			tt.mockFn(repo)

			uc := NewAnalyticsUseCase(repo, logger)
			metrics, total, err := uc.ListMetricsByScope(context.Background(), tt.scope, tt.scopeID, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metrics)
				assert.Zero(t, total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metrics)
				assert.Greater(t, total, int64(0))
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestAnalyticsUseCase_ListMetricsByModule(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name    string
		module  string
		limit   int
		offset  int
		mockFn  func(*mocks.MockMetricRepository)
		wantErr bool
	}{
		{
			name:   "successfully list metrics by module",
			module: "posts",
			limit:  20,
			offset: 0,
			mockFn: func(repo *mocks.MockMetricRepository) {
				metrics := []*entity.Metric{
					{ID: "metric-1", Module: "posts", Name: "post_views", Type: entity.MetricTypeCounter, Value: 100},
					{ID: "metric-2", Module: "posts", Name: "post_likes", Type: entity.MetricTypeCounter, Value: 50},
				}
				repo.On("ListByModule", mock.Anything, "posts", 20, 0).Return(metrics, int64(2), nil)
			},
			wantErr: false,
		},
		{
			name:    "empty module name",
			module:  "",
			limit:   20,
			offset:  0,
			mockFn:  func(repo *mocks.MockMetricRepository) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMetricRepository()
			tt.mockFn(repo)

			uc := NewAnalyticsUseCase(repo, logger)
			metrics, total, err := uc.ListMetricsByModule(context.Background(), tt.module, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metrics)
				assert.Zero(t, total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metrics)
				assert.Greater(t, total, int64(0))
				for _, m := range metrics {
					assert.Equal(t, tt.module, m.Module)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestAnalyticsUseCase_GetLatestMetric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name       string
		module     string
		scope      entity.MetricScope
		scopeID    string
		metricName string
		mockFn     func(*mocks.MockMetricRepository)
		wantErr    bool
	}{
		{
			name:       "successfully get latest metric",
			module:     "posts",
			scope:      entity.MetricScopePost,
			scopeID:    "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			metricName: "post_views",
			mockFn: func(repo *mocks.MockMetricRepository) {
				metric := &entity.Metric{
					ID:        "metric-1",
					Module:    "posts",
					Scope:     entity.MetricScopePost,
					ScopeID:   "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
					Name:      "post_views",
					Type:      entity.MetricTypeCounter,
					Value:     150,
					CreatedAt: time.Now(),
				}
				repo.On("GetLatestByName", mock.Anything, "posts", "post", "01936d6a-8f7c-7890-a1b2-c3d4e5f67890", "post_views").Return(metric, nil)
			},
			wantErr: false,
		},
		{
			name:       "empty module",
			module:     "",
			scope:      entity.MetricScopePost,
			scopeID:    "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			metricName: "post_views",
			mockFn:     func(repo *mocks.MockMetricRepository) {},
			wantErr:    true,
		},
		{
			name:       "empty scope ID",
			module:     "posts",
			scope:      entity.MetricScopePost,
			scopeID:    "",
			metricName: "post_views",
			mockFn:     func(repo *mocks.MockMetricRepository) {},
			wantErr:    true,
		},
		{
			name:       "empty metric name",
			module:     "posts",
			scope:      entity.MetricScopePost,
			scopeID:    "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
			metricName: "",
			mockFn:     func(repo *mocks.MockMetricRepository) {},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMetricRepository()
			tt.mockFn(repo)

			uc := NewAnalyticsUseCase(repo, logger)
			metric, err := uc.GetLatestMetric(context.Background(), tt.module, tt.scope, tt.scopeID, tt.metricName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metric)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metric)
				assert.Equal(t, tt.module, metric.Module)
				assert.Equal(t, tt.scopeID, metric.ScopeID)
				assert.Equal(t, tt.metricName, metric.Name)
			}

			repo.AssertExpectations(t)
		})
	}
}
