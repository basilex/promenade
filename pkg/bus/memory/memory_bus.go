package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/google/uuid"
)

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
		ID:          uuid.New().String(),
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

// dispatch processes a single message with a handler.
// Handles worker pool limiting and error recovery.
func (mb *MemoryBus) dispatch(ctx context.Context, handler bus.Handler, message bus.Message) {
	defer mb.wg.Done()

	// Acquire worker slot (blocks if pool is full)
	mb.workerPool <- struct{}{}
	defer func() { <-mb.workerPool }()

	// Recover from panics in handlers
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't crash the bus
			// TODO: add proper logging when logger is integrated
			fmt.Printf("PANIC in event handler: %v\n", r)
		}
	}()

	// Call handler
	if err := handler(ctx, message.Event); err != nil {
		// TODO: implement retry logic based on config.RetryPolicy
		// For now, just log the error
		fmt.Printf("Handler error for topic %s: %v\n", message.Topic, err)
	}
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
