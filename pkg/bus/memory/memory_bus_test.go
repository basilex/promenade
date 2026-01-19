package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryBus_PublishSubscribe(t *testing.T) {
	config := bus.NewConfig(5, 100, 1, time.Second, time.Second*5, 2.0)
	mb := memory.NewMemoryBus(config)
	defer func() { _ = mb.Close(context.Background()) }()

	received := make(chan bus.Event, 1)
	topic := "test.event"

	err := mb.Subscribe(topic, func(ctx context.Context, event bus.Event) error {
		received <- event
		return nil
	})
	require.NoError(t, err)

	testEvent := bus.NewBaseEvent(topic, uuidv7.New())
	err = mb.Publish(context.Background(), topic, testEvent)
	require.NoError(t, err)

	select {
	case event := <-received:
		assert.Equal(t, topic, event.Type())
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestMemoryBus_Health(t *testing.T) {
	config := bus.NewConfig(5, 100, 1, time.Second, time.Second*5, 2.0)
	mb := memory.NewMemoryBus(config)

	err := mb.Health(context.Background())
	assert.NoError(t, err)

	_ = mb.Close(context.Background())

	err = mb.Health(context.Background())
	assert.Error(t, err)
}
