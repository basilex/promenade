package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowDefinitionRepository defines the interface for workflow definition data access
type IWorkflowDefinitionRepository interface {
	// Create creates a new workflow definition
	Create(ctx context.Context, definition *entity.WorkflowDefinition) error

	// Update updates an existing workflow definition
	Update(ctx context.Context, definition *entity.WorkflowDefinition) error

	// Delete soft-deletes a workflow definition
	Delete(ctx context.Context, id uuidv7.UUID) error

	// GetByID retrieves a workflow definition by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error)

	// GetByName retrieves a workflow definition by name (latest active version)
	GetByName(ctx context.Context, name string) (*entity.WorkflowDefinition, error)

	// GetByNameAndVersion retrieves a specific version of a workflow definition
	GetByNameAndVersion(ctx context.Context, name string, version int) (*entity.WorkflowDefinition, error)

	// ListByStatus lists workflow definitions by status
	ListByStatus(ctx context.Context, status entity.WorkflowDefinitionStatus, limit, offset int) ([]*entity.WorkflowDefinition, error)

	// ListByCategory lists workflow definitions by category
	ListByCategory(ctx context.Context, category string, limit, offset int) ([]*entity.WorkflowDefinition, error)

	// ListByCreator lists workflow definitions by creator
	ListByCreator(ctx context.Context, creatorID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowDefinition, error)

	// ListVersions lists all versions of a workflow definition
	ListVersions(ctx context.Context, name string) ([]*entity.WorkflowDefinition, error)

	// CountByStatus counts workflow definitions by status
	CountByStatus(ctx context.Context, status entity.WorkflowDefinitionStatus) (int64, error)

	// CountByCategory counts workflow definitions by category
	CountByCategory(ctx context.Context, category string) (int64, error)

	// Search searches workflow definitions by name or description
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.WorkflowDefinition, error)
}
