package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowVariableRepository defines the interface for workflow variable data access
type IWorkflowVariableRepository interface {
	// Create creates a new workflow variable
	Create(ctx context.Context, variable *entity.WorkflowVariable) error

	// Update updates an existing workflow variable
	Update(ctx context.Context, variable *entity.WorkflowVariable) error

	// Delete deletes a workflow variable
	Delete(ctx context.Context, id uuidv7.UUID) error

	// GetByID retrieves a workflow variable by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowVariable, error)

	// GetByName retrieves the latest variable by name for an instance
	GetByName(ctx context.Context, instanceID uuidv7.UUID, name string) (*entity.WorkflowVariable, error)

	// ListByInstance lists all variables for a workflow instance
	ListByInstance(ctx context.Context, instanceID uuidv7.UUID) ([]*entity.WorkflowVariable, error)

	// ListByScope lists variables by scope
	ListByScope(ctx context.Context, instanceID uuidv7.UUID, scope entity.WorkflowVariableScope) ([]*entity.WorkflowVariable, error)

	// ListByState lists variables for a specific state
	ListByState(ctx context.Context, instanceID uuidv7.UUID, stateName string) ([]*entity.WorkflowVariable, error)

	// DeleteByInstance deletes all variables for an instance
	DeleteByInstance(ctx context.Context, instanceID uuidv7.UUID) error

	// DeleteByScope deletes variables by scope
	DeleteByScope(ctx context.Context, instanceID uuidv7.UUID, scope entity.WorkflowVariableScope) error
}
