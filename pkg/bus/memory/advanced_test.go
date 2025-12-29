package memory_test

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
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestMemoryBus_MultipleSubscribers(t *testing.T) {
	config := bus.NewConfig(5, 100, 1, 100*time.Millisecond, time.Second, 2.0)
	mb := memory.NewMemoryBus(config)
	defer func() { _ = mb.Close(context.Background()) }()

	topic := "test.multiple"
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

	require.NoError(t, mb.Subscribe(topic, handler1))
	require.NoError(t, mb.Subscribe(topic, handler2))
	require.NoError(t, mb.Subscribe(topic, handler3))

	event := bus.NewBaseEvent("test.event", uuidv7.New())
	require.NoError(t, _ = mb.Publish(context.Background(), topic, event))

	select {
	case e := <-received1:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(2 * time.Second):
		t.Fatal("Handler 1 timeout")
	}

	select {
	case e := <-received2:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(2 * time.Second):
		t.Fatal("Handler 2 timeout")
	}

	select {
	case e := <-received3:
		assert.Equal(t, "test.event", e.Type())
	case <-time.After(2 * time.Second):
		t.Fatal("Handler 3 timeout")
	}
}

func TestMemoryBus_RetryLogic(t *testing.T) {
	config := bus.NewConfig(5, 100, 3, 10*time.Millisecond, 100*time.Millisecond, 1.5)
	mb := memory.NewMemoryBus(config)
	defer func() { _ = mb.Close(context.Background()) }()

	topic := "test.retry"
	var attempts atomic.Int32
	testErr := errors.New("temporary error")

	handler := func(ctx context.Context, e bus.Event) error {
		count := attempts.Add(1)
		if count < 3 {
			return testErr
		}
		return nil
	}

	require.NoError(t, mb.Subscribe(topic, handler))

	event := bus.NewBaseEvent("test.retry.event", uuidv7.New())
	require.NoError(t, _ = mb.Publish(context.Background(), topic, event))

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestMemoryBus_PanicRecovery(t *testing.T) {
	config := bus.NewConfig(5, 100, 1, 100*time.Millisecond, time.Second, 2.0)
	mb := memory.NewMemoryBus(config)
	defer func() { _ = mb.Close(context.Background()) }()

	topic := "test.panic"

	panicHandler := func(ctx context.Context, e bus.Event) error {
		panic("test panic!")
	}

	received := make(chan bus.Event, 1)
	normalHandler := func(ctx context.Context, e bus.Event) error {
		received <- e
		return nil
	}

	require.NoError(t, mb.Subscribe(topic, panicHandler))
	require.NoError(t, mb.Subscribe(topic, normalHandler))

	event := bus.NewBaseEvent("test.panic.event", uuidv7.New())
	require.NoError(t, _ = mb.Publish(context.Background(), topic, event))

	select {
	case e := <-received:
		assert.Equal(t, "test.panic.event", e.Type())
	case <-time.After(2 * time.Second):
		t.Fatal("Normal handler timeout after panic")
	}
}

func TestMemoryBus_Unsubscribe(t *testing.T) {
	config := bus.NewConfig(5, 100, 1, 100*time.Millisecond, time.Second, 2.0)
	mb := memory.NewMemoryBus(config)
	defer func() { _ = mb.Close(context.Background()) }()

	topic := "test.unsubscribe"
	received := make(chan bus.Event, 10)
	handler := func(ctx context.Context, e bus.Event) error {
		received <- e
		return nil
	}

	require.NoError(t, mb.Subscribe(topic, handler))

	event1 := bus.NewBaseEvent("test.event.1", uuidv7.New())
	require.NoError(t, mb.Publish(context.Background(), topic, event1))

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("Did not receive first event")
	}

	require.NoError(t, mb.Unsubscribe(topic, handler))

	event2 := bus.NewBaseEvent("test.event.2", uuidv7.New())
	require.NoError(t, mb.Publish(context.Background(), topic, event2))

	select {
	case <-received:
		t.Fatal("Received event after unsubscribe")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestMemoryBus_ConcurrentPublish(t *testing.T) {
	config := bus.NewConfig(10, 1000, 1, 100*time.Millisecond, time.Second, 2.0)
	mb := memory.NewMemoryBus(config)
	defer func() { _ = mb.Close(context.Background()) }()

	topic := "test.concurrent"
	var receivedCount atomic.Int32
	handler := func(ctx context.Context, e bus.Event) error {
		receivedCount.Add(1)
		return nil
	}

	require.NoError(t, mb.Subscribe(topic, handler))

	numEvents := 100
	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(n int) {
			defer wg.Done()
			event := bus.NewBaseEvent("test.concurrent.event", uuidv7.New())
			_ = mb.Publish(context.Background(), topic, event)
		}(i)
	}

	wg.Wait()
	time.Sleep(time.Second)

	assert.Equal(t, int32(numEvents), receivedCount.Load())
}

func TestMemoryBus_PublishAfterClose(t *testing.T) {
	config := bus.NewConfig(5, 100, 1, 100*time.Millisecond, time.Second, 2.0)
	mb := memory.NewMemoryBus(config)

	require.NoError(t, mb.Close(context.Background()))

	event := bus.NewBaseEvent("test.event", uuidv7.New())
	err := mb.Publish(context.Background(), "test.closed", event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}
