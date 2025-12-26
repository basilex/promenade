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

func TestCreateDefinition_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	name := "order_process"
	description := "Process customer orders"
	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "start"}},
		InitialState: "start",
	}
	createdBy := uuidv7.New()

	// Mock expects GetByName (for duplicate check) and Create to succeed
	mockRepo.On("GetByName", mock.Anything, name).Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).
		Return(nil)

	def, err := uc.Create(context.Background(), name, description, "", schema, []string{}, nil, createdBy)

	assert.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, name, def.Name)
	assert.Equal(t, name, def.DisplayName) // DisplayName defaults to name
	assert.Equal(t, entity.WorkflowDefinitionStatusDraft, def.Status)
	mockRepo.AssertExpectations(t)
}

func TestCreateDefinition_RepositoryError(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "start"}},
		InitialState: "start",
	}

	// Mock GetByName (duplicate check) and Create returns error
	mockRepo.On("GetByName", mock.Anything, "test").Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.Anything, mock.Anything).
		Return(errors.New("database error"))

	def, err := uc.Create(context.Background(), "test", "Test description", "", schema, []string{}, nil, uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestGetDefinition_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	expectedDef := &entity.WorkflowDefinition{
		ID:          defID,
		Name:        "test_workflow",
		DisplayName: "Test Workflow",
		Status:      entity.WorkflowDefinitionStatusActive,
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(expectedDef, nil)

	def, err := uc.GetByID(context.Background(), defID)

	assert.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, defID, def.ID)
	assert.Equal(t, "test_workflow", def.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetDefinition_NotFound(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	mockRepo.On("GetByID", mock.Anything, defID).Return(nil, errors.New("not found"))

	def, err := uc.GetByID(context.Background(), defID)

	assert.Error(t, err)
	assert.Nil(t, def)
	mockRepo.AssertExpectations(t)
}

func TestActivate_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test",
		Status: entity.WorkflowDefinitionStatusDraft,
	}

	// Mock GetByID and Update
	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).Return(nil)

	_, err := uc.Activate(context.Background(), defID)

	assert.NoError(t, err)
	assert.Equal(t, entity.WorkflowDefinitionStatusActive, def.Status)
	mockRepo.AssertExpectations(t)
}

// Phase 2.2: Additional comprehensive tests

func TestCreate_DuplicateName(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	name := "existing_workflow"
	existingDef := &entity.WorkflowDefinition{
		ID:   uuidv7.New(),
		Name: name,
	}

	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "start"}},
		InitialState: "start",
	}

	// Mock GetByName returns existing workflow (duplicate detected)
	mockRepo.On("GetByName", mock.Anything, name).Return(existingDef, nil)

	def, err := uc.Create(context.Background(), name, "Description", "", schema, []string{}, nil, uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Equal(t, ErrWorkflowNameAlreadyExists, err)
	mockRepo.AssertExpectations(t)
}

func TestCreate_ValidationError_EmptyName(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "start"}},
		InitialState: "start",
	}

	// Mock GetByName for duplicate check (empty name won't exist)
	mockRepo.On("GetByName", mock.Anything, "").Return(nil, errors.New("not found"))

	def, err := uc.Create(context.Background(), "", "Description", "", schema, []string{}, nil, uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Contains(t, err.Error(), "invalid workflow definition")
	mockRepo.AssertExpectations(t)
}

func TestCreate_ValidationError_NoStates(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	name := "invalid_workflow"
	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{}, // Empty states
		InitialState: "start",
	}

	// Mock GetByName for duplicate check
	mockRepo.On("GetByName", mock.Anything, name).Return(nil, errors.New("not found"))

	def, err := uc.Create(context.Background(), name, "Description", "", schema, []string{}, nil, uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Contains(t, err.Error(), "invalid workflow definition")
	assert.Contains(t, err.Error(), "at least one state is required")
	mockRepo.AssertExpectations(t)
}

func TestCreate_ValidationError_NoInitialState(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	name := "invalid_workflow"
	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "start"}},
		InitialState: "", // Missing initial state
	}

	// Mock GetByName for duplicate check
	mockRepo.On("GetByName", mock.Anything, name).Return(nil, errors.New("not found"))

	def, err := uc.Create(context.Background(), name, "Description", "", schema, []string{}, nil, uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Contains(t, err.Error(), "invalid workflow definition")
	assert.Contains(t, err.Error(), "initial state is required")
	mockRepo.AssertExpectations(t)
}

func TestGetByName_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	name := "order_workflow"
	expectedDef := &entity.WorkflowDefinition{
		ID:          uuidv7.New(),
		Name:        name,
		DisplayName: "Order Workflow",
		Status:      entity.WorkflowDefinitionStatusActive,
	}

	mockRepo.On("GetByName", mock.Anything, name).Return(expectedDef, nil)

	def, err := uc.GetByName(context.Background(), name)

	assert.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, name, def.Name)
	assert.Equal(t, entity.WorkflowDefinitionStatusActive, def.Status)
	mockRepo.AssertExpectations(t)
}

func TestGetByName_NotFound(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	name := "nonexistent_workflow"
	mockRepo.On("GetByName", mock.Anything, name).Return(nil, errors.New("not found"))

	def, err := uc.GetByName(context.Background(), name)

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Contains(t, err.Error(), "failed to get workflow definition by name")
	mockRepo.AssertExpectations(t)
}

func TestActivate_AlreadyActive(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test",
		Status: entity.WorkflowDefinitionStatusActive, // Already active
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)

	_, err := uc.Activate(context.Background(), defID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only draft definitions can be activated")
	mockRepo.AssertExpectations(t)
}

func TestActivate_NotFound(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	mockRepo.On("GetByID", mock.Anything, defID).Return(nil, errors.New("not found"))

	def, err := uc.Activate(context.Background(), defID)

	assert.Error(t, err)
	assert.Nil(t, def)
	assert.Equal(t, ErrWorkflowNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestActivate_UpdateError(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test",
		Status: entity.WorkflowDefinitionStatusDraft,
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).
		Return(errors.New("database error"))

	_, err := uc.Activate(context.Background(), defID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save workflow definition")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// List Tests
// ============================================================================

func TestList_ByStatus_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	status := entity.WorkflowDefinitionStatusActive
	limit, offset := 10, 0

	definitions := []*entity.WorkflowDefinition{
		{ID: uuidv7.New(), Name: "workflow1", Status: status},
		{ID: uuidv7.New(), Name: "workflow2", Status: status},
	}

	mockRepo.On("ListByStatus", mock.Anything, status, limit, offset).Return(definitions, nil)
	mockRepo.On("CountByStatus", mock.Anything, status).Return(int64(2), nil)

	result, total, err := uc.List(context.Background(), &status, nil, nil, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestList_ByCategory_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	category := "automation"
	limit, offset := 10, 0

	definitions := []*entity.WorkflowDefinition{
		{ID: uuidv7.New(), Name: "workflow1", Category: category},
	}

	mockRepo.On("ListByCategory", mock.Anything, category, limit, offset).Return(definitions, nil)
	mockRepo.On("CountByCategory", mock.Anything, category).Return(int64(1), nil)

	result, total, err := uc.List(context.Background(), nil, &category, nil, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
	mockRepo.AssertExpectations(t)
}

func TestList_DefaultsToActive(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	limit, offset := 10, 0
	activeStatus := entity.WorkflowDefinitionStatusActive

	definitions := []*entity.WorkflowDefinition{
		{ID: uuidv7.New(), Name: "workflow1", Status: activeStatus},
	}

	mockRepo.On("ListByStatus", mock.Anything, activeStatus, limit, offset).Return(definitions, nil)
	mockRepo.On("CountByStatus", mock.Anything, activeStatus).Return(int64(1), nil)

	result, total, err := uc.List(context.Background(), nil, nil, nil, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Update Tests
// ============================================================================

func TestUpdate_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:          defID,
		Name:        "old_name",
		Description: "old description",
		Status:      entity.WorkflowDefinitionStatusDraft,
	}

	newName := "new_name"
	newDescription := "new description"
	newSchema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "new_state", Type: "start"}},
		InitialState: "new_state",
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).
		Run(func(args mock.Arguments) {
			updatedDef := args.Get(1).(*entity.WorkflowDefinition)
			assert.Equal(t, newName, updatedDef.Name)
			assert.Equal(t, newDescription, updatedDef.Description)
		}).
		Return(nil)

	result, err := uc.Update(context.Background(), defID, &newName, &newDescription, &newSchema, []string{"tag1"})

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newName, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	mockRepo.On("GetByID", mock.Anything, defID).Return(nil, errors.New("not found"))

	_, err := uc.Update(context.Background(), defID, nil, nil, nil, nil)

	assert.Equal(t, ErrWorkflowNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_NotDraft_Fails(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test",
		Status: entity.WorkflowDefinitionStatusActive, // Active, not draft
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)

	newName := "updated"
	_, err := uc.Update(context.Background(), defID, &newName, nil, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only draft definitions can be updated")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Delete Tests
// ============================================================================

func TestDelete_Success(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test",
		Status: entity.WorkflowDefinitionStatusDraft,
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockRepo.On("Delete", mock.Anything, defID).Return(nil)

	err := uc.Delete(context.Background(), defID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	mockRepo.On("GetByID", mock.Anything, defID).Return(nil, errors.New("not found"))

	err := uc.Delete(context.Background(), defID)

	assert.Equal(t, ErrWorkflowNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_NotDraft_Fails(t *testing.T) {
	mockRepo := new(MockWorkflowDefinitionRepository)
	uc := NewWorkflowDefinitionUseCase(mockRepo, nil)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test",
		Status: entity.WorkflowDefinitionStatusActive, // Active, not draft
	}

	mockRepo.On("GetByID", mock.Anything, defID).Return(def, nil)

	err := uc.Delete(context.Background(), defID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only draft definitions can be deleted")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// Deprecate Tests (Phase 7.4)
// ============================================================================

func TestDeprecate_Success(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusActive,
	}

	// Mock expectations
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockInstRepo.On("CountActiveByDefinition", mock.Anything, defID).Return(int64(0), nil)
	mockDefRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).Return(nil)

	result, err := uc.Deprecate(context.Background(), defID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowDefinitionStatusDeprecated, result.Status)
	mockDefRepo.AssertExpectations(t)
	mockInstRepo.AssertExpectations(t)
}

func TestDeprecate_NotFound(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(nil, errors.New("not found"))

	result, err := uc.Deprecate(context.Background(), defID)

	assert.Error(t, err)
	assert.Equal(t, ErrWorkflowNotFound, err)
	assert.Nil(t, result)
	mockDefRepo.AssertExpectations(t)
}

func TestDeprecate_HasActiveInstances(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusActive,
	}

	// Mock expectations - 5 active instances exist
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockInstRepo.On("CountActiveByDefinition", mock.Anything, defID).Return(int64(5), nil)

	result, err := uc.Deprecate(context.Background(), defID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "5 active instances exist")
	assert.Nil(t, result)
	mockDefRepo.AssertExpectations(t)
	mockInstRepo.AssertExpectations(t)
}

func TestDeprecate_NotActive(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusDraft, // Not active
	}

	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)

	result, err := uc.Deprecate(context.Background(), defID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only active definitions can be deprecated")
	assert.Nil(t, result)
	mockDefRepo.AssertExpectations(t)
}

// ============================================================================
// Archive Tests (Phase 7.5)
// ============================================================================

func TestArchive_Success_NoInstances(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusActive,
	}

	// Mock expectations - no instances
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockInstRepo.On("ListByDefinition", mock.Anything, defID, 1000, 0).Return([]*entity.WorkflowInstance{}, nil)
	mockDefRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).Return(nil)

	result, err := uc.Archive(context.Background(), defID, false)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowDefinitionStatusArchived, result.Status)
	mockDefRepo.AssertExpectations(t)
	mockInstRepo.AssertExpectations(t)
}

func TestArchive_Success_WithCancelInstances(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusActive,
	}

	// Create 2 running instances
	instance1 := entity.NewWorkflowInstance(defID, 1, "state1", jsonb.Map{}, uuidv7.New())
	instance1.Start()
	instance2 := entity.NewWorkflowInstance(defID, 1, "state2", jsonb.Map{}, uuidv7.New())
	instance2.Start()

	runningInstances := []*entity.WorkflowInstance{instance1, instance2}

	// Mock expectations - has running instances, cancel them
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockInstRepo.On("ListByDefinition", mock.Anything, defID, 1000, 0).Return(runningInstances, nil)
	mockInstRepo.On("Update", mock.Anything, instance1).Return(nil)
	mockInstRepo.On("Update", mock.Anything, instance2).Return(nil)
	mockDefRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkflowDefinition")).Return(nil)

	result, err := uc.Archive(context.Background(), defID, true)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entity.WorkflowDefinitionStatusArchived, result.Status)
	// Verify instances were cancelled
	assert.Equal(t, entity.WorkflowInstanceStatusCancelled, instance1.Status)
	assert.Equal(t, entity.WorkflowInstanceStatusCancelled, instance2.Status)
	mockDefRepo.AssertExpectations(t)
	mockInstRepo.AssertExpectations(t)
}

func TestArchive_Fails_HasRunningInstances_NoCancelFlag(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusActive,
	}

	// Create 3 running instances
	instance1 := entity.NewWorkflowInstance(defID, 1, "state1", jsonb.Map{}, uuidv7.New())
	instance1.Start()
	instance2 := entity.NewWorkflowInstance(defID, 1, "state2", jsonb.Map{}, uuidv7.New())
	instance2.Start()
	instance3 := entity.NewWorkflowInstance(defID, 1, "state3", jsonb.Map{}, uuidv7.New())
	instance3.Start()

	runningInstances := []*entity.WorkflowInstance{instance1, instance2, instance3}

	// Mock expectations - has running instances, but cancelRunningInstances=false
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)
	mockInstRepo.On("ListByDefinition", mock.Anything, defID, 1000, 0).Return(runningInstances, nil)

	result, err := uc.Archive(context.Background(), defID, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "3 running instances exist")
	assert.Contains(t, err.Error(), "use cancelRunningInstances=true")
	assert.Nil(t, result)
	mockDefRepo.AssertExpectations(t)
	mockInstRepo.AssertExpectations(t)
}

func TestArchive_NotFound(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	mockDefRepo.On("GetByID", mock.Anything, defID).Return(nil, errors.New("not found"))

	result, err := uc.Archive(context.Background(), defID, false)

	assert.Error(t, err)
	assert.Equal(t, ErrWorkflowNotFound, err)
	assert.Nil(t, result)
	mockDefRepo.AssertExpectations(t)
}

func TestArchive_AlreadyArchived(t *testing.T) {
	mockDefRepo := new(MockWorkflowDefinitionRepository)
	mockInstRepo := new(MockWorkflowInstanceRepository)
	uc := NewWorkflowDefinitionUseCase(mockDefRepo, mockInstRepo)

	defID := uuidv7.New()
	def := &entity.WorkflowDefinition{
		ID:     defID,
		Name:   "test_workflow",
		Status: entity.WorkflowDefinitionStatusArchived, // Already archived
	}

	mockDefRepo.On("GetByID", mock.Anything, defID).Return(def, nil)

	result, err := uc.Archive(context.Background(), defID, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already archived")
	assert.Nil(t, result)
	mockDefRepo.AssertExpectations(t)
}
