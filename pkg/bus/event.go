package bus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BaseEvent provides a common implementation of the Event interface.
// Domain events should embed this struct.
type BaseEvent struct {
	EventType    string            `json:"event_type"`
	EventID      uuid.UUID         `json:"event_id"`
	AggregateId  uuid.UUID         `json:"aggregate_id"`
	OccurredTime time.Time         `json:"occurred_at"`
	Meta         map[string]string `json:"metadata,omitempty"`
}

// Type implements Event interface.
func (e BaseEvent) Type() string {
	return e.EventType
}

// OccurredAt implements Event interface.
func (e BaseEvent) OccurredAt() time.Time {
	return e.OccurredTime
}

// AggregateID implements Event interface.
func (e BaseEvent) AggregateID() uuid.UUID {
	return e.AggregateId
}

// Metadata implements Event interface.
func (e BaseEvent) Metadata() map[string]string {
	return e.Meta
}

// NewBaseEvent creates a new base event with common fields populated.
func NewBaseEvent(eventType string, aggregateID uuid.UUID) BaseEvent {
	return BaseEvent{
		EventType:    eventType,
		EventID:      uuid.New(),
		AggregateId:  aggregateID,
		OccurredTime: time.Now().UTC(),
		Meta:         make(map[string]string),
	}
}

// EventSerializer handles serialization/deserialization of events for transport.
type EventSerializer interface {
	Serialize(event Event) ([]byte, error)
	Deserialize(data []byte) (Event, error)
}

// JSONEventSerializer implements EventSerializer using JSON encoding.
type JSONEventSerializer struct{}

// Serialize converts an event to JSON bytes.
func (s *JSONEventSerializer) Serialize(event Event) ([]byte, error) {
	return json.Marshal(event)
}

// Deserialize converts JSON bytes back to an event.
// Note: This requires the concrete event type to be registered.
func (s *JSONEventSerializer) Deserialize(data []byte) (Event, error) {
	var base BaseEvent
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}
	return base, nil
}

// EventRegistry allows registration of concrete event types for deserialization.
type EventRegistry struct {
	types map[string]func() Event
}

// NewEventRegistry creates a new event registry.
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		types: make(map[string]func() Event),
	}
}

// Register adds a concrete event type to the registry.
func (r *EventRegistry) Register(eventType string, factory func() Event) {
	r.types[eventType] = factory
}

// Create instantiates a registered event type.
func (r *EventRegistry) Create(eventType string) (Event, bool) {
	factory, exists := r.types[eventType]
	if !exists {
		return nil, false
	}
	return factory(), true
}
