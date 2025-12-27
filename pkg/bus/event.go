package bus

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// BaseEvent provides a common implementation of the Event interface.
// Domain events should embed this struct.
type BaseEvent struct {
	EventType    string            `json:"event_type"`
	EventID      uuidv7.UUID       `json:"event_id"`
	AggregateId  uuidv7.UUID       `json:"aggregate_id"`
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
func (e BaseEvent) AggregateID() uuidv7.UUID {
	return e.AggregateId
}

// Metadata implements Event interface.
func (e BaseEvent) Metadata() map[string]string {
	return e.Meta
}

// NewBaseEvent creates a new base event with common fields populated.
func NewBaseEvent(eventType string, aggregateID uuidv7.UUID) BaseEvent {
	return BaseEvent{
		EventType:    eventType,
		EventID:      uuidv7.New(),
		AggregateId:  aggregateID,
		OccurredTime: time.Now().UTC(),
		Meta:         make(map[string]string),
	}
}
