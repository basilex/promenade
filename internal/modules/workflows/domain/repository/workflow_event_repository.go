package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowEventRepository defines the interface for workflow event data access
type IWorkflowEventRepository interface {
	// Create creates a new workflow event
	Create(ctx context.Context, event *entity.WorkflowEvent) error

	// Update updates an existing workflow event
	Update(ctx context.Context, event *entity.WorkflowEvent) error

	// GetByID retrieves a workflow event by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowEvent, error)

	// ListByInstance lists all events for a specific instance
	ListByInstance(ctx context.Context, instanceID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowEvent, error)

	// ListUnprocessed lists all unprocessed events ready for processing
	ListUnprocessed(ctx context.Context, limit int) ([]*entity.WorkflowEvent, error)

	// ListByType lists events by type
	ListByType(ctx context.Context, eventType entity.WorkflowEventType, limit, offset int) ([]*entity.WorkflowEvent, error)

	// CountByInstance counts events for a specific instance
	CountByInstance(ctx context.Context, instanceID uuidv7.UUID) (int64, error)

	// CountUnprocessed counts unprocessed events
	CountUnprocessed(ctx context.Context) (int64, error)
}
