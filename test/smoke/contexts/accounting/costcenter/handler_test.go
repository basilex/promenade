package costcenter_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/basilex/promenade/internal/contexts/accounting/costcenter"
	costcenterHTTP "github.com/basilex/promenade/internal/contexts/accounting/costcenter/adapter/http"
	costcenterAggregate "github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockCostCenterUseCase implements usecase.ICostCenterUseCase for testing
type MockCostCenterUseCase struct {
	CreateCostCenterFunc              func(ctx context.Context, organizationID uuidv7.UUID, code, name string, centerType costcenterAggregate.CenterType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error)
	GetCostCenterByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*costcenterAggregate.CostCenter, error)
	GetCostCenterByCodeFunc           func(ctx context.Context, organizationID uuidv7.UUID, code string) (*costcenterAggregate.CostCenter, error)
	SetParentFunc                     func(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error)
	SetManagerFunc                    func(ctx context.Context, id, managerID uuidv7.UUID, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error)
	ActivateCostCenterFunc            func(ctx context.Context, id, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error)
	DeactivateCostCenterFunc          func(ctx context.Context, id, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error)
	DeleteCostCenterFunc              func(ctx context.Context, id uuidv7.UUID) error
	ListCostCentersByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*costcenterAggregate.CostCenter, error)
	ListChildCostCentersFunc          func(ctx context.Context, parentID uuidv7.UUID) ([]*costcenterAggregate.CostCenter, error)
	ListCostCentersByTypeFunc         func(ctx context.Context, organizationID uuidv7.UUID, centerType costcenterAggregate.CenterType) ([]*costcenterAggregate.CostCenter, error)
	ListActiveCostCentersFunc         func(ctx context.Context, organizationID uuidv7.UUID) ([]*costcenterAggregate.CostCenter, error)
}

func (m *MockCostCenterUseCase) CreateCostCenter(ctx context.Context, organizationID uuidv7.UUID, code, name string, centerType costcenterAggregate.CenterType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
	if m.CreateCostCenterFunc != nil {
		return m.CreateCostCenterFunc(ctx, organizationID, code, name, centerType, parentID, createdBy)
	}
	return nil, fmt.Errorf("CreateCostCenterFunc not implemented")
}

func (m *MockCostCenterUseCase) GetCostCenterByID(ctx context.Context, id uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
	if m.GetCostCenterByIDFunc != nil {
		return m.GetCostCenterByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetCostCenterByIDFunc not implemented")
}

func (m *MockCostCenterUseCase) GetCostCenterByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*costcenterAggregate.CostCenter, error) {
	if m.GetCostCenterByCodeFunc != nil {
		return m.GetCostCenterByCodeFunc(ctx, organizationID, code)
	}
	return nil, fmt.Errorf("GetCostCenterByCodeFunc not implemented")
}

func (m *MockCostCenterUseCase) SetParent(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
	if m.SetParentFunc != nil {
		return m.SetParentFunc(ctx, id, parentID, updatedBy)
	}
	return nil, fmt.Errorf("SetParentFunc not implemented")
}

func (m *MockCostCenterUseCase) SetManager(ctx context.Context, id, managerID uuidv7.UUID, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
	if m.SetManagerFunc != nil {
		return m.SetManagerFunc(ctx, id, managerID, updatedBy)
	}
	return nil, fmt.Errorf("SetManagerFunc not implemented")
}

func (m *MockCostCenterUseCase) ActivateCostCenter(ctx context.Context, id, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
	if m.ActivateCostCenterFunc != nil {
		return m.ActivateCostCenterFunc(ctx, id, updatedBy)
	}
	return nil, fmt.Errorf("ActivateCostCenterFunc not implemented")
}

func (m *MockCostCenterUseCase) DeactivateCostCenter(ctx context.Context, id, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
	if m.DeactivateCostCenterFunc != nil {
		return m.DeactivateCostCenterFunc(ctx, id, updatedBy)
	}
	return nil, fmt.Errorf("DeactivateCostCenterFunc not implemented")
}

func (m *MockCostCenterUseCase) DeleteCostCenter(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteCostCenterFunc != nil {
		return m.DeleteCostCenterFunc(ctx, id)
	}
	return fmt.Errorf("DeleteCostCenterFunc not implemented")
}

func (m *MockCostCenterUseCase) ListCostCentersByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*costcenterAggregate.CostCenter, error) {
	if m.ListCostCentersByOrganizationFunc != nil {
		return m.ListCostCentersByOrganizationFunc(ctx, organizationID, limit, offset)
	}
	return nil, fmt.Errorf("ListCostCentersByOrganizationFunc not implemented")
}

func (m *MockCostCenterUseCase) ListChildCostCenters(ctx context.Context, parentID uuidv7.UUID) ([]*costcenterAggregate.CostCenter, error) {
	if m.ListChildCostCentersFunc != nil {
		return m.ListChildCostCentersFunc(ctx, parentID)
	}
	return nil, fmt.Errorf("ListChildCostCentersFunc not implemented")
}

func (m *MockCostCenterUseCase) ListCostCentersByType(ctx context.Context, organizationID uuidv7.UUID, centerType costcenterAggregate.CenterType) ([]*costcenterAggregate.CostCenter, error) {
	if m.ListCostCentersByTypeFunc != nil {
		return m.ListCostCentersByTypeFunc(ctx, organizationID, centerType)
	}
	return nil, fmt.Errorf("ListCostCentersByTypeFunc not implemented")
}

func (m *MockCostCenterUseCase) ListActiveCostCenters(ctx context.Context, organizationID uuidv7.UUID) ([]*costcenterAggregate.CostCenter, error) {
	if m.ListActiveCostCentersFunc != nil {
		return m.ListActiveCostCentersFunc(ctx, organizationID)
	}
	return nil, fmt.Errorf("ListActiveCostCentersFunc not implemented")
}

func setupRouter(mockUC *MockCostCenterUseCase) *gin.Engine {
	router := smoke.SetupRouter()
	handler := costcenterHTTP.NewCostCenterHandler(mockUC)

	router.Use(func(c *gin.Context) {
		c.Set("organization_id", uuidv7.New().String())
		c.Set("user_id", uuidv7.New().String())
		c.Next()
	})

	handler.RegisterRoutes(router.Group("/api/v1/accounting"))
	return router
}

func TestCreateCostCenter_Success(t *testing.T) {
	centerID := uuidv7.New()

	mockUC := &MockCostCenterUseCase{
		CreateCostCenterFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, centerType costcenterAggregate.CenterType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			cc, _ := costcenterAggregate.NewCostCenter(organizationID, code, name, centerType, createdBy)
			cc.ID = centerID
			return cc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":        "CC100",
		"name":        "Sales Department",
		"center_type": "cost_center",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/cost-centers", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateCostCenter_ValidationError(t *testing.T) {
	mockUC := &MockCostCenterUseCase{
		CreateCostCenterFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, centerType costcenterAggregate.CenterType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			return nil, costcenter.ErrCodeRequired
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":        "",
		"name":        "Sales Department",
		"center_type": "cost_center",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/cost-centers", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCostCenter_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	centerID := uuidv7.New()

	cc, _ := costcenterAggregate.NewCostCenter(orgID, "CC100", "Sales Department", costcenterAggregate.CenterTypeCost, userID)
	cc.ID = centerID

	mockUC := &MockCostCenterUseCase{
		GetCostCenterByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			return cc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/cost-centers/"+centerID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCostCenter_NotFound(t *testing.T) {
	centerID := uuidv7.New()

	mockUC := &MockCostCenterUseCase{
		GetCostCenterByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			return nil, costcenter.ErrCostCenterNotFound
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/cost-centers/"+centerID.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetManager_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	centerID := uuidv7.New()
	managerID := uuidv7.New()

	cc, _ := costcenterAggregate.NewCostCenter(orgID, "CC100", "Sales Department", costcenterAggregate.CenterTypeCost, userID)
	cc.ID = centerID

	mockUC := &MockCostCenterUseCase{
		SetManagerFunc: func(ctx context.Context, id, mgrID uuidv7.UUID, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			return cc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"manager_id": managerID.String(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/cost-centers/"+centerID.String()+"/manager", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestActivateCostCenter_Success(t *testing.T) {
	centerID := uuidv7.New()
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc, _ := costcenterAggregate.NewCostCenter(orgID, "CC100", "Sales Department", costcenterAggregate.CenterTypeCost, userID)
	cc.ID = centerID

	mockUC := &MockCostCenterUseCase{
		ActivateCostCenterFunc: func(ctx context.Context, id, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			return cc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/cost-centers/"+centerID.String()+"/activate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeactivateCostCenter_Success(t *testing.T) {
	centerID := uuidv7.New()
	orgID := uuidv7.New()
	userID := uuidv7.New()

	cc, _ := costcenterAggregate.NewCostCenter(orgID, "CC100", "Sales Department", costcenterAggregate.CenterTypeCost, userID)
	cc.ID = centerID

	mockUC := &MockCostCenterUseCase{
		DeactivateCostCenterFunc: func(ctx context.Context, id, updatedBy uuidv7.UUID) (*costcenterAggregate.CostCenter, error) {
			return cc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/cost-centers/"+centerID.String()+"/deactivate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListCostCenters_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	centers := []*costcenterAggregate.CostCenter{
		func() *costcenterAggregate.CostCenter {
			cc, _ := costcenterAggregate.NewCostCenter(orgID, "CC100", "Sales", costcenterAggregate.CenterTypeCost, userID)
			return cc
		}(),
		func() *costcenterAggregate.CostCenter {
			cc, _ := costcenterAggregate.NewCostCenter(orgID, "CC200", "Marketing", costcenterAggregate.CenterTypeProfit, userID)
			return cc
		}(),
	}

	mockUC := &MockCostCenterUseCase{
		ListCostCentersByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*costcenterAggregate.CostCenter, error) {
			return centers, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/cost-centers", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}
