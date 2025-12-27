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
var MemoryBusFactory func(Config) IBus

// RedisBusFactory is a function type for creating Redis bus instances
var RedisBusFactory func(config.BusSection, Config) (IBus, error)

// NewBus creates a new event bus based on configuration.
// It implements graceful degradation:
// - Development/Test: Returns error if adapter initialization fails
// - Production: Falls back to in-memory adapter with warning log
func NewBus(cfg config.BusSection) (IBus, error) {
	adapterType := AdapterType(strings.ToLower(cfg.Adapter))

	// Validate adapter type
	switch adapterType {
	case AdapterMemory, AdapterRedis:
		// Valid adapter types
	default:
		return nil, fmt.Errorf("unsupported bus adapter: %s (supported: memory, redis)", cfg.Adapter)
	}

	// Create bus configuration
	busConfig := NewConfig(
		cfg.WorkerPoolSize,
		cfg.BufferSize,
		cfg.RetryAttempts,
		cfg.RetryDelay,
		cfg.RetryMaxDelay,
		cfg.RetryMultiplier,
	)

	var eventBus IBus
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
			// Graceful degradation: fallback to in-memory
			logger.Warn("Failed to initialize Redis event bus, falling back to in-memory adapter",
				slog.Any("error", err),
				slog.String("fallback", "memory"))

			if MemoryBusFactory == nil {
				return nil, fmt.Errorf("cannot fallback to memory bus: factory not initialized")
			}

			eventBus = MemoryBusFactory(busConfig)
			logger.Info("Event bus initialized (fallback)",
				slog.String("adapter", "memory"))
		} else {
			logger.Info("Event bus initialized",
				slog.String("adapter", "redis"),
				slog.String("host", cfg.Redis.Host),
				slog.Int("port", cfg.Redis.Port))
		}
	}

	return eventBus, nil
}

// MustNewBus creates a new bus or panics on error.
func MustNewBus(cfg config.BusSection) IBus {
	eventBus, err := NewBus(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize event bus: %v", err))
	}
	return eventBus
}

// HealthCheck performs a health check on the bus
func HealthCheck(ctx context.Context, eventBus IBus) error {
	if eventBus == nil {
		return fmt.Errorf("bus is nil")
	}
	return eventBus.Health(ctx)
}
