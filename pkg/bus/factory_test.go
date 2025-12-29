package bus_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/bus"
	_ "github.com/basilex/promenade/pkg/bus/memory" // Register memory adapter
)

func TestNewBus_Memory(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
		RetryAttempts:  2,
		RetryDelay:     time.Second,
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	require.NoError(t, err)
	require.NotNil(t, b)

	// Health check should pass
	err = b.Health(context.Background())
	assert.NoError(t, err)

	// Close should work
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = b.Close(ctx)
	assert.NoError(t, err)
}

func TestNewBus_MemoryDefaults(t *testing.T) {
	cfg := config.BusSection{
		Adapter: "memory",
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	require.NoError(t, err)
	require.NotNil(t, b)

	defer b.Close(context.Background())

	// Should work with default values
	err = b.Health(context.Background())
	assert.NoError(t, err)
}

func TestNewBus_InvalidAdapter(t *testing.T) {
	cfg := config.BusSection{
		Adapter: "invalid-adapter",
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	assert.Error(t, err)
	assert.Nil(t, b)
	assert.Contains(t, err.Error(), "unsupported bus adapter")
}

func TestNewBus_EmptyAdapter(t *testing.T) {
	cfg := config.BusSection{
		Adapter: "",
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestMustNewBus_Success(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 10,
		BufferSize:     1000,
	}

	// Should not panic
	b := bus.MustNewBus(cfg, config.RedisSection{})
	require.NotNil(t, b)
	defer b.Close(context.Background())

	err := b.Health(context.Background())
	assert.NoError(t, err)
}

func TestMustNewBus_Panics(t *testing.T) {
	cfg := config.BusSection{
		Adapter: "invalid-adapter",
	}

	assert.Panics(t, func() {
		bus.MustNewBus(cfg, config.RedisSection{})
	}, "MustNewBus should panic with invalid adapter")
}

func TestNewConfig(t *testing.T) {
	cfg := bus.NewConfig(10, 1000, 3, time.Second, 10*time.Second, 2.0)

	assert.Equal(t, 10, cfg.WorkerPoolSize)
	assert.Equal(t, 1000, cfg.BufferSize)
	require.NotNil(t, cfg.RetryPolicy)
	assert.Equal(t, 3, cfg.RetryPolicy.MaxAttempts)
	assert.Equal(t, time.Second, cfg.RetryPolicy.InitialDelay)
	assert.Equal(t, 10*time.Second, cfg.RetryPolicy.MaxDelay)
	assert.Equal(t, 2.0, cfg.RetryPolicy.Multiplier)
}

func TestNewConfig_WithZeroValues(t *testing.T) {
	cfg := bus.NewConfig(0, 0, 0, 0, 0, 0)

	// Should create config even with zero values
	assert.Equal(t, 0, cfg.WorkerPoolSize)
	assert.Equal(t, 0, cfg.BufferSize)
	require.NotNil(t, cfg.RetryPolicy)
	assert.Equal(t, 0, cfg.RetryPolicy.MaxAttempts)
}

func TestNewConfig_RetryPolicy(t *testing.T) {
	tests := []struct {
		name             string
		maxAttempts      int
		initialDelay     time.Duration
		maxDelay         time.Duration
		multiplier       float64
		expectedAttempts int
	}{
		{
			name:             "Standard retry",
			maxAttempts:      3,
			initialDelay:     100 * time.Millisecond,
			maxDelay:         time.Second,
			multiplier:       2.0,
			expectedAttempts: 3,
		},
		{
			name:             "Aggressive retry",
			maxAttempts:      10,
			initialDelay:     10 * time.Millisecond,
			maxDelay:         5 * time.Second,
			multiplier:       1.5,
			expectedAttempts: 10,
		},
		{
			name:             "No retry",
			maxAttempts:      1,
			initialDelay:     0,
			maxDelay:         0,
			multiplier:       1.0,
			expectedAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := bus.NewConfig(10, 100, tt.maxAttempts, tt.initialDelay, tt.maxDelay, tt.multiplier)
			assert.Equal(t, tt.expectedAttempts, cfg.RetryPolicy.MaxAttempts)
			assert.Equal(t, tt.initialDelay, cfg.RetryPolicy.InitialDelay)
			assert.Equal(t, tt.maxDelay, cfg.RetryPolicy.MaxDelay)
			assert.Equal(t, tt.multiplier, cfg.RetryPolicy.Multiplier)
		})
	}
}

func TestBus_HealthCheck(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	require.NoError(t, err)
	defer b.Close(context.Background())

	// Health check should pass
	err = b.Health(context.Background())
	assert.NoError(t, err)

	// Multiple health checks should work
	for i := 0; i < 5; i++ {
		err = b.Health(context.Background())
		assert.NoError(t, err)
	}
}

func TestBus_HealthCheckAfterClose(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	require.NoError(t, err)

	// Close the bus
	err = b.Close(context.Background())
	require.NoError(t, err)

	// Health check should fail after close
	err = b.Health(context.Background())
	assert.Error(t, err)
}

func TestBus_MultipleClose(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	require.NoError(t, err)

	// First close should work
	err = b.Close(context.Background())
	assert.NoError(t, err)

	// Second close should not error (idempotent)
	err = b.Close(context.Background())
	assert.NoError(t, err)
}

func TestBus_CloseWithTimeout(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
	}

	b, err := bus.NewBus(cfg, config.RedisSection{})
	require.NoError(t, err)

	// Close with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = b.Close(ctx)
	assert.NoError(t, err)
}
