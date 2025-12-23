package entity

import (
	"fmt"
	"time"
)

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

type MetricScope string

const (
	MetricScopeUser    MetricScope = "user"
	MetricScopePost    MetricScope = "post"
	MetricScopeComment MetricScope = "comment"
	MetricScopeSystem  MetricScope = "system"
)

type Metric struct {
	ID        string                 `db:"id" json:"id"`
	Module    string                 `db:"module" json:"module"`
	Scope     MetricScope            `db:"scope" json:"scope"`
	ScopeID   string                 `db:"scope_id" json:"scope_id"`
	Name      string                 `db:"name" json:"name"`
	Type      MetricType             `db:"type" json:"type"`
	Value     float64                `db:"value" json:"value"`
	Metadata  map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
}

func (m *Metric) Validate() error {
	if m.Module == "" {
		return ErrInvalidModule
	}
	if m.Name == "" {
		return ErrInvalidMetricName
	}
	if m.Type != MetricTypeCounter && m.Type != MetricTypeGauge && m.Type != MetricTypeHistogram {
		return ErrInvalidMetricType
	}
	if m.Scope != MetricScopeUser && m.Scope != MetricScopePost && m.Scope != MetricScopeComment && m.Scope != MetricScopeSystem {
		return ErrInvalidMetricScope
	}
	return nil
}

func (m *Metric) TableName() string {
	return "analytics_metrics"
}

type MetricAggregate struct {
	ID        string      `db:"id"`
	MetricName string     `db:"metric_name"`
	Interval  string      `db:"interval"`
	StartTime time.Time   `db:"start_time"`
	EndTime   time.Time   `db:"end_time"`
	Count     int64       `db:"count"`
	Sum       float64     `db:"sum"`
	Min       float64     `db:"min"`
	Max       float64     `db:"max"`
	Avg       float64     `db:"avg"`
	CreatedAt time.Time   `db:"created_at"`
}

func (ma *MetricAggregate) Validate() error {
	if ma.MetricName == "" {
		return fmt.Errorf("metric name is required")
	}
	return nil
}

func (ma *MetricAggregate) TableName() string {
	return "analytics_metric_aggregates"
}
