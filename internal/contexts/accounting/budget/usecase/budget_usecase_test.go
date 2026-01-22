package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/budget"
	"github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockBudgetRepository struct {
	CreateFunc             func(ctx context.Context, bdg *aggregate.Budget) error
	GetByIDFunc            func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error)
	GetByNameFunc          func(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.Budget, error)
	UpdateFunc             func(ctx context.Context, bdg *aggregate.Budget) error
	DeleteFunc             func(ctx context.Context, id uuidv7.UUID) error
	ListByOrganizationFunc func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error)
	ListByStatusFunc       func(ctx context.Context, orgID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error)
	ListActiveFunc         func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.Budget, error)
}

func (m *MockBudgetRepository) Create(ctx context.Context, bdg *aggregate.Budget) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, bdg)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockBudgetRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockBudgetRepository) GetByName(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.Budget, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, orgID, name)
	}
	return nil, errors.New("GetByNameFunc not implemented")
}

func (m *MockBudgetRepository) Update(ctx context.Context, bdg *aggregate.Budget) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, bdg)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockBudgetRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockBudgetRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, orgID, limit, offset)
	}
	return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockBudgetRepository) ListByStatus(ctx context.Context, orgID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, orgID, status)
	}
	return nil, errors.New("ListByStatusFunc not implemented")
}

func (m *MockBudgetRepository) ListActive(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.Budget, error) {
	if m.ListActiveFunc != nil {
		return m.ListActiveFunc(ctx, orgID)
	}
	return nil, errors.New("ListActiveFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	mockRepo := &MockBudgetRepository{
		GetByNameFunc: func(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.Budget, error) {
			return nil, budget.ErrBudgetNotFound
		},
		CreateFunc: func(ctx context.Context, bdg *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	bdg, err := uc.CreateBudget(context.Background(), orgID, "Budget 2025", 2025, userID)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, bdg.ID)
	assert.Equal(t, "Budget 2025", bdg.Name)
	assert.Equal(t, 2025, bdg.FiscalYear)
	assert.Equal(t, aggregate.BudgetStatusDraft, bdg.Status)
}

func TestAddLine_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Test Budget", 2025, userID)
	bdg.ID = uuidv7.New()

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	updated, err := uc.AddLine(context.Background(), bdg.ID, accountID, 100000, "Marketing", userID)
	require.NoError(t, err)
	assert.Len(t, updated.Lines, 1)
	assert.Equal(t, accountID, updated.Lines[0].AccountID)
	assert.Equal(t, int64(100000), updated.Lines[0].BudgetAmount)
}

func TestAddLine_CannotModifyApproved(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Approved Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 50000, "Initial")
	_ = bdg.Approve(userID)

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	_, err := uc.AddLine(context.Background(), bdg.ID, accountID, 100000, "Cannot add", userID)
	assert.Error(t, err)
	assert.Equal(t, budget.ErrCannotModifyApprovedBudget, err)
}

func TestUpdateLine_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Test Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(accountID, 100000, "Initial")
	lineID := bdg.Lines[0].ID

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	updated, err := uc.UpdateLine(context.Background(), bdg.ID, lineID, 150000, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(150000), updated.Lines[0].BudgetAmount)
}

func TestRemoveLine_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Test Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 50000, "Line 1")
	_ = bdg.AddLine(uuidv7.New(), 75000, "Line 2")
	lineToRemove := bdg.Lines[0].ID

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	updated, err := uc.RemoveLine(context.Background(), bdg.ID, lineToRemove, userID)
	require.NoError(t, err)
	assert.Len(t, updated.Lines, 1)
}

func TestApproveBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	approverID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Test Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 100000, "Expense")

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	approved, err := uc.ApproveBudget(context.Background(), bdg.ID, approverID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.BudgetStatusApproved, approved.Status)
	assert.NotNil(t, approved.ApprovedAt)
	assert.NotNil(t, approved.ApprovedBy)
	assert.Equal(t, approverID, *approved.ApprovedBy)
}

func TestApproveBudget_EmptyBudget(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Empty Budget", 2025, userID)
	bdg.ID = uuidv7.New()

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	_, err := uc.ApproveBudget(context.Background(), bdg.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, budget.ErrBudgetHasNoLines, err)
}

func TestActivateBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Test Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 100000, "Expense")
	_ = bdg.Approve(userID)

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	activated, err := uc.ActivateBudget(context.Background(), bdg.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.BudgetStatusActive, activated.Status)
}

func TestActivateBudget_NotApproved(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Draft Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 100000, "Expense")

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	_, err := uc.ActivateBudget(context.Background(), bdg.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, budget.ErrBudgetNotApproved, err)
}

func TestCloseBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Test Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 100000, "Expense")
	_ = bdg.Approve(userID)
	_ = bdg.Activate()

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Budget) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	closed, err := uc.CloseBudget(context.Background(), bdg.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.BudgetStatusClosed, closed.Status)
}

func TestGetBudgetByID_Success(t *testing.T) {
	bdg, _ := aggregate.NewBudget(uuidv7.New(), "Test Budget", 2025, uuidv7.New())
	bdg.ID = uuidv7.New()

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	retrieved, err := uc.GetBudgetByID(context.Background(), bdg.ID)
	require.NoError(t, err)
	assert.Equal(t, bdg.ID, retrieved.ID)
	assert.Equal(t, "Test Budget", retrieved.Name)
}

func TestGetBudgetByID_NotFound(t *testing.T) {
	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			return nil, budget.ErrBudgetNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	_, err := uc.GetBudgetByID(context.Background(), uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, budget.ErrBudgetNotFound, err)
}

func TestGetBudgetByName_Success(t *testing.T) {
	orgID := uuidv7.New()
	bdg, _ := aggregate.NewBudget(orgID, "Unique Budget", 2025, uuidv7.New())

	mockRepo := &MockBudgetRepository{
		GetByNameFunc: func(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.Budget, error) {
			if name == "Unique Budget" {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	retrieved, err := uc.GetBudgetByName(context.Background(), orgID, "Unique Budget")
	require.NoError(t, err)
	assert.Equal(t, "Unique Budget", retrieved.Name)
}

func TestDeleteBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Budget to Delete", 2025, userID)
	bdg.ID = uuidv7.New()

	mockRepo := &MockBudgetRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
			if id == bdg.ID {
				return bdg, nil
			}
			return nil, budget.ErrBudgetNotFound
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	err := uc.DeleteBudget(context.Background(), bdg.ID)
	require.NoError(t, err)
}

func TestListBudgetsByOrganization_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg1, _ := aggregate.NewBudget(orgID, "Budget 1", 2025, userID)
	bdg2, _ := aggregate.NewBudget(orgID, "Budget 2", 2026, userID)

	mockRepo := &MockBudgetRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error) {
			return []*aggregate.Budget{bdg1, bdg2}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	budgets, err := uc.ListBudgetsByOrganization(context.Background(), orgID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, budgets, 2)
}

func TestListBudgetsByStatus_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	draftBdg, _ := aggregate.NewBudget(orgID, "Draft Budget", 2025, userID)
	approvedBdg, _ := aggregate.NewBudget(orgID, "Approved Budget", 2026, userID)
	_ = approvedBdg.AddLine(uuidv7.New(), 100000, "Line")
	_ = approvedBdg.Approve(userID)

	mockRepo := &MockBudgetRepository{
		ListByStatusFunc: func(ctx context.Context, orgID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error) {
			if status == aggregate.BudgetStatusDraft {
				return []*aggregate.Budget{draftBdg}, nil
			}
			if status == aggregate.BudgetStatusApproved {
				return []*aggregate.Budget{approvedBdg}, nil
			}
			return []*aggregate.Budget{}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	draftBudgets, err := uc.ListBudgetsByStatus(context.Background(), orgID, aggregate.BudgetStatusDraft)
	require.NoError(t, err)
	assert.Len(t, draftBudgets, 1)
	assert.Equal(t, aggregate.BudgetStatusDraft, draftBudgets[0].Status)
}

func TestListActiveBudgets_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	bdg, _ := aggregate.NewBudget(orgID, "Active Budget", 2025, userID)
	bdg.ID = uuidv7.New()
	_ = bdg.AddLine(uuidv7.New(), 100000, "Expense")
	_ = bdg.Approve(userID)
	_ = bdg.Activate()

	mockRepo := &MockBudgetRepository{
		ListActiveFunc: func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.Budget, error) {
			return []*aggregate.Budget{bdg}, nil
		},
	}

	auditLogger := audit.NewAuditLogger(nil)
	uc := NewBudgetUseCase(mockRepo, auditLogger)

	activeBudgets, err := uc.ListActiveBudgets(context.Background(), orgID)
	require.NoError(t, err)
	assert.Len(t, activeBudgets, 1)
	assert.Equal(t, aggregate.BudgetStatusActive, activeBudgets[0].Status)
}
