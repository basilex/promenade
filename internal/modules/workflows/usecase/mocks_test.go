package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/mock"
)

// MockWorkflowDefinitionRepository is a mock for testing
type MockWorkflowDefinitionRepository struct {
	mock.Mock
}

func (m *MockWorkflowDefinitionRepository) Create(ctx context.Context, def *entity.WorkflowDefinition) error {
	args := m.Called(ctx, def)
	return args.Error(0)
}

func (m *MockWorkflowDefinitionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) GetByName(ctx context.Context, name string) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) Update(ctx context.Context, def *entity.WorkflowDefinition) error {
	args := m.Called(ctx, def)
	return args.Error(0)
}

func (m *MockWorkflowDefinitionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkflowDefinitionRepository) GetByNameAndVersion(ctx context.Context, name string, version int) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, name, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) ListByStatus(ctx context.Context, status entity.WorkflowDefinitionStatus, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, category, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) ListByCreator(ctx context.Context, creatorID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, creatorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) ListVersions(ctx context.Context, name string) ([]*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) CountByStatus(ctx context.Context, status entity.WorkflowDefinitionStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	args := m.Called(ctx, category)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWorkflowDefinitionRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowDefinition), args.Error(1)
}

// MockWorkflowInstanceRepository is a mock for testing
type MockWorkflowInstanceRepository struct {
	mock.Mock
}

func (m *MockWorkflowInstanceRepository) Create(ctx context.Context, instance *entity.WorkflowInstance) error {
	args := m.Called(ctx, instance)
	return args.Error(0)
}

func (m *MockWorkflowInstanceRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) Update(ctx context.Context, instance *entity.WorkflowInstance) error {
	args := m.Called(ctx, instance)
	return args.Error(0)
}

func (m *MockWorkflowInstanceRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkflowInstanceRepository) GetByExternalReference(ctx context.Context, externalRef string) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, externalRef)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListByDefinition(ctx context.Context, definitionID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, definitionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListByStatus(ctx context.Context, status entity.WorkflowInstanceStatus, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListByAssignee(ctx context.Context, assigneeID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, assigneeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListByCreator(ctx context.Context, creatorID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, creatorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListOverdue(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListTimedOut(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) ListActive(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) CountByStatus(ctx context.Context, status entity.WorkflowInstanceStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) CountByDefinition(ctx context.Context, definitionID uuidv7.UUID) (int64, error) {
	args := m.Called(ctx, definitionID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) CountActiveByDefinition(ctx context.Context, definitionID uuidv7.UUID) (int64, error) {
	args := m.Called(ctx, definitionID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) CountByAssignee(ctx context.Context, assigneeID uuidv7.UUID) (int64, error) {
	args := m.Called(ctx, assigneeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWorkflowInstanceRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
