package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type testEvent struct {
	*bus.BaseEvent
	Message string
}

func newTestEvent(message string) *testEvent {
	return &testEvent{
		BaseEvent: bus.NewBaseEvent("test.event", uuidv7.New()),
		Message:   message,
	}
}

func TestMemoryBus_PublishSubscribe(t *testing.T) {
	mb := NewDefaultMemoryBus()
	defer mb.Close(context.Background())

	received := make(chan string, 1)

	// Subscribe to topic
	handler := func(ctx context.Context, event bus.Event) error {
		if te, ok := event.(*testEvent); ok {
			received <- te.Message
		}
		return nil
	}

	err := mb.Subscribe("test.topic", handler)
	require.NoError(t, err)

	// Publish event
	event := newTestEvent("hello world")
	err = mb.Publish(context.Background(), "test.topic", event)
	require.NoError(t, err)

	// Wait for message
	select {
	case msg := <-received:
		assert.Equal(t, "hello world", msg)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestMemoryBus_MultipleSubscribers(t *testing.T) {
	mb := NewDefaultMemoryBus()
	defer mb.Close(context.Background())

	received1 := make(chan string, 1)
	received2 := make(chan string, 1)

	handler1 := func(ctx context.Context, event bus.Event) error {
		if te, ok := event.(*testEvent); ok {
			received1 <- "handler1:" + te.Message
		}
		return nil
	}

	handler2 := func(ctx context.Context, event bus.Event) error {
		if te, ok := event.(*testEvent); ok {
			received2 <- "handler2:" + te.Message
		}
		return nil
	}

	// Subscribe both handlers
	require.NoError(t, mb.Subscribe("test.topic", handler1))
	require.NoError(t, mb.Subscribe("test.topic", handler2))

	// Publish event
	event := newTestEvent("broadcast")
	require.NoError(t, mb.Publish(context.Background(), "test.topic", event))

	// Both handlers should receive the message
	select {
	case msg := <-received1:
		assert.Equal(t, "handler1:broadcast", msg)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for handler1")
	}

	select {
	case msg := <-received2:
		assert.Equal(t, "handler2:broadcast", msg)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for handler2")
	}
}

func TestMemoryBus_NoSubscribers(t *testing.T) {
	mb := NewDefaultMemoryBus()
	defer mb.Close(context.Background())

	// Publishing to non-existent topic should not error
	event := newTestEvent("no one listening")
	err := mb.Publish(context.Background(), "empty.topic", event)
	assert.NoError(t, err)
}

func TestMemoryBus_Close(t *testing.T) {
	mb := NewDefaultMemoryBus()

	received := make(chan string, 10)

	handler := func(ctx context.Context, event bus.Event) error {
		time.Sleep(100 * time.Millisecond) // Simulate slow processing
		if te, ok := event.(*testEvent); ok {
			received <- te.Message
		}
		return nil
	}

	require.NoError(t, mb.Subscribe("test.topic", handler))

	// Publish multiple events
	for i := 0; i < 5; i++ {
		event := newTestEvent("message")
		require.NoError(t, mb.Publish(context.Background(), "test.topic", event))
	}

	// Close should wait for all handlers to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := mb.Close(ctx)
	assert.NoError(t, err)

	// Verify all messages were processed
	assert.Equal(t, 5, len(received))
}

func TestMemoryBus_PublishAfterClose(t *testing.T) {
	mb := NewDefaultMemoryBus()
	require.NoError(t, mb.Close(context.Background()))

	// Publishing after close should error
	event := newTestEvent("too late")
	err := mb.Publish(context.Background(), "test.topic", event)
	assert.Error(t, err)
}

func TestMemoryBus_Health(t *testing.T) {
	mb := NewDefaultMemoryBus()

	// Should be healthy initially
	err := mb.Health(context.Background())
	assert.NoError(t, err)

	require.NoError(t, mb.Close(context.Background()))

	// Should be unhealthy after close
	err = mb.Health(context.Background())
	assert.Error(t, err)
}

func TestMemoryBus_Stats(t *testing.T) {
	mb := NewDefaultMemoryBus()
	defer mb.Close(context.Background())

	handler := func(ctx context.Context, event bus.Event) error { return nil }

	require.NoError(t, mb.Subscribe("topic1", handler))
	require.NoError(t, mb.Subscribe("topic1", handler)) // Same topic, different handler
	require.NoError(t, mb.Subscribe("topic2", handler))

	stats := mb.Stats()
	assert.Equal(t, 2, stats["total_topics"])
	assert.Equal(t, 3, stats["total_subscribers"])
}
