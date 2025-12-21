package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/basilex/promenade/pkg/uuidv7"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
)

// init registers the Redis bus factory
func init() {
	bus.RedisBusFactory = func(cfg config.BusConfig, busConfig bus.BusConfig) (bus.Bus, error) {
		return NewRedisBus(
			cfg.Redis.Host,
			cfg.Redis.Port,
			cfg.Redis.Password,
			cfg.Redis.DB,
			cfg.Redis.PoolSize,
			busConfig,
		)
	}
}

// RedisBus implements the Bus interface using Redis Pub/Sub
type RedisBus struct {
	client     *redis.Client
	config     bus.BusConfig
	handlers   map[string][]bus.Handler
	handlersMu sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	closed     bool
	closeMu    sync.Mutex
}

type eventEnvelope struct {
	Topic       string          `json:"topic"`
	EventType   string          `json:"event_type"`
	AggregateID uuidv7.UUID     `json:"aggregate_id"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Payload     json.RawMessage `json:"payload"`
}

// NewRedisBus creates a new Redis-based event bus
func NewRedisBus(host string, port int, password string, db int, poolSize int, config bus.BusConfig) (*RedisBus, error) {
	// Determine max retries from config (default: 3 for production reliability)
	maxRetries := 3
	if config.RetryPolicy != nil && config.RetryPolicy.MaxAttempts > 0 {
		maxRetries = config.RetryPolicy.MaxAttempts
	}

	// Production-ready timeouts (balance between responsiveness and reliability)
	dialTimeout := 5 * time.Second  // Allow time for network/DNS resolution
	readTimeout := 3 * time.Second  // Allow time for Redis to respond
	writeTimeout := 3 * time.Second // Allow time for Redis to write

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: poolSize / 2,
		MaxRetries:   maxRetries,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	// Allow sufficient time for initial connection in production
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	busCtx, busCancel := context.WithCancel(context.Background())

	rb := &RedisBus{
		client:   client,
		config:   config,
		handlers: make(map[string][]bus.Handler),
		ctx:      busCtx,
		cancel:   busCancel,
	}

	logger.Info("Redis bus initialized",
		slog.String("host", host),
		slog.Int("port", port),
		slog.Int("db", db))

	return rb, nil
}

// Publish sends an event via Redis Pub/Sub
func (rb *RedisBus) Publish(ctx context.Context, topic string, event bus.Event) error {
	rb.closeMu.Lock()
	if rb.closed {
		rb.closeMu.Unlock()
		return fmt.Errorf("bus is closed")
	}
	rb.closeMu.Unlock()

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	envelope := eventEnvelope{
		Topic:       topic,
		EventType:   event.Type(),
		AggregateID: event.AggregateID(),
		OccurredAt:  event.OccurredAt(),
		Payload:     payload,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	if err := rb.client.Publish(ctx, topic, data).Err(); err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	return nil
}

// Subscribe registers a handler for messages
func (rb *RedisBus) Subscribe(topic string, handler bus.Handler) error {
	rb.closeMu.Lock()
	if rb.closed {
		rb.closeMu.Unlock()
		return fmt.Errorf("bus is closed")
	}
	rb.closeMu.Unlock()

	rb.handlersMu.Lock()
	defer rb.handlersMu.Unlock()

	isFirstHandler := len(rb.handlers[topic]) == 0
	rb.handlers[topic] = append(rb.handlers[topic], handler)

	if isFirstHandler {
		rb.wg.Add(1)
		go rb.subscriptionWorker(topic)
	}

	return nil
}

// Unsubscribe removes a handler
func (rb *RedisBus) Unsubscribe(topic string, handler bus.Handler) error {
	rb.handlersMu.Lock()
	defer rb.handlersMu.Unlock()

	handlers, exists := rb.handlers[topic]
	if !exists {
		return fmt.Errorf("no handlers for topic: %s", topic)
	}

	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			rb.handlers[topic] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// subscriptionWorker listens to Redis Pub/Sub
func (rb *RedisBus) subscriptionWorker(topic string) {
	defer rb.wg.Done()

	pubsub := rb.client.Subscribe(rb.ctx, topic)
	defer pubsub.Close()

	logger.Info("Started subscription worker", slog.String("topic", topic))

	ch := pubsub.Channel()

	for {
		select {
		case <-rb.ctx.Done():
			logger.Info("Subscription worker stopped", slog.String("topic", topic))
			return

		case msg, ok := <-ch:
			if !ok {
				logger.Warn("Redis channel closed", slog.String("topic", topic))
				return
			}

			var envelope eventEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
				logger.Error("Failed to unmarshal event envelope",
					slog.String("topic", topic),
					slog.Any("error", err))
				continue
			}

			rb.handlersMu.RLock()
			handlers := make([]bus.Handler, len(rb.handlers[topic]))
			copy(handlers, rb.handlers[topic])
			rb.handlersMu.RUnlock()

			for _, handler := range handlers {
				rb.wg.Add(1)
				go rb.dispatchEvent(handler, &envelope)
			}
		}
	}
}

// dispatchEvent calls a handler with retry logic
func (rb *RedisBus) dispatchEvent(handler bus.Handler, envelope *eventEnvelope) {
	defer rb.wg.Done()

	genericEvent := &bus.BaseEvent{
		EventType:    envelope.EventType,
		EventID:      uuidv7.New(),
		AggregateId:  envelope.AggregateID,
		OccurredTime: envelope.OccurredAt,
	}

	// Get retry parameters
	maxAttempts := 3
	initialDelay := 1 * time.Second
	maxDelay := 30 * time.Second
	multiplier := 2.0

	if rb.config.RetryPolicy != nil {
		if rb.config.RetryPolicy.MaxAttempts > 0 {
			maxAttempts = rb.config.RetryPolicy.MaxAttempts
		}
		if rb.config.RetryPolicy.InitialDelay > 0 {
			initialDelay = rb.config.RetryPolicy.InitialDelay
		}
		if rb.config.RetryPolicy.MaxDelay > 0 {
			maxDelay = rb.config.RetryPolicy.MaxDelay
		}
		if rb.config.RetryPolicy.Multiplier > 0 {
			multiplier = rb.config.RetryPolicy.Multiplier
		}
	}

	var err error
	for attempt := 0; attempt <= maxAttempts; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(float64(initialDelay) * float64(attempt) * multiplier)
			if delay > maxDelay {
				delay = maxDelay
			}
			time.Sleep(delay)
		}

		err = handler(rb.ctx, genericEvent)
		if err == nil {
			return
		}

		logger.Error("Event handler failed",
			slog.String("event_type", envelope.EventType),
			slog.Int("attempt", attempt),
			slog.Any("error", err))
	}

	logger.Error("Event handler failed after all retries",
		slog.String("event_type", envelope.EventType),
		slog.Any("error", err))
}

// Close gracefully shuts down the bus
func (rb *RedisBus) Close(ctx context.Context) error {
	rb.closeMu.Lock()
	if rb.closed {
		rb.closeMu.Unlock()
		return nil
	}
	rb.closed = true
	rb.closeMu.Unlock()

	logger.Info("Closing Redis bus...")

	rb.cancel()

	done := make(chan struct{})
	go func() {
		rb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All Redis bus workers stopped")
	case <-ctx.Done():
		logger.Warn("Redis bus shutdown timeout")
	}

	if err := rb.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	return nil
}

// Health checks if the bus is healthy
func (rb *RedisBus) Health(ctx context.Context) error {
	rb.closeMu.Lock()
	if rb.closed {
		rb.closeMu.Unlock()
		return fmt.Errorf("bus is closed")
	}
	rb.closeMu.Unlock()

	if err := rb.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}
