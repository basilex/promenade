package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockCostCenterRepository struct {
	CreateFunc             func(ctx context.Context, cc *aggregate.CostCenter) error
	GetByIDFunc            func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error)
	GetByCodeFunc          func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.CostCenter, error)
	UpdateFunc             func(ctx context.Context, cc *aggregate.CostCenter) error
	DeleteFunc             func(ctx context.Context, id uuidv7.UUID) error
	ListByOrganizationFunc func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error)
	ListChildrenFunc       func(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error)
	ListByTypeFunc         func(ctx context.Context, orgID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error)
	ListActiveFunc         func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.CostCenter, error)
}

func (m *MockCostCenterRepository) Create(ctx context.Context, cc *aggregate.CostCenter) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, cc)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockCostCenterRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockCostCenterRepository) GetByCode(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, orgID, code)
	}
	return nil, errors.New("GetByCodeFunc not implemented")
}

func (m *MockCostCenterRepository) Update(ctx context.Context, cc *aggregate.CostCenter) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, cc)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockCostCenterRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockCostCenterRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, orgID, limit, offset)
	}
	return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockCostCenterRepository) ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
	if m.ListChildrenFunc != nil {
		return m.ListChildrenFunc(ctx, parentID)
	}
	return nil, errors.New("ListChildrenFunc not implemented")
}

func (m *MockCostCenterRepository) ListByType(ctx context.Context, orgID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error) {
	if m.ListByTypeFunc != nil {
		return m.ListByTypeFunc(ctx, orgID, centerType)
	}
	return nil, errors.New("ListByTypeFunc not implemented")
}

func (m *MockCostCenterRepository) ListActive(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
	if m.ListActiveFunc != nil {
		return m.ListActiveFunc(ctx, orgID)
	}
	return nil, errors.New("ListActiveFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateCostCenter_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	mockRepo := &MockCostCenterRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
			return nil, costcenter.ErrCostCenterNotFound
		},
		CreateFunc: func(ctx context.Context, cc *aggregate.CostCenter) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	cc, err := uc.CreateCostCenter(context.Background(), orgID, "CC001", "Sales Department", aggregate.CenterTypeCost, nil, userID)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, cc.ID)
	assert.Equal(t, "CC001", cc.Code)
	assert.Equal(t, "Sales Department", cc.Name)
	assert.Equal(t, aggregate.CenterTypeCost, cc.CenterType)
	assert.True(t, cc.IsActive)
	assert.Equal(t, 1, cc.Level)
}

func TestCreateCostCenter_DuplicateCode(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	existing, _ := aggregate.NewCostCenter(orgID, "CC001", "Existing", aggregate.CenterTypeCost, userID)

	mockRepo := &MockCostCenterRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
			return existing, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	_, err := uc.CreateCostCenter(context.Background(), orgID, "CC001", "Duplicate", aggregate.CenterTypeCost, nil, userID)
	assert.Error(t, err)
	assert.Equal(t, costcenter.ErrCodeAlreadyExists, err)
}

func TestCreateCostCenter_WithParent(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()

	parent, _ := aggregate.NewCostCenter(orgID, "CC000", "Parent Dept", aggregate.CenterTypeCost, userID)
	parent.ID = parentID

	mockRepo := &MockCostCenterRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
			return nil, costcenter.ErrCostCenterNotFound
		},
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == parentID {
				return parent, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		CreateFunc: func(ctx context.Context, cc *aggregate.CostCenter) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	child, err := uc.CreateCostCenter(context.Background(), orgID, "CC001", "Child Dept", aggregate.CenterTypeCost, &parentID, userID)
	require.NoError(t, err)
	assert.Equal(t, parentID, *child.ParentID)
	assert.Equal(t, 2, child.Level)
}

func TestGetCostCenterByID_Success(t *testing.T) {
	cc, _ := aggregate.NewCostCenter(uuidv7.New(), "CC001", "Test Center", aggregate.CenterTypeCost, uuidv7.New())
	cc.ID = uuidv7.New()

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == cc.ID {
				return cc, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	retrieved, err := uc.GetCostCenterByID(context.Background(), cc.ID)
	require.NoError(t, err)
	assert.Equal(t, cc.ID, retrieved.ID)
	assert.Equal(t, "Test Center", retrieved.Name)
}

func TestGetCostCenterByID_NotFound(t *testing.T) {
	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			return nil, costcenter.ErrCostCenterNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	_, err := uc.GetCostCenterByID(context.Background(), uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, costcenter.ErrCostCenterNotFound, err)
}

func TestGetCostCenterByCode_Success(t *testing.T) {
	orgID := uuidv7.New()
	cc, _ := aggregate.NewCostCenter(orgID, "CC001", "Test Center", aggregate.CenterTypeCost, uuidv7.New())

	mockRepo := &MockCostCenterRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
			if code == "CC001" {
				return cc, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	retrieved, err := uc.GetCostCenterByCode(context.Background(), orgID, "CC001")
	require.NoError(t, err)
	assert.Equal(t, "CC001", retrieved.Code)
	assert.Equal(t, "Test Center", retrieved.Name)
}

func TestSetParent_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()
	childID := uuidv7.New()

	parent, _ := aggregate.NewCostCenter(orgID, "CC000", "Parent", aggregate.CenterTypeCost, userID)
	parent.ID = parentID

	child, _ := aggregate.NewCostCenter(orgID, "CC001", "Child", aggregate.CenterTypeCost, userID)
	child.ID = childID

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == childID {
				return child, nil
			}
			if id == parentID {
				return parent, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CostCenter) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	updated, err := uc.SetParent(context.Background(), childID, parentID, userID)
	require.NoError(t, err)
	assert.NotNil(t, updated.ParentID)
	assert.Equal(t, parentID, *updated.ParentID)
	assert.Equal(t, 2, updated.Level)
}

func TestSetManager_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	managerID := uuidv7.New()

	cc, _ := aggregate.NewCostCenter(orgID, "CC001", "Test Center", aggregate.CenterTypeCost, userID)
	cc.ID = uuidv7.New()

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == cc.ID {
				return cc, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CostCenter) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	updated, err := uc.SetManager(context.Background(), cc.ID, managerID, userID)
	require.NoError(t, err)
	assert.NotNil(t, updated.ManagerID)
	assert.Equal(t, managerID, *updated.ManagerID)
}

func TestActivateCostCenter_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc, _ := aggregate.NewCostCenter(orgID, "CC001", "Test Center", aggregate.CenterTypeCost, userID)
	cc.ID = uuidv7.New()
	cc.Deactivate()

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == cc.ID {
				return cc, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CostCenter) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	activated, err := uc.ActivateCostCenter(context.Background(), cc.ID, userID)
	require.NoError(t, err)
	assert.True(t, activated.IsActive)
}

func TestDeactivateCostCenter_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc, _ := aggregate.NewCostCenter(orgID, "CC001", "Test Center", aggregate.CenterTypeCost, userID)
	cc.ID = uuidv7.New()

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == cc.ID {
				return cc, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CostCenter) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	deactivated, err := uc.DeactivateCostCenter(context.Background(), cc.ID, userID)
	require.NoError(t, err)
	assert.False(t, deactivated.IsActive)
}

func TestDeleteCostCenter_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc, _ := aggregate.NewCostCenter(orgID, "CC001", "Test Center", aggregate.CenterTypeCost, userID)
	cc.ID = uuidv7.New()

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == cc.ID {
				return cc, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		ListChildrenFunc: func(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
			return []*aggregate.CostCenter{}, nil
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	err := uc.DeleteCostCenter(context.Background(), cc.ID)
	require.NoError(t, err)
}

func TestDeleteCostCenter_WithChildren(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	parent, _ := aggregate.NewCostCenter(orgID, "CC000", "Parent", aggregate.CenterTypeCost, userID)
	parent.ID = uuidv7.New()

	child, _ := aggregate.NewCostCenter(orgID, "CC001", "Child", aggregate.CenterTypeCost, userID)
	child.ParentID = &parent.ID

	mockRepo := &MockCostCenterRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
			if id == parent.ID {
				return parent, nil
			}
			return nil, costcenter.ErrCostCenterNotFound
		},
		ListChildrenFunc: func(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
			if parentID == parent.ID {
				return []*aggregate.CostCenter{child}, nil
			}
			return []*aggregate.CostCenter{}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	err := uc.DeleteCostCenter(context.Background(), parent.ID)
	assert.Error(t, err)
	assert.Equal(t, costcenter.ErrHasChildren, err)
}

func TestListCostCentersByOrganization_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc1, _ := aggregate.NewCostCenter(orgID, "CC001", "Center 1", aggregate.CenterTypeCost, userID)
	cc2, _ := aggregate.NewCostCenter(orgID, "CC002", "Center 2", aggregate.CenterTypeProfit, userID)

	mockRepo := &MockCostCenterRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error) {
			return []*aggregate.CostCenter{cc1, cc2}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	centers, err := uc.ListCostCentersByOrganization(context.Background(), orgID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, centers, 2)
}

func TestListChildCostCenters_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()

	child1, _ := aggregate.NewChildCostCenter(orgID, "CC001", "Child 1", aggregate.CenterTypeCost, parentID, 1, userID)
	child2, _ := aggregate.NewChildCostCenter(orgID, "CC002", "Child 2", aggregate.CenterTypeCost, parentID, 1, userID)

	mockRepo := &MockCostCenterRepository{
		ListChildrenFunc: func(ctx context.Context, pID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
			if pID == parentID {
				return []*aggregate.CostCenter{child1, child2}, nil
			}
			return []*aggregate.CostCenter{}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	children, err := uc.ListChildCostCenters(context.Background(), parentID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	for _, child := range children {
		assert.Equal(t, parentID, *child.ParentID)
	}
}

func TestListCostCentersByType_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc1, _ := aggregate.NewCostCenter(orgID, "CC001", "Dept 1", aggregate.CenterTypeCost, userID)
	cc2, _ := aggregate.NewCostCenter(orgID, "CC002", "Dept 2", aggregate.CenterTypeCost, userID)

	mockRepo := &MockCostCenterRepository{
		ListByTypeFunc: func(ctx context.Context, orgID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error) {
			if centerType == aggregate.CenterTypeCost {
				return []*aggregate.CostCenter{cc1, cc2}, nil
			}
			return []*aggregate.CostCenter{}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	centers, err := uc.ListCostCentersByType(context.Background(), orgID, aggregate.CenterTypeCost)
	require.NoError(t, err)
	assert.Len(t, centers, 2)
	for _, cc := range centers {
		assert.Equal(t, aggregate.CenterTypeCost, cc.CenterType)
	}
}

func TestListActiveCostCenters_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc1, _ := aggregate.NewCostCenter(orgID, "CC001", "Active 1", aggregate.CenterTypeCost, userID)
	cc2, _ := aggregate.NewCostCenter(orgID, "CC002", "Active 2", aggregate.CenterTypeCost, userID)

	mockRepo := &MockCostCenterRepository{
		ListActiveFunc: func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
			return []*aggregate.CostCenter{cc1, cc2}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewCostCenterUseCase(mockRepo, auditLogger)

	activeCenters, err := uc.ListActiveCostCenters(context.Background(), orgID)
	require.NoError(t, err)
	assert.Len(t, activeCenters, 2)
	for _, cc := range activeCenters {
		assert.True(t, cc.IsActive)
	}
}
