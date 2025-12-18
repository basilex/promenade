package bus

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Bus represents a message bus for publishing and subscribing to events.
// It provides an abstraction over different message transport implementations
// (Redis Pub/Sub, NATS, Kafka, in-memory, etc.).
//
// Design principles:
// - Transport-agnostic interface
// - Fire-and-forget publishing (no acks required for basic use)
// - Multiple subscribers per topic
// - Graceful shutdown support
type Bus interface {
	// Publish sends an event to the specified topic.
	// Returns immediately after queuing the message.
	Publish(ctx context.Context, topic string, event Event) error

	// Subscribe registers a handler for messages on the given topic.
	// The handler will be called asynchronously for each incoming message.
	// Multiple subscribers can listen to the same topic.
	Subscribe(topic string, handler Handler) error

	// Unsubscribe removes a handler from the given topic.
	Unsubscribe(topic string, handler Handler) error

	// Close gracefully shuts down the bus, waiting for in-flight messages.
	Close(ctx context.Context) error

	// Health returns the current health status of the bus.
	Health(ctx context.Context) error
}

// Event represents a domain event that can be published to the bus.
type Event interface {
	// Type returns the event type identifier (e.g., "user.registered", "post.published")
	Type() string

	// OccurredAt returns when the event occurred
	OccurredAt() time.Time

	// AggregateID returns the ID of the aggregate that generated this event
	AggregateID() uuid.UUID

	// Metadata returns additional event metadata (optional)
	Metadata() map[string]string
}

// Handler is a function that processes an event.
// It should be idempotent as events may be delivered more than once.
// Return an error to indicate processing failure (may trigger retry depending on bus implementation).
type Handler func(ctx context.Context, event Event) error

// Message wraps an event with transport metadata.
type Message struct {
	ID          string            // Unique message ID
	Topic       string            // Topic/channel name
	Event       Event             // The actual event
	PublishedAt time.Time         // When the message was published to the bus
	Metadata    map[string]string // Transport-specific metadata
}

// BusConfig holds common configuration for all bus implementations.
type BusConfig struct {
	// WorkerPoolSize defines how many concurrent workers process messages
	WorkerPoolSize int

	// BufferSize defines the internal message buffer size
	BufferSize int

	// EnableMetrics enables prometheus metrics collection
	EnableMetrics bool

	// RetryPolicy defines retry behavior on handler failures
	RetryPolicy *RetryPolicy
}

// RetryPolicy defines how to handle failed message processing.
type RetryPolicy struct {
	MaxAttempts  int           // Maximum number of retry attempts
	InitialDelay time.Duration // Initial delay before first retry
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Backoff multiplier
}

// NewBusConfig creates a bus configuration from environment variables.
// Parameters should be loaded from config package.
func NewBusConfig(workerPoolSize, bufferSize, retryAttempts int, retryDelay time.Duration) BusConfig {
	return BusConfig{
		WorkerPoolSize: workerPoolSize,
		BufferSize:     bufferSize,
		EnableMetrics:  true, // Always enable metrics
		RetryPolicy: &RetryPolicy{
			MaxAttempts:  retryAttempts,
			InitialDelay: retryDelay,
			MaxDelay:     5 * time.Second,
			Multiplier:   2.0,
		},
	}
}
