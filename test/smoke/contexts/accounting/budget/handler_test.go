package budget_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/basilex/promenade/internal/contexts/accounting/budget"
	budgetHTTP "github.com/basilex/promenade/internal/contexts/accounting/budget/adapter/http"
	budgetAggregate "github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockBudgetUseCase implements usecase.IBudgetUseCase for testing
type MockBudgetUseCase struct {
	CreateBudgetFunc              func(ctx context.Context, organizationID uuidv7.UUID, name string, fiscalYear int, createdBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	AddLineFunc                   func(ctx context.Context, budgetID, accountID uuidv7.UUID, budgetAmount int64, description string, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	UpdateLineFunc                func(ctx context.Context, budgetID, lineID uuidv7.UUID, budgetAmount int64, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	RemoveLineFunc                func(ctx context.Context, budgetID, lineID uuidv7.UUID, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	ApproveBudgetFunc             func(ctx context.Context, budgetID, approvedBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	ActivateBudgetFunc            func(ctx context.Context, budgetID, activatedBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	CloseBudgetFunc               func(ctx context.Context, budgetID, closedBy uuidv7.UUID) (*budgetAggregate.Budget, error)
	GetBudgetByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*budgetAggregate.Budget, error)
	GetBudgetByNameFunc           func(ctx context.Context, organizationID uuidv7.UUID, name string) (*budgetAggregate.Budget, error)
	DeleteBudgetFunc              func(ctx context.Context, id uuidv7.UUID) error
	ListBudgetsByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*budgetAggregate.Budget, error)
	ListBudgetsByStatusFunc       func(ctx context.Context, organizationID uuidv7.UUID, status budgetAggregate.BudgetStatus) ([]*budgetAggregate.Budget, error)
	ListActiveBudgetsFunc         func(ctx context.Context, organizationID uuidv7.UUID) ([]*budgetAggregate.Budget, error)
}

func (m *MockBudgetUseCase) CreateBudget(ctx context.Context, organizationID uuidv7.UUID, name string, fiscalYear int, createdBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.CreateBudgetFunc != nil {
		return m.CreateBudgetFunc(ctx, organizationID, name, fiscalYear, createdBy)
	}
	return nil, fmt.Errorf("CreateBudgetFunc not implemented")
}

func (m *MockBudgetUseCase) AddLine(ctx context.Context, budgetID, accountID uuidv7.UUID, budgetAmount int64, description string, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.AddLineFunc != nil {
		return m.AddLineFunc(ctx, budgetID, accountID, budgetAmount, description, updatedBy)
	}
	return nil, fmt.Errorf("AddLineFunc not implemented")
}

func (m *MockBudgetUseCase) UpdateLine(ctx context.Context, budgetID, lineID uuidv7.UUID, budgetAmount int64, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.UpdateLineFunc != nil {
		return m.UpdateLineFunc(ctx, budgetID, lineID, budgetAmount, updatedBy)
	}
	return nil, fmt.Errorf("UpdateLineFunc not implemented")
}

func (m *MockBudgetUseCase) RemoveLine(ctx context.Context, budgetID, lineID uuidv7.UUID, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.RemoveLineFunc != nil {
		return m.RemoveLineFunc(ctx, budgetID, lineID, updatedBy)
	}
	return nil, fmt.Errorf("RemoveLineFunc not implemented")
}

func (m *MockBudgetUseCase) ApproveBudget(ctx context.Context, budgetID, approvedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.ApproveBudgetFunc != nil {
		return m.ApproveBudgetFunc(ctx, budgetID, approvedBy)
	}
	return nil, fmt.Errorf("ApproveBudgetFunc not implemented")
}

func (m *MockBudgetUseCase) ActivateBudget(ctx context.Context, budgetID, activatedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.ActivateBudgetFunc != nil {
		return m.ActivateBudgetFunc(ctx, budgetID, activatedBy)
	}
	return nil, fmt.Errorf("ActivateBudgetFunc not implemented")
}

func (m *MockBudgetUseCase) CloseBudget(ctx context.Context, budgetID, closedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.CloseBudgetFunc != nil {
		return m.CloseBudgetFunc(ctx, budgetID, closedBy)
	}
	return nil, fmt.Errorf("CloseBudgetFunc not implemented")
}

func (m *MockBudgetUseCase) GetBudgetByID(ctx context.Context, id uuidv7.UUID) (*budgetAggregate.Budget, error) {
	if m.GetBudgetByIDFunc != nil {
		return m.GetBudgetByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetBudgetByIDFunc not implemented")
}

func (m *MockBudgetUseCase) GetBudgetByName(ctx context.Context, organizationID uuidv7.UUID, name string) (*budgetAggregate.Budget, error) {
	if m.GetBudgetByNameFunc != nil {
		return m.GetBudgetByNameFunc(ctx, organizationID, name)
	}
	return nil, fmt.Errorf("GetBudgetByNameFunc not implemented")
}

func (m *MockBudgetUseCase) DeleteBudget(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteBudgetFunc != nil {
		return m.DeleteBudgetFunc(ctx, id)
	}
	return fmt.Errorf("DeleteBudgetFunc not implemented")
}

func (m *MockBudgetUseCase) ListBudgetsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*budgetAggregate.Budget, error) {
	if m.ListBudgetsByOrganizationFunc != nil {
		return m.ListBudgetsByOrganizationFunc(ctx, organizationID, limit, offset)
	}
	return nil, fmt.Errorf("ListBudgetsByOrganizationFunc not implemented")
}

func (m *MockBudgetUseCase) ListBudgetsByStatus(ctx context.Context, organizationID uuidv7.UUID, status budgetAggregate.BudgetStatus) ([]*budgetAggregate.Budget, error) {
	if m.ListBudgetsByStatusFunc != nil {
		return m.ListBudgetsByStatusFunc(ctx, organizationID, status)
	}
	return nil, fmt.Errorf("ListBudgetsByStatusFunc not implemented")
}

func (m *MockBudgetUseCase) ListActiveBudgets(ctx context.Context, organizationID uuidv7.UUID) ([]*budgetAggregate.Budget, error) {
	if m.ListActiveBudgetsFunc != nil {
		return m.ListActiveBudgetsFunc(ctx, organizationID)
	}
	return nil, fmt.Errorf("ListActiveBudgetsFunc not implemented")
}

func setupRouter(mockUC *MockBudgetUseCase) *gin.Engine {
	router := smoke.SetupRouter()
	handler := budgetHTTP.NewBudgetHandler(mockUC)

	router.Use(func(c *gin.Context) {
		c.Set("organization_id", uuidv7.New().String())
		c.Set("user_id", uuidv7.New().String())
		c.Next()
	})

	handler.RegisterRoutes(router.Group("/api/v1/accounting"))
	return router
}

func TestCreateBudget_Success(t *testing.T) {
	budgetID := uuidv7.New()

	mockUC := &MockBudgetUseCase{
		CreateBudgetFunc: func(ctx context.Context, organizationID uuidv7.UUID, name string, fiscalYear int, createdBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
			b, _ := budgetAggregate.NewBudget(organizationID, name, fiscalYear, createdBy)
			b.ID = budgetID
			return b, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"name":        "FY2024 Budget",
		"fiscal_year": 2024,
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/budgets", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateBudget_ValidationError(t *testing.T) {
	mockUC := &MockBudgetUseCase{
		CreateBudgetFunc: func(ctx context.Context, organizationID uuidv7.UUID, name string, fiscalYear int, createdBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
			return nil, budget.ErrBudgetNameEmpty
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"name":        "",
		"fiscal_year": 2024,
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/budgets", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	budgetID := uuidv7.New()

	b, _ := budgetAggregate.NewBudget(orgID, "FY2024 Budget", 2024, userID)
	b.ID = budgetID

	mockUC := &MockBudgetUseCase{
		GetBudgetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*budgetAggregate.Budget, error) {
			return b, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/budgets/"+budgetID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetBudget_NotFound(t *testing.T) {
	budgetID := uuidv7.New()

	mockUC := &MockBudgetUseCase{
		GetBudgetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*budgetAggregate.Budget, error) {
			return nil, budget.ErrBudgetNotFound
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/budgets/"+budgetID.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddBudgetLine_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	budgetID := uuidv7.New()
	accountID := uuidv7.New()

	b, _ := budgetAggregate.NewBudget(orgID, "FY2024 Budget", 2024, userID)
	b.ID = budgetID

	mockUC := &MockBudgetUseCase{
		AddLineFunc: func(ctx context.Context, bID, accID uuidv7.UUID, budgetAmount int64, description string, updatedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
			return b, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"account_id":    accountID.String(),
		"budget_amount": 100000,
		"description":   "Marketing budget",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/budgets/"+budgetID.String()+"/lines", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestApproveBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	budgetID := uuidv7.New()

	b, _ := budgetAggregate.NewBudget(orgID, "FY2024 Budget", 2024, userID)
	b.ID = budgetID

	mockUC := &MockBudgetUseCase{
		ApproveBudgetFunc: func(ctx context.Context, bID, approvedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
			return b, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/budgets/"+budgetID.String()+"/approve", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestActivateBudget_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	budgetID := uuidv7.New()

	b, _ := budgetAggregate.NewBudget(orgID, "FY2024 Budget", 2024, userID)
	b.ID = budgetID

	mockUC := &MockBudgetUseCase{
		ActivateBudgetFunc: func(ctx context.Context, bID, activatedBy uuidv7.UUID) (*budgetAggregate.Budget, error) {
			return b, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/budgets/"+budgetID.String()+"/activate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListBudgets_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	budgets := []*budgetAggregate.Budget{
		func() *budgetAggregate.Budget {
			b, _ := budgetAggregate.NewBudget(orgID, "FY2024", 2024, userID)
			return b
		}(),
		func() *budgetAggregate.Budget {
			b, _ := budgetAggregate.NewBudget(orgID, "FY2025", 2025, userID)
			return b
		}(),
	}

	mockUC := &MockBudgetUseCase{
		ListBudgetsByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*budgetAggregate.Budget, error) {
			return budgets, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/budgets", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}
