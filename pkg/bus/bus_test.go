package bus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBusConfig(t *testing.T) {
	tests := []struct {
		name            string
		workerPoolSize  int
		bufferSize      int
		retryAttempts   int
		retryDelay      time.Duration
		retryMaxDelay   time.Duration
		retryMultiplier float64
	}{
		{
			name:            "default values",
			workerPoolSize:  10,
			bufferSize:      1000,
			retryAttempts:   3,
			retryDelay:      1 * time.Second,
			retryMaxDelay:   5 * time.Second,
			retryMultiplier: 2.0,
		},
		{
			name:            "custom values",
			workerPoolSize:  20,
			bufferSize:      2000,
			retryAttempts:   5,
			retryDelay:      2 * time.Second,
			retryMaxDelay:   10 * time.Second,
			retryMultiplier: 3.5,
		},
		{
			name:            "minimal values",
			workerPoolSize:  1,
			bufferSize:      10,
			retryAttempts:   1,
			retryDelay:      100 * time.Millisecond,
			retryMaxDelay:   500 * time.Millisecond,
			retryMultiplier: 1.5,
		},
		{
			name:            "large values",
			workerPoolSize:  100,
			bufferSize:      10000,
			retryAttempts:   10,
			retryDelay:      5 * time.Second,
			retryMaxDelay:   1 * time.Minute,
			retryMultiplier: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewBusConfig(
				tt.workerPoolSize,
				tt.bufferSize,
				tt.retryAttempts,
				tt.retryDelay,
				tt.retryMaxDelay,
				tt.retryMultiplier,
			)

			assert.Equal(t, tt.workerPoolSize, config.WorkerPoolSize)
			assert.Equal(t, tt.bufferSize, config.BufferSize)
			assert.True(t, config.EnableMetrics, "metrics should always be enabled")

			assert.NotNil(t, config.RetryPolicy, "retry policy should not be nil")
			assert.Equal(t, tt.retryAttempts, config.RetryPolicy.MaxAttempts)
			assert.Equal(t, tt.retryDelay, config.RetryPolicy.InitialDelay)
			assert.Equal(t, tt.retryMaxDelay, config.RetryPolicy.MaxDelay)
			assert.Equal(t, tt.retryMultiplier, config.RetryPolicy.Multiplier)
		})
	}
}

func TestBusConfig_RetryPolicyNotNil(t *testing.T) {
	config := NewBusConfig(10, 1000, 3, 1*time.Second, 5*time.Second, 2.0)

	assert.NotNil(t, config.RetryPolicy, "RetryPolicy should never be nil")
}

func TestBusConfig_EnableMetricsAlwaysTrue(t *testing.T) {
	configs := []BusConfig{
		NewBusConfig(10, 1000, 3, 1*time.Second, 5*time.Second, 2.0),
		NewBusConfig(1, 10, 1, 100*time.Millisecond, 1*time.Second, 1.5),
		NewBusConfig(100, 10000, 10, 5*time.Second, 1*time.Minute, 5.0),
	}

	for i, config := range configs {
		assert.True(t, config.EnableMetrics, "EnableMetrics should always be true (test case %d)", i)
	}
}

func TestRetryPolicy_AllFieldsSet(t *testing.T) {
	config := NewBusConfig(10, 1000, 3, 1*time.Second, 5*time.Second, 2.0)

	assert.NotZero(t, config.RetryPolicy.MaxAttempts, "MaxAttempts should be set")
	assert.NotZero(t, config.RetryPolicy.InitialDelay, "InitialDelay should be set")
	assert.NotZero(t, config.RetryPolicy.MaxDelay, "MaxDelay should be set")
	assert.NotZero(t, config.RetryPolicy.Multiplier, "Multiplier should be set")
}

func TestBusConfig_ZeroValues(t *testing.T) {
	config := NewBusConfig(0, 0, 0, 0, 0, 0.0)

	assert.Equal(t, 0, config.WorkerPoolSize)
	assert.Equal(t, 0, config.BufferSize)
	assert.True(t, config.EnableMetrics)

	assert.NotNil(t, config.RetryPolicy)
	assert.Equal(t, 0, config.RetryPolicy.MaxAttempts)
	assert.Equal(t, time.Duration(0), config.RetryPolicy.InitialDelay)
	assert.Equal(t, time.Duration(0), config.RetryPolicy.MaxDelay)
	assert.Equal(t, 0.0, config.RetryPolicy.Multiplier)
}

func TestRetryPolicy_BackoffCalculation(t *testing.T) {
	policy := &RetryPolicy{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}

	delays := []time.Duration{}
	currentDelay := policy.InitialDelay

	for i := 0; i < policy.MaxAttempts; i++ {
		if currentDelay > policy.MaxDelay {
			currentDelay = policy.MaxDelay
		}
		delays = append(delays, currentDelay)
		currentDelay = time.Duration(float64(currentDelay) * policy.Multiplier)
	}

	assert.Equal(t, 1*time.Second, delays[0])
	assert.Equal(t, 2*time.Second, delays[1])
	assert.Equal(t, 4*time.Second, delays[2])
	assert.Equal(t, 8*time.Second, delays[3])
	assert.Equal(t, 16*time.Second, delays[4])

	for _, delay := range delays {
		assert.LessOrEqual(t, delay, policy.MaxDelay)
	}
}

func TestNewBusConfig_DifferentMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		multiplier float64
	}{
		{"multiplier 1.5", 1.5},
		{"multiplier 2.0", 2.0},
		{"multiplier 3.0", 3.0},
		{"multiplier 5.0", 5.0},
		{"multiplier 10.0", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewBusConfig(10, 1000, 3, 1*time.Second, 5*time.Second, tt.multiplier)
			assert.Equal(t, tt.multiplier, config.RetryPolicy.Multiplier)
		})
	}
}

func TestNewBusConfig_DifferentMaxDelays(t *testing.T) {
	tests := []struct {
		name     string
		maxDelay time.Duration
	}{
		{"100ms", 100 * time.Millisecond},
		{"1s", 1 * time.Second},
		{"5s", 5 * time.Second},
		{"30s", 30 * time.Second},
		{"1m", 1 * time.Minute},
		{"1h", 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewBusConfig(10, 1000, 3, 1*time.Second, tt.maxDelay, 2.0)
			assert.Equal(t, tt.maxDelay, config.RetryPolicy.MaxDelay)
		})
	}
}
