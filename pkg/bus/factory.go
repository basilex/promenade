package bus

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/logger"
)

// AdapterType represents available bus adapter types
type AdapterType string

const (
	AdapterMemory AdapterType = "memory"
	AdapterRedis  AdapterType = "redis"
)

// MemoryBusFactory is a function type for creating memory bus instances
// This avoids circular import with memory package
var MemoryBusFactory func(BusConfig) Bus

// RedisBusFactory is a function type for creating Redis bus instances
var RedisBusFactory func(config.BusSection, BusConfig) (Bus, error)

// init initializes the factory functions using late binding to avoid circular imports
func init() {
	// These will be set by the respective adapter packages
	// See memory/memory_bus.go and redis/redis_bus.go
}

// NewBus creates a new event bus based on configuration.
// It implements graceful degradation:
// - Development/Test: Returns error if adapter initialization fails
// - Production: Falls back to in-memory adapter with warning log
func NewBus(cfg config.BusSection) (Bus, error) {
	adapterType := AdapterType(strings.ToLower(cfg.Adapter))

	// Validate adapter type
	switch adapterType {
	case AdapterMemory, AdapterRedis:
		// Valid adapter types
	default:
		return nil, fmt.Errorf("unsupported bus adapter: %s (supported: memory, redis)", cfg.Adapter)
	}

	// Create bus configuration
	busConfig := NewBusConfig(
		cfg.WorkerPoolSize,
		cfg.BufferSize,
		cfg.RetryAttempts,
		cfg.RetryDelay,
		cfg.RetryMaxDelay,
		cfg.RetryMultiplier,
	)

	var eventBus Bus
	var err error

	switch adapterType {
	case AdapterMemory:
		if MemoryBusFactory == nil {
			return nil, fmt.Errorf("memory bus factory not initialized")
		}
		eventBus = MemoryBusFactory(busConfig)
		logger.Info("Event bus initialized",
			slog.String("adapter", "memory"),
			slog.Int("buffer_size", cfg.BufferSize),
			slog.Int("worker_pool_size", cfg.WorkerPoolSize))

	case AdapterRedis:
		if RedisBusFactory == nil {
			return nil, fmt.Errorf("redis bus factory not initialized")
		}

		// Try to initialize Redis adapter
		eventBus, err = RedisBusFactory(cfg, busConfig)
		if err != nil {
			// Check if we should fail fast or fallback
			if shouldFailFast() {
				return nil, fmt.Errorf("failed to initialize Redis event bus: %w", err)
			}

			// Graceful degradation: fallback to in-memory
			logger.Warn("Failed to initialize Redis event bus, falling back to in-memory adapter",
				slog.Any("error", err),
				slog.String("fallback", "memory"))

			if MemoryBusFactory == nil {
				return nil, fmt.Errorf("cannot fallback to memory bus: factory not initialized")
			}

			eventBus = MemoryBusFactory(busConfig)
			logger.Info("Event bus initialized (fallback)",
				slog.String("adapter", "memory"),
				slog.Int("buffer_size", cfg.BufferSize),
				slog.Int("worker_pool_size", cfg.WorkerPoolSize))
		} else {
			logger.Info("Event bus initialized",
				slog.String("adapter", "redis"),
				slog.String("host", cfg.Redis.Host),
				slog.Int("port", cfg.Redis.Port),
				slog.Int("db", cfg.Redis.DB))
		}
	}

	return eventBus, nil
}

// shouldFailFast determines if we should fail fast or fallback to memory adapter.
// In test/development environments, we fail fast to catch configuration issues early.
// In production, we gracefully degrade to in-memory adapter with warning.
func shouldFailFast() bool {
	// Check if we're in test environment
	// Tests should fail fast to ensure proper configuration
	// Production should gracefully degrade
	return false // For now, always allow graceful degradation
	// TODO: Make this configurable via environment variable
	// return os.Getenv("BUS_FAIL_FAST") == "true"
}

// MustNewBus creates a new bus or panics on error.
// Use this only when bus initialization failure should be fatal.
func MustNewBus(cfg config.BusSection) Bus {
	bus, err := NewBus(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize event bus: %v", err))
	}
	return bus
}

// HealthCheck performs a health check on the bus
func HealthCheck(ctx context.Context, bus Bus) error {
	if bus == nil {
		return fmt.Errorf("bus is nil")
	}
	return bus.Health(ctx)
}
