package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetric_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metric  *Metric
		wantErr error
	}{
		{
			name: "valid counter metric",
			metric: &Metric{
				ID:        "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				IModule:    "posts",
				Scope:     MetricScopePost,
				ScopeID:   "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:      "post_views",
				Type:      MetricTypeCounter,
				Value:     100,
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid gauge metric",
			metric: &Metric{
				ID:        "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				IModule:    "system",
				Scope:     MetricScopeSystem,
				ScopeID:   "system",
				Name:      "cpu_usage",
				Type:      MetricTypeGauge,
				Value:     75.5,
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid histogram metric",
			metric: &Metric{
				ID:        "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				IModule:    "api",
				Scope:     MetricScopeSystem,
				ScopeID:   "api",
				Name:      "request_duration",
				Type:      MetricTypeHistogram,
				Value:     250.5,
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty module",
			metric: &Metric{
				IModule:    "",
				Scope:     MetricScopeUser,
				Name:      "login_count",
				Type:      MetricTypeCounter,
				Value:     1,
				CreatedAt: time.Now(),
			},
			wantErr: ErrInvalidModule,
		},
		{
			name: "empty name",
			metric: &Metric{
				IModule:    "users",
				Scope:     MetricScopeUser,
				Name:      "",
				Type:      MetricTypeCounter,
				Value:     1,
				CreatedAt: time.Now(),
			},
			wantErr: ErrInvalidMetricName,
		},
		{
			name: "invalid type",
			metric: &Metric{
				IModule:    "users",
				Scope:     MetricScopeUser,
				Name:      "login_count",
				Type:      "invalid",
				Value:     1,
				CreatedAt: time.Now(),
			},
			wantErr: ErrInvalidMetricType,
		},
		{
			name: "invalid scope",
			metric: &Metric{
				IModule:    "users",
				Scope:     "invalid",
				Name:      "login_count",
				Type:      MetricTypeCounter,
				Value:     1,
				CreatedAt: time.Now(),
			},
			wantErr: ErrInvalidMetricScope,
		},
		{
			name: "user scope with metadata",
			metric: &Metric{
				IModule:  "users",
				Scope:   MetricScopeUser,
				ScopeID: "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:    "profile_updates",
				Type:    MetricTypeCounter,
				Value:   1,
				Metadata: map[string]interface{}{
					"field":     "name",
					"old_value": "John",
					"new_value": "John Doe",
				},
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "comment scope metric",
			metric: &Metric{
				IModule:    "posts",
				Scope:     MetricScopeComment,
				ScopeID:   "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				Name:      "comment_likes",
				Type:      MetricTypeCounter,
				Value:     5,
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetric_TableName(t *testing.T) {
	metric := &Metric{}
	assert.Equal(t, "analytics_metrics", metric.TableName())
}

func TestMetricAggregate_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		aggregate *MetricAggregate
		wantErr   bool
	}{
		{
			name: "valid aggregate",
			aggregate: &MetricAggregate{
				ID:         "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				MetricName: "post_views",
				Interval:   "1h",
				StartTime:  now.Add(-1 * time.Hour),
				EndTime:    now,
				Count:      100,
				Sum:        500,
				Min:        1,
				Max:        20,
				Avg:        5,
				CreatedAt:  now,
			},
			wantErr: false,
		},
		{
			name: "empty metric name",
			aggregate: &MetricAggregate{
				ID:        "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				MetricName: "",
				Interval:  "1h",
				StartTime: now.Add(-1 * time.Hour),
				EndTime:   now,
				Count:     100,
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "daily aggregate",
			aggregate: &MetricAggregate{
				ID:         "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
				MetricName: "daily_active_users",
				Interval:   "24h",
				StartTime:  now.Add(-24 * time.Hour),
				EndTime:    now,
				Count:      1500,
				Sum:        1500,
				Min:        1,
				Max:        1,
				Avg:        1,
				CreatedAt:  now,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.aggregate.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricAggregate_TableName(t *testing.T) {
	aggregate := &MetricAggregate{}
	assert.Equal(t, "analytics_metric_aggregates", aggregate.TableName())
}

func TestMetricType_Constants(t *testing.T) {
	assert.Equal(t, MetricType("counter"), MetricTypeCounter)
	assert.Equal(t, MetricType("gauge"), MetricTypeGauge)
	assert.Equal(t, MetricType("histogram"), MetricTypeHistogram)
}

func TestMetricScope_Constants(t *testing.T) {
	assert.Equal(t, MetricScope("user"), MetricScopeUser)
	assert.Equal(t, MetricScope("post"), MetricScopePost)
	assert.Equal(t, MetricScope("comment"), MetricScopeComment)
	assert.Equal(t, MetricScope("system"), MetricScopeSystem)
}
