package bus_test

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/bus"
	_ "github.com/basilex/promenade/pkg/bus/memory" // Register memory adapter
	_ "github.com/basilex/promenade/pkg/bus/redis"  // Register redis adapter
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestBus_MemoryAdapter_EndToEnd tests complete event flow with memory adapter
func TestBus_MemoryAdapter_EndToEnd(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 10,
		BufferSize:     1000,
		RetryAttempts:  3,
		RetryDelay:     100 * time.Millisecond,
	}

	b, err := bus.NewBus(cfg)
	require.NoError(t, err)
	defer b.Close(context.Background())

	// Multiple topics and handlers
	userEvents := make(chan bus.Event, 10)
	contactEvents := make(chan bus.Event, 10)

	userHandler := func(ctx context.Context, e bus.Event) error {
		userEvents <- e
		return nil
	}

	contactHandler := func(ctx context.Context, e bus.Event) error {
		contactEvents <- e
		return nil
	}

	// Subscribe to multiple topics
	require.NoError(t, b.Subscribe("user.created", userHandler))
	require.NoError(t, b.Subscribe("contact.created", contactHandler))

	// Publish events to different topics
	userEvent := bus.NewBaseEvent("user.created", uuidv7.New())
	contactEvent := bus.NewBaseEvent("contact.created", uuidv7.New())

	require.NoError(t, b.Publish(context.Background(), "user.created", userEvent))
	require.NoError(t, b.Publish(context.Background(), "contact.created", contactEvent))

	// Verify events received
	select {
	case e := <-userEvents:
		assert.Equal(t, "user.created", e.Type())
	case <-time.After(2 * time.Second):
		t.Fatal("User event not received")
	}

	select {
	case e := <-contactEvents:
		assert.Equal(t, "contact.created", e.Type())
	case <-time.After(2 * time.Second):
		t.Fatal("Contact event not received")
	}
}

// TestBus_MemoryAdapter_MultipleSubscribers tests fan-out pattern
func TestBus_MemoryAdapter_MultipleSubscribers(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 10,
		BufferSize:     1000,
	}

	b, err := bus.NewBus(cfg)
	require.NoError(t, err)
	defer b.Close(context.Background())

	topic := "integration.fanout"
	var count atomic.Int32

	// 5 subscribers
	for i := 0; i < 5; i++ {
		handler := func(ctx context.Context, e bus.Event) error {
			count.Add(1)
			return nil
		}
		require.NoError(t, b.Subscribe(topic, handler))
	}

	// Publish one event
	event := bus.NewBaseEvent("test.event", uuidv7.New())
	require.NoError(t, b.Publish(context.Background(), topic, event))

	// Wait for all handlers
	time.Sleep(time.Second)

	// All 5 subscribers should receive event
	assert.Equal(t, int32(5), count.Load())
}

// TestBus_MemoryAdapter_HighThroughput tests performance with many events
func TestBus_MemoryAdapter_HighThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high throughput test in short mode")
	}

	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 20,
		BufferSize:     10000,
	}

	b, err := bus.NewBus(cfg)
	require.NoError(t, err)
	defer b.Close(context.Background())

	topic := "integration.throughput"
	var receivedCount atomic.Int32

	handler := func(ctx context.Context, e bus.Event) error {
		receivedCount.Add(1)
		return nil
	}

	require.NoError(t, b.Subscribe(topic, handler))

	// Publish 1000 events concurrently
	numEvents := 1000
	var wg sync.WaitGroup
	wg.Add(numEvents)

	start := time.Now()
	for i := 0; i < numEvents; i++ {
		go func(n int) {
			defer wg.Done()
			event := bus.NewBaseEvent("throughput.event", uuidv7.New())
			b.Publish(context.Background(), topic, event)
		}(i)
	}

	wg.Wait()
	publishDuration := time.Since(start)

	// Wait for processing
	time.Sleep(3 * time.Second)

	received := receivedCount.Load()
	assert.Equal(t, int32(numEvents), received, "Not all events received")

	t.Logf("Published %d events in %v (%d events/sec)",
		numEvents, publishDuration, int(float64(numEvents)/publishDuration.Seconds()))
}

// TestBus_MemoryAdapter_RetryPolicy tests retry behavior
func TestBus_MemoryAdapter_RetryPolicy(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
		RetryAttempts:  3,
		RetryDelay:     50 * time.Millisecond,
	}

	b, err := bus.NewBus(cfg)
	require.NoError(t, err)
	defer b.Close(context.Background())

	topic := "integration.retry"
	var attempts atomic.Int32
	success := make(chan struct{})

	handler := func(ctx context.Context, e bus.Event) error {
		count := attempts.Add(1)
		if count < 3 {
			return assert.AnError // Fail first 2 attempts
		}
		close(success)
		return nil // Succeed on 3rd attempt
	}

	require.NoError(t, b.Subscribe(topic, handler))

	event := bus.NewBaseEvent("retry.event", uuidv7.New())
	require.NoError(t, b.Publish(context.Background(), topic, event))

	select {
	case <-success:
		assert.Equal(t, int32(3), attempts.Load())
	case <-time.After(5 * time.Second):
		t.Fatalf("Handler did not succeed, attempts: %d", attempts.Load())
	}
}

// TestBus_MemoryAdapter_GracefulShutdown tests clean shutdown
func TestBus_MemoryAdapter_GracefulShutdown(t *testing.T) {
	cfg := config.BusSection{
		Adapter:        "memory",
		WorkerPoolSize: 5,
		BufferSize:     100,
	}

	b, err := bus.NewBus(cfg)
	require.NoError(t, err)

	topic := "integration.shutdown"
	processing := make(chan struct{})
	done := make(chan struct{})

	handler := func(ctx context.Context, e bus.Event) error {
		close(processing)
		time.Sleep(500 * time.Millisecond) // Simulate work
		close(done)
		return nil
	}

	require.NoError(t, b.Subscribe(topic, handler))

	event := bus.NewBaseEvent("shutdown.event", uuidv7.New())
	require.NoError(t, b.Publish(context.Background(), topic, event))

	// Wait for handler to start
	<-processing

	// Close with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = b.Close(ctx)
	assert.NoError(t, err)

	// Handler should complete
	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Handler did not complete during graceful shutdown")
	}
}

// TestBus_RedisAdapter_CrossProcess tests distributed events (if Redis available)
func TestBus_RedisAdapter_CrossProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis test in short mode")
	}

	// Use test Redis port (6380) to match docker-compose.test.yml
	redisPort := 6380
	if os.Getenv("ENVIRONMENT") == "test" {
		redisPort = 6380
	}

	cfg := config.BusSection{
		Adapter:        "redis",
		WorkerPoolSize: 10,
		BufferSize:     1000,
		Redis: config.RedisSection{
			Host:     "localhost",
			Port:     redisPort,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
	}

	// Try to create first instance
	b1, err := bus.NewBus(cfg)
	if err != nil {
		t.Skipf("Redis not available on port %d: %v", redisPort, err)
	}

	// Check if we actually got Redis adapter (not fallback to memory)
	if b1 == nil {
		t.Skip("Redis fallback to memory, skipping cross-process test")
	}

	defer b1.Close(context.Background())

	// Create second instance (simulating another process)
	b2, err := bus.NewBus(cfg)
	require.NoError(t, err)
	defer b2.Close(context.Background())

	topic := "integration.redis.distributed"
	received1 := make(chan bus.Event, 1)
	received2 := make(chan bus.Event, 1)

	handler1 := func(ctx context.Context, e bus.Event) error {
		received1 <- e
		return nil
	}

	handler2 := func(ctx context.Context, e bus.Event) error {
		received2 <- e
		return nil
	}

	require.NoError(t, b1.Subscribe(topic, handler1))
	require.NoError(t, b2.Subscribe(topic, handler2))

	time.Sleep(500 * time.Millisecond) // Wait for subscriptions

	// Publish from first instance
	event := bus.NewBaseEvent("distributed.event", uuidv7.New())
	require.NoError(t, b1.Publish(context.Background(), topic, event))

	// Both instances should receive
	select {
	case e := <-received1:
		assert.Equal(t, "distributed.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Instance 1 did not receive event")
	}

	select {
	case e := <-received2:
		assert.Equal(t, "distributed.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Instance 2 did not receive event")
	}
}

// TestBus_FactorySelection tests adapter factory selection
func TestBus_FactorySelection(t *testing.T) {
	tests := []struct {
		name    string
		adapter string
		wantErr bool
	}{
		{"Memory adapter", "memory", false},
		{"Invalid adapter", "invalid", true},
		{"Empty adapter", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.BusSection{
				Adapter:        tt.adapter,
				WorkerPoolSize: 5,
				BufferSize:     100,
			}

			b, err := bus.NewBus(cfg)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, b)
			} else {
				require.NoError(t, err)
				require.NotNil(t, b)
				b.Close(context.Background())
			}
		})
	}
}
