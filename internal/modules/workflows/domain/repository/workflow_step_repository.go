package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowStepRepository defines the interface for workflow step data access
type IWorkflowStepRepository interface {
	// Create creates a new workflow step
	Create(ctx context.Context, step *entity.WorkflowStep) error

	// Update updates an existing workflow step
	Update(ctx context.Context, step *entity.WorkflowStep) error

	// GetByID retrieves a workflow step by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowStep, error)

	// ListByInstance lists all steps for a workflow instance
	ListByInstance(ctx context.Context, instanceID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowStep, error)

	// ListByInstanceAndState lists steps for a specific instance and state
	ListByInstanceAndState(ctx context.Context, instanceID uuidv7.UUID, state string) ([]*entity.WorkflowStep, error)

	// GetLatestByInstance retrieves the most recent step for an instance
	GetLatestByInstance(ctx context.Context, instanceID uuidv7.UUID) (*entity.WorkflowStep, error)

	// CountByInstance counts steps for a workflow instance
	CountByInstance(ctx context.Context, instanceID uuidv7.UUID) (int64, error)

	// CountByStatus counts steps by status for an instance
	CountByStatus(ctx context.Context, instanceID uuidv7.UUID, status entity.WorkflowStepStatus) (int64, error)

	// GetAverageDuration calculates average step duration for a definition
	GetAverageDuration(ctx context.Context, definitionID uuidv7.UUID, stepType entity.WorkflowStepType) (int64, error)
}
