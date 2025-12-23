package entity

import "errors"

var (
	ErrInvalidModule      = errors.New("invalid module name")
	ErrInvalidMetricName  = errors.New("invalid metric name")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricScope = errors.New("invalid metric scope")
	ErrInvalidMetricValue = errors.New("invalid metric value")
)
