package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// StartWorkflow Tests
// ============================================================================

func TestStartWorkflow_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	definitionID := uuidv7.New()
	startedBy := uuidv7.New()
	definition := &entity.WorkflowDefinition{
		ID:      definitionID,
		Name:    "test_workflow",
		Version: 1,
		Status:  entity.WorkflowDefinitionStatusActive,
		Definition: jsonb.JSON[entity.WorkflowSchema]{
			Data: entity.WorkflowSchema{
				InitialState: "start",
				States: []entity.WorkflowState{
					{Name: "start", Type: "start"},
				},
			},
			Valid: true,
		},
	}

	workflowContext := map[string]interface{}{
		"order_id": "12345",
		"customer": "John Doe",
	}

	// Mock expectations
	definitionRepo.On("GetByID", mock.Anything, definitionID).Return(definition, nil)
	instanceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	instance, err := uc.StartWorkflow(context.Background(), definitionID, nil, 5, workflowContext, nil, startedBy)

	assert.NoError(t, err)
	require.NotNil(t, instance)
	assert.Equal(t, definitionID, instance.DefinitionID)
	assert.Equal(t, entity.WorkflowInstanceStatusRunning, instance.Status) // StartWorkflow calls instance.Start()
	assert.Equal(t, "start", instance.CurrentState)
	assert.Equal(t, 5, instance.Priority)
	definitionRepo.AssertExpectations(t)
	instanceRepo.AssertExpectations(t)
}

func TestStartWorkflow_WithExternalReference(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	definitionID := uuidv7.New()
	startedBy := uuidv7.New()
	externalRef := "ORDER-2025-001"
	
	definition := &entity.WorkflowDefinition{
		ID:      definitionID,
		Status:  entity.WorkflowDefinitionStatusActive,
		Version: 1,
		Definition: jsonb.JSON[entity.WorkflowSchema]{
			Data: entity.WorkflowSchema{
				InitialState: "start",
				States: []entity.WorkflowState{
					{Name: "start", Type: "start"},
				},
			},
			Valid: true,
		},
	}

	definitionRepo.On("GetByID", mock.Anything, definitionID).Return(definition, nil)
	instanceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	instance, err := uc.StartWorkflow(context.Background(), definitionID, &externalRef, 5, nil, nil, startedBy)

	assert.NoError(t, err)
	require.NotNil(t, instance)
	assert.NotNil(t, instance.ExternalReference)
	assert.Equal(t, externalRef, *instance.ExternalReference)
	definitionRepo.AssertExpectations(t)
	instanceRepo.AssertExpectations(t)
}

func TestStartWorkflow_WithAssignee(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	definitionID := uuidv7.New()
	startedBy := uuidv7.New()
	assignedTo := uuidv7.New()
	
	definition := &entity.WorkflowDefinition{
		ID:      definitionID,
		Status:  entity.WorkflowDefinitionStatusActive,
		Version: 1,
		Definition: jsonb.JSON[entity.WorkflowSchema]{
			Data: entity.WorkflowSchema{
				InitialState: "start",
				States: []entity.WorkflowState{
					{Name: "start", Type: "start"},
				},
			},
			Valid: true,
		},
	}

	definitionRepo.On("GetByID", mock.Anything, definitionID).Return(definition, nil)
	instanceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	instance, err := uc.StartWorkflow(context.Background(), definitionID, nil, 5, nil, &assignedTo, startedBy)

	assert.NoError(t, err)
	require.NotNil(t, instance)
	assert.NotNil(t, instance.AssignedTo)
	assert.Equal(t, assignedTo, *instance.AssignedTo)
	definitionRepo.AssertExpectations(t)
	instanceRepo.AssertExpectations(t)
}

func TestStartWorkflow_DefinitionNotFound(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	definitionID := uuidv7.New()
	startedBy := uuidv7.New()

	definitionRepo.On("GetByID", mock.Anything, definitionID).Return(nil, errors.New("not found"))

	instance, err := uc.StartWorkflow(context.Background(), definitionID, nil, 5, nil, nil, startedBy)

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.Equal(t, ErrWorkflowNotFound, err)
	definitionRepo.AssertExpectations(t)
}

func TestStartWorkflow_DefinitionNotActive(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	definitionID := uuidv7.New()
	startedBy := uuidv7.New()
	definition := &entity.WorkflowDefinition{
		ID:      definitionID,
		Status:  entity.WorkflowDefinitionStatusDraft, // Not active
		Version: 1,
		Definition: jsonb.JSON[entity.WorkflowSchema]{
			Data: entity.WorkflowSchema{
				InitialState: "start",
			},
			Valid: true,
		},
	}

	definitionRepo.On("GetByID", mock.Anything, definitionID).Return(definition, nil)

	instance, err := uc.StartWorkflow(context.Background(), definitionID, nil, 5, nil, nil, startedBy)

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.Equal(t, ErrWorkflowNotActive, err)
	definitionRepo.AssertExpectations(t)
}

func TestStartWorkflow_RepositoryError(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	definitionID := uuidv7.New()
	startedBy := uuidv7.New()
	definition := &entity.WorkflowDefinition{
		ID:      definitionID,
		Status:  entity.WorkflowDefinitionStatusActive,
		Version: 1,
		Definition: jsonb.JSON[entity.WorkflowSchema]{
			Data: entity.WorkflowSchema{
				InitialState: "start",
				States: []entity.WorkflowState{
					{Name: "start", Type: "start"},
				},
			},
			Valid: true,
		},
	}

	definitionRepo.On("GetByID", mock.Anything, definitionID).Return(definition, nil)
	instanceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).
		Return(errors.New("database error"))

	instance, err := uc.StartWorkflow(context.Background(), definitionID, nil, 5, nil, nil, startedBy)

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.Contains(t, err.Error(), "failed to create workflow instance")
	definitionRepo.AssertExpectations(t)
	instanceRepo.AssertExpectations(t)
}

// ============================================================================
// GetInstance Tests
// ============================================================================

func TestGetInstance_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	expectedInstance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(expectedInstance, nil)

	instance, err := uc.GetInstance(context.Background(), instanceID)

	assert.NoError(t, err)
	require.NotNil(t, instance)
	assert.Equal(t, instanceID, instance.ID)
	assert.Equal(t, entity.WorkflowInstanceStatusRunning, instance.Status)
	instanceRepo.AssertExpectations(t)
}

func TestGetInstance_NotFound(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(nil, errors.New("not found"))

	instance, err := uc.GetInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.Equal(t, ErrInstanceNotFound, err)
	instanceRepo.AssertExpectations(t)
}

// ============================================================================
// ListInstances Tests
// ============================================================================

func TestListInstances_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	expectedInstances := []*entity.WorkflowInstance{
		{
			ID:           uuidv7.New(),
			DefinitionID: uuidv7.New(),
			Status:       entity.WorkflowInstanceStatusRunning,
		},
		{
			ID:           uuidv7.New(),
			DefinitionID: uuidv7.New(),
			Status:       entity.WorkflowInstanceStatusPending,
		},
	}

	instanceRepo.On("ListActive", mock.Anything, 10, 0).Return(expectedInstances, nil)

	instances, err := uc.ListInstances(context.Background(), nil, "", nil, 10, 0)

	assert.NoError(t, err)
	require.NotNil(t, instances)
	assert.Len(t, instances, 2)
	instanceRepo.AssertExpectations(t)
}

func TestListInstances_RepositoryError(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceRepo.On("ListActive", mock.Anything, 10, 0).Return(nil, errors.New("database error"))

	instances, err := uc.ListInstances(context.Background(), nil, "", nil, 10, 0)

	assert.Error(t, err)
	assert.Nil(t, instances)
	assert.Contains(t, err.Error(), "failed to list workflow instances")
	instanceRepo.AssertExpectations(t)
}

// ============================================================================
// PauseInstance Tests
// ============================================================================

func TestPauseInstance_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	result, err := uc.PauseInstance(context.Background(), instanceID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowInstanceStatusPaused, result.Status)
	instanceRepo.AssertExpectations(t)
}

func TestPauseInstance_NotFound(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(nil, errors.New("not found"))

	result, err := uc.PauseInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInstanceNotFound, err)
	instanceRepo.AssertExpectations(t)
}

func TestPauseInstance_InvalidState(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusCompleted, // Cannot pause completed instance
		CurrentState: "final",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)

	result, err := uc.PauseInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to pause workflow instance")
	instanceRepo.AssertExpectations(t)
}

func TestPauseInstance_UpdateError(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).
		Return(errors.New("database error"))

	result, err := uc.PauseInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save workflow instance")
	instanceRepo.AssertExpectations(t)
}

// ============================================================================
// ResumeInstance Tests
// ============================================================================

func TestResumeInstance_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusPaused,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	result, err := uc.ResumeInstance(context.Background(), instanceID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowInstanceStatusRunning, result.Status)
	instanceRepo.AssertExpectations(t)
}

func TestResumeInstance_NotFound(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(nil, errors.New("not found"))

	result, err := uc.ResumeInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInstanceNotFound, err)
	instanceRepo.AssertExpectations(t)
}

func TestResumeInstance_NotPaused(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning, // Not paused
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)

	result, err := uc.ResumeInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to resume workflow instance")
	instanceRepo.AssertExpectations(t)
}

func TestResumeInstance_UpdateError(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusPaused,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).
		Return(errors.New("database error"))

	result, err := uc.ResumeInstance(context.Background(), instanceID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save workflow instance")
	instanceRepo.AssertExpectations(t)
}

// ============================================================================
// CancelInstance Tests
// ============================================================================

func TestCancelInstance_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		CurrentState: "processing",
	}

	reason := "User requested cancellation"
	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	result, err := uc.CancelInstance(context.Background(), instanceID, reason)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowInstanceStatusCancelled, result.Status)
	assert.NotNil(t, result.ErrorMessage)
	assert.Equal(t, reason, *result.ErrorMessage)
	instanceRepo.AssertExpectations(t)
}

func TestCancelInstance_WithoutReason(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).Return(nil)

	result, err := uc.CancelInstance(context.Background(), instanceID, "")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowInstanceStatusCancelled, result.Status)
	instanceRepo.AssertExpectations(t)
}

func TestCancelInstance_NotFound(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(nil, errors.New("not found"))

	result, err := uc.CancelInstance(context.Background(), instanceID, "test reason")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInstanceNotFound, err)
	instanceRepo.AssertExpectations(t)
}

func TestCancelInstance_UpdateError(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		CurrentState: "processing",
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).
		Return(errors.New("database error"))

	result, err := uc.CancelInstance(context.Background(), instanceID, "test reason")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save workflow instance")
	instanceRepo.AssertExpectations(t)
}

// ============================================================================
// UpdateInstance Tests
// ============================================================================

func TestUpdateInstance_UpdatePriority_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		Priority:     5,
	}

	newPriority := 8

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).
		Run(func(args mock.Arguments) {
			updated := args.Get(1).(*entity.WorkflowInstance)
			assert.Equal(t, newPriority, updated.Priority)
		}).
		Return(nil)

	result, err := uc.UpdateInstance(context.Background(), instanceID, &newPriority, nil)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newPriority, result.Priority)
	instanceRepo.AssertExpectations(t)
}

func TestUpdateInstance_UpdateAssignee_Success(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	newAssignee := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
		AssignedTo:   nil,
	}

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)
	instanceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowInstance")).
		Run(func(args mock.Arguments) {
			updated := args.Get(1).(*entity.WorkflowInstance)
			require.NotNil(t, updated.AssignedTo)
			assert.Equal(t, newAssignee, *updated.AssignedTo)
		}).
		Return(nil)

	result, err := uc.UpdateInstance(context.Background(), instanceID, nil, &newAssignee)

	assert.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.AssignedTo)
	assert.Equal(t, newAssignee, *result.AssignedTo)
	instanceRepo.AssertExpectations(t)
}

func TestUpdateInstance_InvalidPriority_TooLow(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
	}

	invalidPriority := 0 // Must be 1-10

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)

	result, err := uc.UpdateInstance(context.Background(), instanceID, &invalidPriority, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "priority must be between 1 and 10")
	instanceRepo.AssertExpectations(t)
}

func TestUpdateInstance_InvalidPriority_TooHigh(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	instance := &entity.WorkflowInstance{
		ID:           instanceID,
		DefinitionID: uuidv7.New(),
		Status:       entity.WorkflowInstanceStatusRunning,
	}

	invalidPriority := 11 // Must be 1-10

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(instance, nil)

	result, err := uc.UpdateInstance(context.Background(), instanceID, &invalidPriority, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "priority must be between 1 and 10")
	instanceRepo.AssertExpectations(t)
}

func TestUpdateInstance_NotFound(t *testing.T) {
	instanceRepo := new(MockWorkflowInstanceRepository)
	definitionRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	instanceID := uuidv7.New()
	newPriority := 5

	instanceRepo.On("GetByID", mock.Anything, instanceID).Return(nil, errors.New("not found"))

	result, err := uc.UpdateInstance(context.Background(), instanceID, &newPriority, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInstanceNotFound, err)
	instanceRepo.AssertExpectations(t)
}
