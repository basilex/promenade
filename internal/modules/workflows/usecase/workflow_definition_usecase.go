package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/internal/modules/workflows/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowDefinitionUseCase defines the interface for workflow definition use cases
type IWorkflowDefinitionUseCase interface {
	// Create creates a new workflow definition
	Create(ctx context.Context, name, description, version string, schema entity.WorkflowSchema, tags []string, metadata map[string]interface{}, createdBy uuidv7.UUID) (*entity.WorkflowDefinition, error)

	// GetByID retrieves a workflow definition by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error)

	// GetByName retrieves the latest active version by name
	GetByName(ctx context.Context, name string) (*entity.WorkflowDefinition, error)

	// List lists workflow definitions with filters
	List(ctx context.Context, status *entity.WorkflowDefinitionStatus, category *string, creatorID *uuidv7.UUID, limit, offset int) ([]*entity.WorkflowDefinition, int64, error)

	// Update updates a workflow definition (draft only)
	Update(ctx context.Context, id uuidv7.UUID, name, description *string, schema *entity.WorkflowSchema, tags []string) (*entity.WorkflowDefinition, error)

	// Delete soft-deletes a workflow definition (draft only)
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Activate activates a workflow definition
	Activate(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error)

	// Deprecate marks a workflow definition as deprecated (checks for active instances)
	Deprecate(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error)

	// Archive archives a workflow definition (optionally cancels running instances)
	Archive(ctx context.Context, id uuidv7.UUID, cancelRunningInstances bool) (*entity.WorkflowDefinition, error)
}

// workflowDefinitionUseCase implements IWorkflowDefinitionUseCase
type workflowDefinitionUseCase struct {
	definitionRepo repository.IWorkflowDefinitionRepository
	instanceRepo   repository.IWorkflowInstanceRepository
}

// NewWorkflowDefinitionUseCase creates a new workflow definition use case
func NewWorkflowDefinitionUseCase(
	definitionRepo repository.IWorkflowDefinitionRepository,
	instanceRepo repository.IWorkflowInstanceRepository,
) IWorkflowDefinitionUseCase {
	return &workflowDefinitionUseCase{
		definitionRepo: definitionRepo,
		instanceRepo:   instanceRepo,
	}
}

// Create creates a new workflow definition
func (uc *workflowDefinitionUseCase) Create(ctx context.Context, name, description, version string, schema entity.WorkflowSchema, tags []string, metadata map[string]interface{}, createdBy uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	// Check if name already exists
	existing, err := uc.definitionRepo.GetByName(ctx, name)
	if err == nil && existing != nil {
		return nil, ErrWorkflowNameAlreadyExists
	}

	// Create new workflow definition (version is optional, use default name as DisplayName)
	definition := entity.NewWorkflowDefinition(
		name,
		name, // DisplayName defaults to name
		description,
		schema,
		createdBy,
	)

	// Set optional fields
	if len(tags) > 0 {
		definition.Tags = tags
	}

	// Validate
	if err := definition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid workflow definition: %w", err)
	}

	// Save to repository
	if err := uc.definitionRepo.Create(ctx, definition); err != nil {
		return nil, fmt.Errorf("failed to create workflow definition: %w", err)
	}

	return definition, nil
}

// GetByID retrieves a workflow definition by ID
func (uc *workflowDefinitionUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	definition, err := uc.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow definition: %w", err)
	}
	return definition, nil
}

// GetByName retrieves the latest active version by name
func (uc *workflowDefinitionUseCase) GetByName(ctx context.Context, name string) (*entity.WorkflowDefinition, error) {
	definition, err := uc.definitionRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow definition by name: %w", err)
	}
	return definition, nil
}

// Activate activates a workflow definition
func (uc *workflowDefinitionUseCase) Activate(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	definition, err := uc.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}

	if err := definition.Activate(); err != nil {
		return nil, fmt.Errorf("failed to activate workflow definition: %w", err)
	}

	if err := uc.definitionRepo.Update(ctx, definition); err != nil {
		return nil, fmt.Errorf("failed to save workflow definition: %w", err)
	}

	return definition, nil
}

// List lists workflow definitions with filters
func (uc *workflowDefinitionUseCase) List(ctx context.Context, status *entity.WorkflowDefinitionStatus, category *string, creatorID *uuidv7.UUID, limit, offset int) ([]*entity.WorkflowDefinition, int64, error) {
	var definitions []*entity.WorkflowDefinition
	var total int64
	var err error

	// Apply filters based on provided parameters
	if status != nil {
		definitions, err = uc.definitionRepo.ListByStatus(ctx, *status, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list workflows by status: %w", err)
		}
		total, err = uc.definitionRepo.CountByStatus(ctx, *status)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count workflows by status: %w", err)
		}
	} else if category != nil {
		definitions, err = uc.definitionRepo.ListByCategory(ctx, *category, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list workflows by category: %w", err)
		}
		total, err = uc.definitionRepo.CountByCategory(ctx, *category)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count workflows by category: %w", err)
		}
	} else if creatorID != nil {
		definitions, err = uc.definitionRepo.ListByCreator(ctx, *creatorID, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list workflows by creator: %w", err)
		}
		// Note: CountByCreator not in interface, estimate from results
		total = int64(len(definitions))
	} else {
		// Default: list active workflows
		activeStatus := entity.WorkflowDefinitionStatusActive
		definitions, err = uc.definitionRepo.ListByStatus(ctx, activeStatus, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list active workflows: %w", err)
		}
		total, err = uc.definitionRepo.CountByStatus(ctx, activeStatus)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count active workflows: %w", err)
		}
	}

	return definitions, total, nil
}

// Update updates a workflow definition (draft only)
func (uc *workflowDefinitionUseCase) Update(ctx context.Context, id uuidv7.UUID, name, description *string, schema *entity.WorkflowSchema, tags []string) (*entity.WorkflowDefinition, error) {
	definition, err := uc.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}

	// Only draft definitions can be updated
	if definition.Status != entity.WorkflowDefinitionStatusDraft {
		return nil, fmt.Errorf("cannot update workflow definition in status %s: only draft definitions can be updated", definition.Status)
	}

	// Update fields if provided
	if name != nil {
		definition.Name = *name
	}
	if description != nil {
		definition.Description = *description
	}
	if schema != nil {
		// Set schema using jsonb.JSON.Set method
		definition.Definition.Set(*schema)
	}
	if tags != nil {
		definition.Tags = tags
	}

	if err := uc.definitionRepo.Update(ctx, definition); err != nil {
		return nil, fmt.Errorf("failed to update workflow definition: %w", err)
	}

	return definition, nil
}

// Delete soft-deletes a workflow definition (draft only)
func (uc *workflowDefinitionUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	definition, err := uc.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return ErrWorkflowNotFound
	}

	// Only draft definitions can be deleted
	if definition.Status != entity.WorkflowDefinitionStatusDraft {
		return fmt.Errorf("cannot delete workflow definition in status %s: only draft definitions can be deleted", definition.Status)
	}

	if err := uc.definitionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow definition: %w", err)
	}

	return nil
}

// Deprecate marks a workflow definition as deprecated
func (uc *workflowDefinitionUseCase) Deprecate(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	// Get workflow definition
	definition, err := uc.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}

	// Only active definitions can be deprecated
	if definition.Status != entity.WorkflowDefinitionStatusActive {
		return nil, fmt.Errorf("only active definitions can be deprecated, current status: %s", definition.Status)
	}

	// Check for active instances
	activeCount, err := uc.instanceRepo.CountActiveByDefinition(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to count active instances: %w", err)
	}

	if activeCount > 0 {
		return nil, fmt.Errorf("cannot deprecate workflow definition: %d active instances exist", activeCount)
	}

	// Mark as deprecated
	definition.Deprecate()

	// Update in repository
	if err := uc.definitionRepo.Update(ctx, definition); err != nil {
		return nil, fmt.Errorf("failed to deprecate workflow definition: %w", err)
	}

	return definition, nil
}

// Archive archives a workflow definition
func (uc *workflowDefinitionUseCase) Archive(ctx context.Context, id uuidv7.UUID, cancelRunningInstances bool) (*entity.WorkflowDefinition, error) {
	// Get workflow definition
	definition, err := uc.definitionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}

	// Cannot archive if already archived
	if definition.Status == entity.WorkflowDefinitionStatusArchived {
		return nil, fmt.Errorf("workflow definition is already archived")
	}

	// Get running instances (pending, running, waiting)
	activeInstances, err := uc.instanceRepo.ListByDefinition(ctx, id, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	// Filter only running instances
	var runningInstances []*entity.WorkflowInstance
	for _, instance := range activeInstances {
		if instance.IsActive() {
			runningInstances = append(runningInstances, instance)
		}
	}

	// If there are running instances
	if len(runningInstances) > 0 {
		if !cancelRunningInstances {
			return nil, fmt.Errorf("cannot archive workflow definition: %d running instances exist (use cancelRunningInstances=true to cancel them)", len(runningInstances))
		}

		// Cancel all running instances
		for _, instance := range runningInstances {
			if err := instance.Cancel(); err != nil {
				return nil, fmt.Errorf("failed to cancel instance %s: %w", instance.ID, err)
			}
			if err := uc.instanceRepo.Update(ctx, instance); err != nil {
				return nil, fmt.Errorf("failed to update cancelled instance %s: %w", instance.ID, err)
			}
		}
	}

	// Mark as archived
	if err := definition.Archive(); err != nil {
		return nil, fmt.Errorf("failed to archive: %w", err)
	}

	// Update in repository
	if err := uc.definitionRepo.Update(ctx, definition); err != nil {
		return nil, fmt.Errorf("failed to save archived workflow definition: %w", err)
	}

	return definition, nil
}
