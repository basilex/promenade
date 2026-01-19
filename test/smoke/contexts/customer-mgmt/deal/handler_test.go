package deal_test

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal"
	dealHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/adapter/http"
	dealAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/smoke"
	"github.com/stretchr/testify/assert"
)

// MockDealUseCase implements minimal deal.IUseCase interface for testing
type MockDealUseCase struct {
	CreateDealFunc func(ctx context.Context, name string, customerID, assignedTo uuidv7.UUID, value int64, currency string, expectedCloseDate string) (*dealAggregate.Deal, error)
	GetDealFunc    func(ctx context.Context, id uuidv7.UUID) (*dealAggregate.Deal, error)
	DeleteDealFunc func(ctx context.Context, id uuidv7.UUID) error
	ListDealsFunc  func(ctx context.Context, page, pageSize int) ([]*dealAggregate.Deal, int64, error)
}

func (m *MockDealUseCase) CreateDeal(ctx context.Context, name string, customerID, assignedTo uuidv7.UUID, value int64, currency string, expectedCloseDate string) (*dealAggregate.Deal, error) {
	if m.CreateDealFunc != nil {
		return m.CreateDealFunc(ctx, name, customerID, assignedTo, value, currency, expectedCloseDate)
	}
	return fakeDeal(), nil
}

func (m *MockDealUseCase) GetDeal(ctx context.Context, id uuidv7.UUID) (*dealAggregate.Deal, error) {
	if m.GetDealFunc != nil {
		return m.GetDealFunc(ctx, id)
	}
	return fakeDeal(), nil
}

func (m *MockDealUseCase) DeleteDeal(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteDealFunc != nil {
		return m.DeleteDealFunc(ctx, id)
	}
	return nil
}

func (m *MockDealUseCase) ListDeals(ctx context.Context, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
	if m.ListDealsFunc != nil {
		return m.ListDealsFunc(ctx, page, pageSize)
	}
	return []*dealAggregate.Deal{fakeDeal()}, 1, nil
}

// Stub implementations for other IUseCase methods
func (m *MockDealUseCase) UpdateDealBasicInfo(ctx context.Context, id uuidv7.UUID, name, description string) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) UpdateDealValue(ctx context.Context, id uuidv7.UUID, value int64, currency string) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) MoveDealToStage(ctx context.Context, id uuidv7.UUID, stage dealAggregate.DealStage) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) MarkDealAsWon(ctx context.Context, id uuidv7.UUID, reason string) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) MarkDealAsLost(ctx context.Context, id uuidv7.UUID, reason string) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) UpdateDealProbability(ctx context.Context, id uuidv7.UUID, probability int) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) UpdateDealExpectedCloseDate(ctx context.Context, id uuidv7.UUID, date string) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) AssignDealToSalesRep(ctx context.Context, id, userID uuidv7.UUID) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) LinkDealToCompany(ctx context.Context, id, companyID uuidv7.UUID) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) SetDealSource(ctx context.Context, id uuidv7.UUID, source dealAggregate.DealSource) (*dealAggregate.Deal, error) {
	return fakeDeal(), nil
}
func (m *MockDealUseCase) ListDealsByStage(ctx context.Context, stage dealAggregate.DealStage, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
	return []*dealAggregate.Deal{fakeDeal()}, 1, nil
}
func (m *MockDealUseCase) ListDealsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
	return []*dealAggregate.Deal{fakeDeal()}, 1, nil
}
func (m *MockDealUseCase) ListDealsByAssignedTo(ctx context.Context, userID uuidv7.UUID, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
	return []*dealAggregate.Deal{fakeDeal()}, 1, nil
}
func (m *MockDealUseCase) GetPipelineStats(ctx context.Context) (map[dealAggregate.DealStage]int64, error) {
	return map[dealAggregate.DealStage]int64{}, nil
}
func (m *MockDealUseCase) GetWonDealsStats(ctx context.Context) (int64, int64, error) {
	return 0, 0, nil
}
func (m *MockDealUseCase) GetTotalValue(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockDealUseCase) GetWonDeals(ctx context.Context) (int64, int64, error) {
	return 0, 0, nil
}
func (m *MockDealUseCase) ListDealsByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
	return []*dealAggregate.Deal{fakeDeal()}, 1, nil
}
func (m *MockDealUseCase) ListDealsBySource(ctx context.Context, source dealAggregate.DealSource, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
	return []*dealAggregate.Deal{fakeDeal()}, 1, nil
}

func fakeDeal() *dealAggregate.Deal {
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()
	money, _ := valueobject.NewMoney(100000, "USD")
	expectedDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	d, _ := dealAggregate.NewDeal(customerID, "Test Deal", money, assignedTo, expectedDate)
	return d
}

func TestDealHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		CreateDealFunc: func(ctx context.Context, name string, customerID, assignedTo uuidv7.UUID, value int64, currency string, expectedCloseDate string) (*dealAggregate.Deal, error) {
			return fakeDeal(), nil
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.POST("/deals", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/deals", map[string]any{
		"name":                "Test Deal",
		"customer_id":         smoke.FakeUUID(),
		"assigned_to":         smoke.FakeUUID(),
		"value":               100000,
		"currency":            "USD",
		"expected_close_date": "2026-12-31",
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestDealHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()
	mockUC := &MockDealUseCase{}
	handler := dealHTTP.NewDealHandler(mockUC)
	router.POST("/deals", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/deals", map[string]any{
		"name": "", // Empty name
	})

	smoke.AssertErrorResponse(t, resp, 400, "VALIDATION_ERROR")
}

func TestDealHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		GetDealFunc: func(ctx context.Context, id uuidv7.UUID) (*dealAggregate.Deal, error) {
			return fakeDeal(), nil
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.GET("/deals/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/deals/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestDealHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		GetDealFunc: func(ctx context.Context, id uuidv7.UUID) (*dealAggregate.Deal, error) {
			return nil, deal.ErrDealNotFound
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.GET("/deals/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/deals/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "DEAL_NOT_FOUND")
}

func TestDealHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		ListDealsFunc: func(ctx context.Context, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
			return []*dealAggregate.Deal{fakeDeal()}, 1, nil
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.GET("/deals", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/deals?page=1&page_size=10", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestDealHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		ListDealsFunc: func(ctx context.Context, page, pageSize int) ([]*dealAggregate.Deal, int64, error) {
			return []*dealAggregate.Deal{}, 0, nil
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.GET("/deals", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/deals?page=1&page_size=10", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestDealHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		DeleteDealFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.DELETE("/deals/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/deals/"+smoke.FakeUUID(), nil)

	assert.Equal(t, 204, resp.Code)
}

func TestDealHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockDealUseCase{
		DeleteDealFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return deal.ErrDealNotFound
		},
	}

	handler := dealHTTP.NewDealHandler(mockUC)
	router.DELETE("/deals/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/deals/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "DEAL_NOT_FOUND")
}
