package usecase

import (
	"time"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// NotificationRequestedEvent is published when a notification is requested
type NotificationRequestedEvent struct {
	*bus.BaseEvent
	NotificationID uuidv7.UUID `json:"notification_id"`
	UserID         uuidv7.UUID `json:"user_id"`
	NotifType      string      `json:"type"`
	Channel        string      `json:"channel"`
}

// Type returns event type
func (e *NotificationRequestedEvent) Type() string {
	return e.EventType
}

// OccurredAt returns event time
func (e *NotificationRequestedEvent) OccurredAt() time.Time {
	return e.OccurredTime
}

// AggregateID returns aggregate ID
func (e *NotificationRequestedEvent) AggregateID() uuidv7.UUID {
	return e.AggregateId
}

// Metadata returns event metadata
func (e *NotificationRequestedEvent) Metadata() map[string]string {
	return e.Meta
}

// NotificationSentEvent is published when a notification is sent
type NotificationSentEvent struct {
	*bus.BaseEvent
	NotificationID uuidv7.UUID `json:"notification_id"`
	UserID         uuidv7.UUID `json:"user_id"`
	NotifType      string      `json:"type"`
	Channel        string      `json:"channel"`
	Success        bool        `json:"success"`
	ErrorMessage   string      `json:"error_message,omitempty"`
}

// Type returns event type
func (e *NotificationSentEvent) Type() string {
	return e.EventType
}

// OccurredAt returns event time
func (e *NotificationSentEvent) OccurredAt() time.Time {
	return e.OccurredTime
}

// AggregateID returns aggregate ID
func (e *NotificationSentEvent) AggregateID() uuidv7.UUID {
	return e.AggregateId
}

// Metadata returns event metadata
func (e *NotificationSentEvent) Metadata() map[string]string {
	return e.Meta
}
