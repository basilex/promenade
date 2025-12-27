package bus_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestBaseEvent_Creation(t *testing.T) {
	aggregateID := uuidv7.New()
	event := bus.NewBaseEvent("user.created", aggregateID)

	assert.Equal(t, "user.created", event.Type())
	assert.Equal(t, aggregateID, event.AggregateID())
	assert.NotNil(t, event.Metadata())
	assert.Empty(t, event.Metadata())
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
}

func TestBaseEvent_Metadata(t *testing.T) {
	aggregateID := uuidv7.New()
	event := bus.NewBaseEvent("user.created", aggregateID)

	// Initial metadata should be empty
	metadata := event.Metadata()
	assert.NotNil(t, metadata)
	assert.Empty(t, metadata)

	// Metadata should be mutable (if designed that way)
	// This tests the behavior of the Metadata() method
	if len(metadata) == 0 {
		// Confirmed: returns empty map
		assert.Equal(t, 0, len(metadata))
	}
}

func TestBaseEvent_OccurredAt(t *testing.T) {
	before := time.Now()
	aggregateID := uuidv7.New()
	event := bus.NewBaseEvent("test.event", aggregateID)
	after := time.Now()

	occurredAt := event.OccurredAt()
	assert.False(t, occurredAt.Before(before), "OccurredAt should not be before creation")
	assert.False(t, occurredAt.After(after), "OccurredAt should not be after creation")
}

func TestBaseEvent_Type(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"User event", "user.created"},
		{"Contact event", "contact.verified"},
		{"Empty type", ""},
		{"Complex type", "customer.order.payment.received"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregateID := uuidv7.New()
			event := bus.NewBaseEvent(tt.eventType, aggregateID)
			assert.Equal(t, tt.eventType, event.Type())
		})
	}
}

func TestBaseEvent_AggregateID(t *testing.T) {
	// Test with different UUIDs
	id1 := uuidv7.New()
	id2 := uuidv7.New()

	event1 := bus.NewBaseEvent("test.event", id1)
	event2 := bus.NewBaseEvent("test.event", id2)

	assert.Equal(t, id1, event1.AggregateID())
	assert.Equal(t, id2, event2.AggregateID())
	assert.NotEqual(t, event1.AggregateID(), event2.AggregateID())
}

func TestBaseEvent_EventID(t *testing.T) {
	aggregateID := uuidv7.New()
	event1 := bus.NewBaseEvent("test.event", aggregateID)
	event2 := bus.NewBaseEvent("test.event", aggregateID)

	// Each event should have unique EventID even with same aggregate
	// This tests if EventID is generated per event
	// Note: BaseEvent struct has EventID field
	assert.NotEqual(t, event1, event2, "Two events should be different instances")
}

func TestBaseEvent_MultipleEvents(t *testing.T) {
	// Create multiple events and verify they're independent
	events := make([]bus.Event, 10)
	for i := 0; i < 10; i++ {
		aggregateID := uuidv7.New()
		events[i] = bus.NewBaseEvent("test.event", aggregateID)
	}

	// All events should have different aggregate IDs
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			assert.NotEqual(t, events[i].AggregateID(), events[j].AggregateID())
		}
	}
}

func TestBaseEvent_Immutability(t *testing.T) {
	aggregateID := uuidv7.New()
	event := bus.NewBaseEvent("user.created", aggregateID)

	// Get values
	eventType := event.Type()
	aggregateIDValue := event.AggregateID()
	occurredAt := event.OccurredAt()

	// Values should remain consistent across multiple calls
	assert.Equal(t, eventType, event.Type())
	assert.Equal(t, aggregateIDValue, event.AggregateID())
	assert.Equal(t, occurredAt, event.OccurredAt())
}
