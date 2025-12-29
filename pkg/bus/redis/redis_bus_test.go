package redis_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/redis"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// skipIfRedisUnavailable skips test if Redis is not available
func skipIfRedisUnavailable(t *testing.T) *redis.RedisBus {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	config := bus.NewConfig(5, 100, 2, 100*time.Millisecond, time.Second, 2.0)
	rb, err := redis.NewRedisBus("localhost:6379", "", 0, 10, config)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	return rb
}

func TestRedisBus_NewRedisBus_Success(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	assert.NotNil(t, rb)
}

func TestRedisBus_NewRedisBus_InvalidHost(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	config := bus.NewConfig(5, 100, 1, 100*time.Millisecond, time.Second, 2.0)
	rb, err := redis.NewRedisBus("invalid-host-12345:6379", "", 0, 10, config)

	assert.Error(t, err)
	assert.Nil(t, rb)
	assert.Contains(t, err.Error(), "failed to ping Redis")
}

func TestRedisBus_NewRedisBus_InvalidPort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	config := bus.NewConfig(5, 100, 1, 100*time.Millisecond, time.Second, 2.0)
	rb, err := redis.NewRedisBus("localhost:9999", "", 0, 10, config)

	assert.Error(t, err)
	assert.Nil(t, rb)
}

func TestRedisBus_Health(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	err := rb.Health(context.Background())
	assert.NoError(t, err)
}

func TestRedisBus_Health_AfterClose(t *testing.T) {
	rb := skipIfRedisUnavailable(t)

	err := rb.Close(context.Background())
	require.NoError(t, err)

	err = rb.Health(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestRedisBus_PublishSubscribe(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	topic := "test.redis.pubsub"
	received := make(chan bus.Event, 1)

	handler := func(ctx context.Context, e bus.Event) error {
		received <- e
		return nil
	}

	err := rb.Subscribe(topic, handler)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	event := bus.NewBaseEvent("test.event", uuidv7.New())
	err = _ = rb.Publish(context.Background(), topic, event)
	require.NoError(t, err)

	select {
	case e := <-received:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Did not receive event")
	}
}

func TestRedisBus_MultipleSubscribers(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	topic := "test.redis.multiple"
	received1 := make(chan bus.Event, 1)
	received2 := make(chan bus.Event, 1)
	received3 := make(chan bus.Event, 1)

	handler1 := func(ctx context.Context, e bus.Event) error {
		received1 <- e
		return nil
	}
	handler2 := func(ctx context.Context, e bus.Event) error {
		received2 <- e
		return nil
	}
	handler3 := func(ctx context.Context, e bus.Event) error {
		received3 <- e
		return nil
	}

	require.NoError(t, rb.Subscribe(topic, handler1))
	require.NoError(t, rb.Subscribe(topic, handler2))
	require.NoError(t, rb.Subscribe(topic, handler3))

	time.Sleep(300 * time.Millisecond)

	event := bus.NewBaseEvent("test.event", uuidv7.New())
	err := _ = rb.Publish(context.Background(), topic, event)
	require.NoError(t, err)

	select {
	case e := <-received1:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Handler 1 timeout")
	}

	select {
	case e := <-received2:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Handler 2 timeout")
	}

	select {
	case e := <-received3:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Handler 3 timeout")
	}
}

func TestRedisBus_RetryLogic(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	topic := "test.redis.retry"
	var attempts atomic.Int32
	testErr := errors.New("temporary error")
	done := make(chan struct{})

	handler := func(ctx context.Context, e bus.Event) error {
		count := attempts.Add(1)
		if count < 2 {
			return testErr
		}
		close(done)
		return nil
	}

	require.NoError(t, rb.Subscribe(topic, handler))
	time.Sleep(200 * time.Millisecond)

	event := bus.NewBaseEvent("test.retry.event", uuidv7.New())
	err := _ = rb.Publish(context.Background(), topic, event)
	require.NoError(t, err)

	select {
	case <-done:
		assert.GreaterOrEqual(t, attempts.Load(), int32(2))
	case <-time.After(5 * time.Second):
		t.Fatalf("Handler did not succeed, attempts: %d", attempts.Load())
	}
}

func TestRedisBus_PanicRecovery(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	topic := "test.redis.panic"

	panicHandler := func(ctx context.Context, e bus.Event) error {
		panic("test panic!")
	}

	received := make(chan bus.Event, 1)
	normalHandler := func(ctx context.Context, e bus.Event) error {
		received <- e
		return nil
	}

	require.NoError(t, rb.Subscribe(topic, panicHandler))
	require.NoError(t, rb.Subscribe(topic, normalHandler))

	time.Sleep(200 * time.Millisecond)

	event := bus.NewBaseEvent("test.panic.event", uuidv7.New())
	err := _ = rb.Publish(context.Background(), topic, event)
	require.NoError(t, err)

	select {
	case e := <-received:
		assert.Equal(t, "test.panic.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Normal handler did not receive event after panic")
	}
}

func TestRedisBus_Unsubscribe(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	topic := "test.redis.unsubscribe"
	received := make(chan bus.Event, 10)

	handler := func(ctx context.Context, e bus.Event) error {
		received <- e
		return nil
	}

	require.NoError(t, rb.Subscribe(topic, handler))
	time.Sleep(200 * time.Millisecond)

	event1 := bus.NewBaseEvent("test.event.1", uuidv7.New())
	err := rb.Publish(context.Background(), topic, event1)
	require.NoError(t, err)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("Did not receive first event")
	}

	err = rb.Unsubscribe(topic, handler)
	require.NoError(t, err)

	event2 := bus.NewBaseEvent("test.event.2", uuidv7.New())
	err = rb.Publish(context.Background(), topic, event2)
	require.NoError(t, err)

	select {
	case <-received:
		t.Fatal("Received event after unsubscribe")
	case <-time.After(500 * time.Millisecond):
	}
}

func TestRedisBus_MultipleInstances(t *testing.T) {
	rb1 := skipIfRedisUnavailable(t)
	defer func() { _ = rb1.Close(context.Background()) }()

	config := bus.NewConfig(5, 100, 2, 100*time.Millisecond, time.Second, 2.0)
	rb2, err := redis.NewRedisBus("localhost:6379", "", 0, 10, config)
	require.NoError(t, err)
	defer func() { _ = rb2.Close(context.Background()) }()

	topic := "test.redis.distributed"
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

	require.NoError(t, rb1.Subscribe(topic, handler1))
	require.NoError(t, rb2.Subscribe(topic, handler2))

	time.Sleep(300 * time.Millisecond)

	event := bus.NewBaseEvent("test.distributed.event", uuidv7.New())
	err = rb1.Publish(context.Background(), topic, event)
	require.NoError(t, err)

	select {
	case e := <-received1:
		assert.Equal(t, "test.distributed.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Instance 1 timeout")
	}

	select {
	case e := <-received2:
		assert.Equal(t, "test.distributed.event", e.Type())
	case <-time.After(3 * time.Second):
		t.Fatal("Instance 2 timeout")
	}
}

func TestRedisBus_ConcurrentPublish(t *testing.T) {
	rb := skipIfRedisUnavailable(t)
	defer func() { _ = rb.Close(context.Background()) }()

	topic := "test.redis.concurrent"
	var receivedCount atomic.Int32

	handler := func(ctx context.Context, e bus.Event) error {
		receivedCount.Add(1)
		return nil
	}

	require.NoError(t, rb.Subscribe(topic, handler))
	time.Sleep(200 * time.Millisecond)

	numEvents := 50
	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(n int) {
			defer wg.Done()
			event := bus.NewBaseEvent("test.concurrent.event", uuidv7.New())
			_ = rb.Publish(context.Background(), topic, event)
		}(i)
	}

	wg.Wait()
	time.Sleep(2 * time.Second)

	assert.Equal(t, int32(numEvents), receivedCount.Load())
}

func TestRedisBus_PublishAfterClose(t *testing.T) {
	rb := skipIfRedisUnavailable(t)

	err := rb.Close(context.Background())
	require.NoError(t, err)

	event := bus.NewBaseEvent("test.event", uuidv7.New())
	err = rb.Publish(context.Background(), "test.closed", event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestRedisBus_SubscribeAfterClose(t *testing.T) {
	rb := skipIfRedisUnavailable(t)

	err := rb.Close(context.Background())
	require.NoError(t, err)

	handler := func(ctx context.Context, e bus.Event) error {
		return nil
	}

	err = rb.Subscribe("test.closed", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestRedisBus_GracefulShutdown(t *testing.T) {
	rb := skipIfRedisUnavailable(t)

	topic := "test.redis.shutdown"
	processing := make(chan struct{})
	done := make(chan struct{})

	handler := func(ctx context.Context, e bus.Event) error {
		close(processing)
		time.Sleep(200 * time.Millisecond)
		close(done)
		return nil
	}

	require.NoError(t, rb.Subscribe(topic, handler))
	time.Sleep(200 * time.Millisecond)

	event := bus.NewBaseEvent("test.shutdown.event", uuidv7.New())
	err := _ = rb.Publish(context.Background(), topic, event)
	require.NoError(t, err)

	<-processing

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = rb.Close(ctx)
	assert.NoError(t, err)

	select {
	case <-done:
	case <-time.After(4 * time.Second):
		t.Fatal("Handler did not complete")
	}
}
