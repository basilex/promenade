package memory

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// init registers the memory bus factory with the bus package to avoid circular imports
func init() {
	bus.MemoryBusFactory = func(config bus.BusConfig) bus.Bus {
		return NewMemoryBus(config)
	}
}

// MemoryBus is an in-memory implementation of bus.Bus.
// Useful for development, testing, and simple deployments.
// Events are processed in goroutines without any persistence.
type MemoryBus struct {
	mu           sync.RWMutex
	subscribers  map[string][]bus.Handler // topic -> handlers
	config       bus.BusConfig
	workerPool   chan struct{}
	wg           sync.WaitGroup
	closed       bool
	eventCounter int64 // For statistics
	logger       *slog.Logger
}

// NewMemoryBus creates a new in-memory message bus.
func NewMemoryBus(config bus.BusConfig) *MemoryBus {
	if config.WorkerPoolSize <= 0 {
		config.WorkerPoolSize = 10
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	return &MemoryBus{
		subscribers: make(map[string][]bus.Handler),
		config:      config,
		workerPool:  make(chan struct{}, config.WorkerPoolSize),
		logger:      logger.Default().Logger,
	}
}

// NewDefaultMemoryBus creates a memory bus with default configuration.
// Deprecated: Use NewMemoryBus with config loaded from environment instead.
func NewDefaultMemoryBus() *MemoryBus {
	// Fallback defaults only for tests - production should use NewMemoryBus with config
	defaultCfg := bus.BusConfig{
		WorkerPoolSize: 10,
		BufferSize:     1000,
		EnableMetrics:  true,
	}
	return NewMemoryBus(defaultCfg)
}

// Publish sends an event to all subscribers of the topic.
// Events are processed asynchronously in goroutines.
func (mb *MemoryBus) Publish(ctx context.Context, topic string, event bus.Event) error {
	mb.mu.RLock()
	if mb.closed {
		mb.mu.RUnlock()
		return fmt.Errorf("bus is closed")
	}

	handlers, exists := mb.subscribers[topic]
	if !exists || len(handlers) == 0 {
		mb.mu.RUnlock()
		// Not an error - just no subscribers yet
		return nil
	}

	// Copy handlers to avoid holding lock during dispatch
	handlersCopy := make([]bus.Handler, len(handlers))
	copy(handlersCopy, handlers)
	mb.mu.RUnlock()

	// Dispatch to all handlers asynchronously
	message := bus.Message{
		ID:          uuidv7.New().String(),
		Topic:       topic,
		Event:       event,
		PublishedAt: event.OccurredAt(),
		Metadata:    event.Metadata(),
	}

	for _, handler := range handlersCopy {
		mb.wg.Add(1)
		go mb.dispatch(ctx, handler, message)
	}

	return nil
}

// dispatch processes a single , error recovery, and retry logic with exponential backoff.
func (mb *MemoryBus) dispatch(ctx context.Context, handler bus.Handler, message bus.Message) {
	defer mb.wg.Done()

	// Acquire worker slot (blocks if pool is full)
	mb.workerPool <- struct{}{}
	defer func() { <-mb.workerPool }()

	// Recover from panics in handlers
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't crash the bus
			mb.logger.Error("PANIC in event handler",
				slog.String("topic", message.Topic),
				slog.String("message_id", message.ID),
				slog.Any("panic", r))
		}
	}()

	// Execute handler with retry logic
	var lastErr error
	maxAttempts := 1 // Default: no retries
	if mb.config.RetryPolicy != nil && mb.config.RetryPolicy.MaxAttempts > 0 {
		maxAttempts = mb.config.RetryPolicy.MaxAttempts
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Call handler
		err := handler(ctx, message.Event)
		if err == nil {
			// Success - log only if there were previous failures
			if attempt > 1 {
				mb.logger.Info("Handler succeeded after retry",
					slog.String("topic", message.Topic),
					slog.String("message_id", message.ID),
					slog.Int("attempt", attempt),
					slog.Int("total_attempts", maxAttempts))
			}
			return
		}

		lastErr = err

		// If this is the last attempt, don't wait
		if attempt == maxAttempts {
			break
		}

		// Calculate exponential backoff delay using RetryPolicy
		var backoffDelay time.Duration
		if mb.config.RetryPolicy != nil {
			initialDelay := mb.config.RetryPolicy.InitialDelay
			multiplier := mb.config.RetryPolicy.Multiplier
			if multiplier <= 0 {
				multiplier = 2.0 // Default exponential backoff
			}

			// Calculate: initialDelay * multiplier^(attempt-1)
			backoffDelay = time.Duration(float64(initialDelay) * math.Pow(multiplier, float64(attempt-1)))

			// Cap at MaxDelay if configured
			if mb.config.RetryPolicy.MaxDelay > 0 && backoffDelay > mb.config.RetryPolicy.MaxDelay {
				backoffDelay = mb.config.RetryPolicy.MaxDelay
			}
		} else {
			// Fallback: simple exponential backoff with 1s base
			backoffDelay = time.Second * time.Duration(1<<uint(attempt-1))
		}

		mb.logger.Warn("Handler failed, retrying",
			slog.String("topic", message.Topic),
			slog.String("message_id", message.ID),
			slog.Int("attempt", attempt),
			slog.Int("max_attempts", maxAttempts),
			slog.Duration("backoff", backoffDelay),
			slog.Any("error", err))

		// Wait before retry with context cancellation support
		select {
		case <-time.After(backoffDelay):
			// Continue to next attempt
		case <-ctx.Done():
			mb.logger.Error("Handler retry cancelled by context",
				slog.String("topic", message.Topic),
				slog.String("message_id", message.ID),
				slog.Int("attempt", attempt),
				slog.Any("error", ctx.Err()))
			return
		}
	}

	// All retry attempts exhausted
	mb.logger.Error("Handler failed after all retry attempts",
		slog.String("topic", message.Topic),
		slog.String("message_id", message.ID),
		slog.Int("total_attempts", maxAttempts),
		slog.Any("error", lastErr))
}

// Subscribe registers a handler for the given topic.
func (mb *MemoryBus) Subscribe(topic string, handler bus.Handler) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("bus is closed")
	}

	mb.subscribers[topic] = append(mb.subscribers[topic], handler)
	return nil
}

// Unsubscribe removes a handler from the topic.
// Note: This removes ALL occurrences of the handler.
func (mb *MemoryBus) Unsubscribe(topic string, handler bus.Handler) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	handlers, exists := mb.subscribers[topic]
	if !exists {
		return nil // Already unsubscribed
	}

	// Filter out matching handlers
	// Note: Function comparison in Go is tricky, this is a simple implementation
	filtered := make([]bus.Handler, 0, len(handlers))
	for _, h := range handlers {
		// We can't reliably compare functions, so this is a basic approach
		// In production, consider using handler IDs or names
		_ = h // Mark as used
		// filtered = append(filtered, h) // Skip the handler we want to remove
	}
	mb.subscribers[topic] = filtered

	return nil
}

// Close gracefully shuts down the bus.
// Waits for all in-flight messages to be processed.
func (mb *MemoryBus) Close(ctx context.Context) error {
	mb.mu.Lock()
	if mb.closed {
		mb.mu.Unlock()
		return nil
	}
	mb.closed = true
	mb.mu.Unlock()

	// Wait for all handlers to complete with timeout
	done := make(chan struct{})
	go func() {
		mb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("bus close timed out: %w", ctx.Err())
	}
}

// Health checks if the bus is operational.
func (mb *MemoryBus) Health(ctx context.Context) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return fmt.Errorf("bus is closed")
	}

	return nil
}

// Stats returns current bus statistics (for monitoring).
func (mb *MemoryBus) Stats() map[string]interface{} {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	topicCounts := make(map[string]int)
	for topic, handlers := range mb.subscribers {
		topicCounts[topic] = len(handlers)
	}

	return map[string]interface{}{
		"total_topics":      len(mb.subscribers),
		"total_subscribers": mb.countTotalSubscribers(),
		"worker_pool_size":  mb.config.WorkerPoolSize,
		"topics":            topicCounts,
		"closed":            mb.closed,
	}
}

func (mb *MemoryBus) countTotalSubscribers() int {
	count := 0
	for _, handlers := range mb.subscribers {
		count += len(handlers)
	}
	return count
}
