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

// init registers the memory bus factory with the bus package
func init() {
	bus.MemoryBusFactory = func(config bus.Config) bus.IBus {
		return NewMemoryBus(config)
	}
}

// MemoryBus is an in-memory implementation of bus.IBus.
// Useful for development, testing, and simple deployments.
// Events are processed in goroutines without any persistence.
type MemoryBus struct {
	mu          sync.RWMutex
	subscribers map[string][]bus.Handler // topic -> handlers
	config      bus.Config
	workerPool  chan struct{}
	wg          sync.WaitGroup
	closed      bool
	logger      *slog.Logger
}

// NewMemoryBus creates a new in-memory message bus.
func NewMemoryBus(config bus.Config) *MemoryBus {
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

// dispatch processes a single message with error recovery and retry logic.
func (mb *MemoryBus) dispatch(ctx context.Context, handler bus.Handler, message bus.Message) {
	defer mb.wg.Done()

	// Acquire worker slot (blocks if pool is full)
	mb.workerPool <- struct{}{}
	defer func() { <-mb.workerPool }()

	// Recover from panics in handlers
	defer func() {
		if r := recover(); r != nil {
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
			// Success
			if attempt > 1 {
				mb.logger.Info("Handler succeeded after retry",
					slog.String("topic", message.Topic),
					slog.Int("attempt", attempt))
			}
			return
		}

		lastErr = err

		// If this is the last attempt, don't wait
		if attempt == maxAttempts {
			break
		}

		// Calculate exponential backoff delay
		var backoffDelay time.Duration
		if mb.config.RetryPolicy != nil {
			initialDelay := mb.config.RetryPolicy.InitialDelay
			multiplier := mb.config.RetryPolicy.Multiplier
			if multiplier <= 0 {
				multiplier = 2.0
			}

			backoffDelay = time.Duration(float64(initialDelay) * math.Pow(multiplier, float64(attempt-1)))

			if mb.config.RetryPolicy.MaxDelay > 0 && backoffDelay > mb.config.RetryPolicy.MaxDelay {
				backoffDelay = mb.config.RetryPolicy.MaxDelay
			}
		} else {
			backoffDelay = time.Second * time.Duration(1<<uint(attempt-1))
		}

		mb.logger.Warn("Handler failed, retrying",
			slog.String("topic", message.Topic),
			slog.Int("attempt", attempt),
			slog.Duration("backoff", backoffDelay),
			slog.Any("error", err))

		// Wait before retry
		select {
		case <-time.After(backoffDelay):
			// Continue to next attempt
		case <-ctx.Done():
			mb.logger.Error("Handler retry cancelled by context",
				slog.String("topic", message.Topic),
				slog.Any("error", ctx.Err()))
			return
		}
	}

	// All retry attempts exhausted
	mb.logger.Error("Handler failed after all retry attempts",
		slog.String("topic", message.Topic),
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
func (mb *MemoryBus) Unsubscribe(topic string, handler bus.Handler) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	handlers, exists := mb.subscribers[topic]
	if !exists {
		return nil
	}

	// Note: Function comparison in Go is limited, this removes all handlers
	mb.subscribers[topic] = handlers[:0]
	return nil
}

// Close gracefully shuts down the bus.
func (mb *MemoryBus) Close(ctx context.Context) error {
	mb.mu.Lock()
	if mb.closed {
		mb.mu.Unlock()
		return nil
	}
	mb.closed = true
	mb.mu.Unlock()

	// Wait for all handlers to complete
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
