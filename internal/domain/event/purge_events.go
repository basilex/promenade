package event

import (
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/google/uuid"
)

// PurgeCompletedEvent is published when a purge operation completes successfully
type PurgeCompletedEvent struct {
	bus.BaseEvent
	Summary entity.PurgeSummary `json:"summary"`
}

// NewPurgeCompletedEvent creates a new PurgeCompletedEvent
func NewPurgeCompletedEvent(summary entity.PurgeSummary) *PurgeCompletedEvent {
	// Use a zero UUID for system-level events
	return &PurgeCompletedEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicPurgeCompleted, uuid.Nil),
		Summary:   summary,
	}
}

// PurgeFailedEvent is published when a purge operation fails
type PurgeFailedEvent struct {
	bus.BaseEvent
	EntityName string `json:"entity_name"`
	Error      string `json:"error"`
}

// NewPurgeFailedEvent creates a new PurgeFailedEvent
func NewPurgeFailedEvent(entityName string, err error) *PurgeFailedEvent {
	// Use a zero UUID for system-level events
	return &PurgeFailedEvent{
		BaseEvent:  bus.NewBaseEvent(bus.TopicPurgeFailed, uuid.Nil),
		EntityName: entityName,
		Error:      err.Error(),
	}
}
