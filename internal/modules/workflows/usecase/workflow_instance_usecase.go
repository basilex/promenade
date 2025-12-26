package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/internal/modules/workflows/domain/repository"
	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IWorkflowInstanceUseCase defines the interface for workflow instance operations
type IWorkflowInstanceUseCase interface {
	// StartWorkflow starts a new workflow instance
	StartWorkflow(ctx context.Context, definitionID uuidv7.UUID, externalReference *string, priority int, workflowContext map[string]interface{}, assignedTo *uuidv7.UUID, startedBy uuidv7.UUID) (*entity.WorkflowInstance, error)

	// GetInstance retrieves a workflow instance by ID
	GetInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error)

	// ListInstances lists workflow instances with filters
	ListInstances(ctx context.Context, definitionID *uuidv7.UUID, status string, assignedTo *uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error)

	// UpdateInstance updates workflow instance properties (priority, assignee)
	UpdateInstance(ctx context.Context, id uuidv7.UUID, priority *int, assignedTo *uuidv7.UUID) (*entity.WorkflowInstance, error)

	// PauseInstance pauses a running workflow instance
	PauseInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error)

	// ResumeInstance resumes a paused workflow instance
	ResumeInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error)

	// CancelInstance cancels a workflow instance
	CancelInstance(ctx context.Context, id uuidv7.UUID, reason string) (*entity.WorkflowInstance, error)
}

// workflowInstanceUseCase implements IWorkflowInstanceUseCase
type workflowInstanceUseCase struct {
	instanceRepo   repository.IWorkflowInstanceRepository
	definitionRepo repository.IWorkflowDefinitionRepository
}

// NewWorkflowInstanceUseCase creates a new workflow instance use case
func NewWorkflowInstanceUseCase(
	instanceRepo repository.IWorkflowInstanceRepository,
	definitionRepo repository.IWorkflowDefinitionRepository,
) IWorkflowInstanceUseCase {
	return &workflowInstanceUseCase{
		instanceRepo:   instanceRepo,
		definitionRepo: definitionRepo,
	}
}

// StartWorkflow starts a new workflow instance
func (uc *workflowInstanceUseCase) StartWorkflow(ctx context.Context, definitionID uuidv7.UUID, externalReference *string, priority int, workflowContext map[string]interface{}, assignedTo *uuidv7.UUID, startedBy uuidv7.UUID) (*entity.WorkflowInstance, error) {
	// Get workflow definition
	definition, err := uc.definitionRepo.GetByID(ctx, definitionID)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}

	// Check if definition is active
	if definition.Status != entity.WorkflowDefinitionStatusActive {
		return nil, ErrWorkflowNotActive
	}

	// Use context map directly as jsonb.Map
	var context jsonb.Map
	if workflowContext != nil {
		context = workflowContext
	} else {
		context = jsonb.Map{}
	}

	// Create workflow instance
	instance := entity.NewWorkflowInstance(
		definitionID,
		definition.Version,
		definition.Definition.Data.InitialState,
		context,
		startedBy,
	)

	if externalReference != nil {
		instance.ExternalReference = externalReference
	}
	if assignedTo != nil {
		instance.AssignedTo = assignedTo
	}
	instance.Priority = priority

	// Start the instance immediately (transition from pending to running)
	if err := instance.Start(); err != nil {
		return nil, fmt.Errorf("failed to start workflow instance: %w", err)
	}

	// Save to repository
	if err := uc.instanceRepo.Create(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to create workflow instance: %w", err)
	}

	return instance, nil
}

// GetInstance retrieves a workflow instance by ID
func (uc *workflowInstanceUseCase) GetInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	instance, err := uc.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInstanceNotFound
	}
	return instance, nil
}

// ListInstances lists workflow instances with filters
func (uc *workflowInstanceUseCase) ListInstances(ctx context.Context, definitionID *uuidv7.UUID, status string, assignedTo *uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	// TODO: Implement proper filtering based on parameters
	// For now, return all active instances
	instances, err := uc.instanceRepo.ListActive(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow instances: %w", err)
	}
	return instances, nil
}

// UpdateInstance updates workflow instance properties (priority, assignee)
func (uc *workflowInstanceUseCase) UpdateInstance(ctx context.Context, id uuidv7.UUID, priority *int, assignedTo *uuidv7.UUID) (*entity.WorkflowInstance, error) {
	instance, err := uc.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInstanceNotFound
	}

	// Update priority if provided
	if priority != nil {
		// Validate priority range (1-10)
		if *priority < 1 || *priority > 10 {
			return nil, fmt.Errorf("priority must be between 1 and 10")
		}
		instance.Priority = *priority
	}

	// Update assignee if provided
	if assignedTo != nil {
		instance.AssignedTo = assignedTo
	}

	if err := uc.instanceRepo.Update(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to update workflow instance: %w", err)
	}

	return instance, nil
}

// PauseInstance pauses a running workflow instance
func (uc *workflowInstanceUseCase) PauseInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	instance, err := uc.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInstanceNotFound
	}

	if err := instance.Pause(); err != nil {
		return nil, fmt.Errorf("failed to pause workflow instance: %w", err)
	}

	if err := uc.instanceRepo.Update(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to save workflow instance: %w", err)
	}

	return instance, nil
}

// ResumeInstance resumes a paused workflow instance
func (uc *workflowInstanceUseCase) ResumeInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	instance, err := uc.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInstanceNotFound
	}

	if err := instance.Resume(); err != nil {
		return nil, fmt.Errorf("failed to resume workflow instance: %w", err)
	}

	if err := uc.instanceRepo.Update(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to save workflow instance: %w", err)
	}

	return instance, nil
}

// CancelInstance cancels a workflow instance
func (uc *workflowInstanceUseCase) CancelInstance(ctx context.Context, id uuidv7.UUID, reason string) (*entity.WorkflowInstance, error) {
	instance, err := uc.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInstanceNotFound
	}

	instance.Cancel()
	if reason != "" {
		instance.ErrorMessage = &reason
	}

	if err := uc.instanceRepo.Update(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to save workflow instance: %w", err)
	}

	return instance, nil
}
