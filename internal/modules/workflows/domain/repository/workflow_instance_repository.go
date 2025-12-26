package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowInstanceRepository defines the interface for workflow instance data access
type IWorkflowInstanceRepository interface {
	// Create creates a new workflow instance
	Create(ctx context.Context, instance *entity.WorkflowInstance) error

	// Update updates an existing workflow instance
	Update(ctx context.Context, instance *entity.WorkflowInstance) error

	// Delete soft-deletes a workflow instance
	Delete(ctx context.Context, id uuidv7.UUID) error

	// GetByID retrieves a workflow instance by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error)

	// GetByExternalReference retrieves a workflow instance by external reference
	GetByExternalReference(ctx context.Context, externalRef string) (*entity.WorkflowInstance, error)

	// ListByDefinition lists workflow instances for a specific definition
	ListByDefinition(ctx context.Context, definitionID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error)

	// ListByStatus lists workflow instances by status
	ListByStatus(ctx context.Context, status entity.WorkflowInstanceStatus, limit, offset int) ([]*entity.WorkflowInstance, error)

	// ListByAssignee lists workflow instances assigned to a user
	ListByAssignee(ctx context.Context, assigneeID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error)

	// ListByCreator lists workflow instances created by a user
	ListByCreator(ctx context.Context, creatorID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error)

	// ListOverdue lists workflow instances that are past their due date
	ListOverdue(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error)

	// ListTimedOut lists workflow instances where current state has timed out
	ListTimedOut(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error)

	// ListActive lists all active workflow instances (pending, running, waiting)
	ListActive(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error)

	// CountByStatus counts workflow instances by status
	CountByStatus(ctx context.Context, status entity.WorkflowInstanceStatus) (int64, error)

	// CountByDefinition counts workflow instances for a definition
	CountByDefinition(ctx context.Context, definitionID uuidv7.UUID) (int64, error)

	// CountActiveByDefinition counts active instances (pending, running, waiting) for a definition
	CountActiveByDefinition(ctx context.Context, definitionID uuidv7.UUID) (int64, error)

	// CountByAssignee counts workflow instances assigned to a user
	CountByAssignee(ctx context.Context, assigneeID uuidv7.UUID) (int64, error)

	// CountActive counts all active workflow instances
	CountActive(ctx context.Context) (int64, error)
}
