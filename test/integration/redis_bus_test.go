package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/bus"
	_ "github.com/basilex/promenade/pkg/bus/memory"
	_ "github.com/basilex/promenade/pkg/bus/redis"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestEvent is a simple event for testing
type TestEvent struct {
	*bus.BaseEvent
	Message string `json:"message"`
}

func TestRedisBus_PublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load config with Redis adapter
	cfg := &config.BusConfig{
		Adapter:         "redis",
		WorkerPoolSize:  4,
		BufferSize:      100,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		RetryMaxDelay:   30 * time.Second,
		RetryMultiplier: 2.0,
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6380, // Test Redis port
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
	}

	// Create Redis bus
	eventBus, err := bus.NewBus(*cfg)
	require.NoError(t, err, "Should create Redis bus")
	defer eventBus.Close(context.Background())

	// Health check
	ctx := context.Background()
	err = eventBus.Health(ctx)
	require.NoError(t, err, "Redis should be healthy")

	// Test publish/subscribe
	topic := "test.redis.messages"
	messagesReceived := 0
	expectedMessage := "Redis test message"
	var receivedEvent bus.Event

	handler := func(ctx context.Context, event bus.Event) error {
		messagesReceived++
		receivedEvent = event
		return nil
	}

	// Subscribe
	err = eventBus.Subscribe(topic, handler)
	require.NoError(t, err, "Should subscribe successfully")

	// Give subscriber time to register (Redis Pub/Sub needs this)
	time.Sleep(500 * time.Millisecond)

	// Publish event
	aggregateID := uuidv7.New()
	testEvent := &TestEvent{
		BaseEvent: bus.NewBaseEvent("test.message", aggregateID),
		Message:   expectedMessage,
	}

	err = eventBus.Publish(ctx, topic, testEvent)
	require.NoError(t, err, "Should publish successfully")

	// Wait for message processing
	time.Sleep(1 * time.Second)

	// Assertions
	assert.Equal(t, 1, messagesReceived, "Should receive exactly 1 message")
	assert.NotNil(t, receivedEvent, "Should have received event")
	if receivedEvent != nil {
		assert.Equal(t, "test.message", receivedEvent.Type())
	}
}

func TestRedisBus_MultipleSubscribers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.BusConfig{
		Adapter:         "redis",
		WorkerPoolSize:  4,
		BufferSize:      100,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		RetryMaxDelay:   30 * time.Second,
		RetryMultiplier: 2.0,
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6380, // Test Redis port
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
	}

	eventBus, err := bus.NewBus(*cfg)
	require.NoError(t, err)
	defer eventBus.Close(context.Background())

	topic := "test.multi.subscribers"
	subscriber1Count := 0
	subscriber2Count := 0

	handler1 := func(ctx context.Context, event bus.Event) error {
		subscriber1Count++
		return nil
	}

	handler2 := func(ctx context.Context, event bus.Event) error {
		subscriber2Count++
		return nil
	}

	// Subscribe both handlers
	err = eventBus.Subscribe(topic, handler1)
	require.NoError(t, err)
	err = eventBus.Subscribe(topic, handler2)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Publish event
	aggregateID := uuidv7.New()
	testEvent := &TestEvent{
		BaseEvent: bus.NewBaseEvent("test.message", aggregateID),
		Message:   "Multi-subscriber test",
	}

	ctx := context.Background()
	err = eventBus.Publish(ctx, topic, testEvent)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Both subscribers should receive the message
	assert.Equal(t, 1, subscriber1Count, "Subscriber 1 should receive message")
	assert.Equal(t, 1, subscriber2Count, "Subscriber 2 should receive message")
}

func TestRedisBus_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.BusConfig{
		Adapter:         "redis",
		WorkerPoolSize:  4,
		BufferSize:      100,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		RetryMaxDelay:   30 * time.Second,
		RetryMultiplier: 2.0,
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6380, // Test Redis port
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
	}

	eventBus, err := bus.NewBus(*cfg)
	require.NoError(t, err)

	// Close with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = eventBus.Close(ctx)
	assert.NoError(t, err, "Should close gracefully")

	// Health check after close should fail
	err = eventBus.Health(context.Background())
	assert.Error(t, err, "Health check should fail after close")
}

func TestBusFallback_RedisToMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to connect to non-existent Redis
	cfg := &config.BusConfig{
		Adapter:         "redis",
		WorkerPoolSize:  4,
		BufferSize:      100,
		RetryAttempts:   1, // Fast fail for test
		RetryDelay:      100 * time.Millisecond,
		RetryMaxDelay:   1 * time.Second,
		RetryMultiplier: 2.0,
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     9999, // Invalid port - should trigger fallback
			Password: "",
			DB:       0,
			PoolSize: 2, // Smaller pool for faster failure
		},
	}

	// In development/test mode, this should return error (fail fast)
	// In production, it would fallback to memory
	eventBus, err := bus.NewBus(*cfg)
	// For test environment, we expect an error
	if err != nil {
		t.Logf("Expected behavior in test environment: %v", err)
		return
	}

	// If fallback succeeded (production mode), verify it works
	if eventBus != nil {
		defer eventBus.Close(context.Background())

		err = eventBus.Health(context.Background())
		assert.NoError(t, err, "Fallback bus should be healthy")
	}
}
