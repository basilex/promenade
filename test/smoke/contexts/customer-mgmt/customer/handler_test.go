package customer_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	customerHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/http"
	customerAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/aggregate"
	customerUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockCustomerUseCase is a mock implementation of customer.ICustomerUseCase for testing
type MockCustomerUseCase struct {
	CreateCustomerFunc            func(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*customerAggregate.Customer, error)
	CreateB2BCustomerFunc         func(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*customerAggregate.Customer, error)
	GetCustomerFunc               func(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error)
	GetCustomerByEmailFunc        func(ctx context.Context, email string) (*customerAggregate.Customer, error)
	UpdateCustomerFunc            func(ctx context.Context, c *customerAggregate.Customer) error
	SetCustomerPhoneFunc          func(ctx context.Context, customerID uuidv7.UUID, phone string) error
	QualifyAsProspectFunc         func(ctx context.Context, customerID uuidv7.UUID) error
	ConvertToCustomerFunc         func(ctx context.Context, customerID uuidv7.UUID) error
	ChurnCustomerFunc             func(ctx context.Context, customerID uuidv7.UUID, reason string) error
	ReactivateCustomerFunc        func(ctx context.Context, customerID uuidv7.UUID) error
	UpgradeCustomerTierFunc       func(ctx context.Context, customerID uuidv7.UUID, newTier customerAggregate.CustomerTier) error
	DowngradeCustomerTierFunc     func(ctx context.Context, customerID uuidv7.UUID, newTier customerAggregate.CustomerTier) error
	ReassignCustomerFunc          func(ctx context.Context, customerID, newRepID uuidv7.UUID) error
	LinkCustomerToUserFunc        func(ctx context.Context, customerID, userID uuidv7.UUID) error
	AddTagToCustomerFunc          func(ctx context.Context, customerID uuidv7.UUID, tag string) error
	RemoveTagFromCustomerFunc     func(ctx context.Context, customerID uuidv7.UUID, tag string) error
	ListCustomersByAssignedToFunc func(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*customerAggregate.Customer, int, error)
	ListCustomersByStatusFunc     func(ctx context.Context, status customerAggregate.CustomerStatus, limit, offset int) ([]*customerAggregate.Customer, int, error)
	ListCustomersByTierFunc       func(ctx context.Context, tier customerAggregate.CustomerTier, limit, offset int) ([]*customerAggregate.Customer, int, error)
	ListCustomersFunc             func(ctx context.Context, limit, offset int) ([]*customerAggregate.Customer, int, error)
	GetCustomerStatsFunc          func(ctx context.Context) (*customerUseCase.CustomerStats, error)
	DeleteCustomerFunc            func(ctx context.Context, id uuidv7.UUID) error
}

// Implement all ICustomerUseCase methods with nil checks
func (m *MockCustomerUseCase) CreateCustomer(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*customerAggregate.Customer, error) {
	if m.CreateCustomerFunc != nil {
		return m.CreateCustomerFunc(ctx, name, email, source, assignedTo)
	}
	return nil, nil
}

func (m *MockCustomerUseCase) CreateB2BCustomer(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*customerAggregate.Customer, error) {
	if m.CreateB2BCustomerFunc != nil {
		return m.CreateB2BCustomerFunc(ctx, name, email, source, companyID, assignedTo)
	}
	return nil, nil
}

func (m *MockCustomerUseCase) GetCustomer(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error) {
	if m.GetCustomerFunc != nil {
		return m.GetCustomerFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCustomerUseCase) GetCustomerByEmail(ctx context.Context, email string) (*customerAggregate.Customer, error) {
	if m.GetCustomerByEmailFunc != nil {
		return m.GetCustomerByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockCustomerUseCase) UpdateCustomer(ctx context.Context, c *customerAggregate.Customer) error {
	if m.UpdateCustomerFunc != nil {
		return m.UpdateCustomerFunc(ctx, c)
	}
	return nil
}

func (m *MockCustomerUseCase) SetCustomerPhone(ctx context.Context, customerID uuidv7.UUID, phone string) error {
	if m.SetCustomerPhoneFunc != nil {
		return m.SetCustomerPhoneFunc(ctx, customerID, phone)
	}
	return nil
}

func (m *MockCustomerUseCase) QualifyAsProspect(ctx context.Context, customerID uuidv7.UUID) error {
	if m.QualifyAsProspectFunc != nil {
		return m.QualifyAsProspectFunc(ctx, customerID)
	}
	return nil
}

func (m *MockCustomerUseCase) ConvertToCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	if m.ConvertToCustomerFunc != nil {
		return m.ConvertToCustomerFunc(ctx, customerID)
	}
	return nil
}

func (m *MockCustomerUseCase) ChurnCustomer(ctx context.Context, customerID uuidv7.UUID, reason string) error {
	if m.ChurnCustomerFunc != nil {
		return m.ChurnCustomerFunc(ctx, customerID, reason)
	}
	return nil
}

func (m *MockCustomerUseCase) ReactivateCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	if m.ReactivateCustomerFunc != nil {
		return m.ReactivateCustomerFunc(ctx, customerID)
	}
	return nil
}

func (m *MockCustomerUseCase) UpgradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier customerAggregate.CustomerTier) error {
	if m.UpgradeCustomerTierFunc != nil {
		return m.UpgradeCustomerTierFunc(ctx, customerID, newTier)
	}
	return nil
}

func (m *MockCustomerUseCase) DowngradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier customerAggregate.CustomerTier) error {
	if m.DowngradeCustomerTierFunc != nil {
		return m.DowngradeCustomerTierFunc(ctx, customerID, newTier)
	}
	return nil
}

func (m *MockCustomerUseCase) ReassignCustomer(ctx context.Context, customerID, newRepID uuidv7.UUID) error {
	if m.ReassignCustomerFunc != nil {
		return m.ReassignCustomerFunc(ctx, customerID, newRepID)
	}
	return nil
}

func (m *MockCustomerUseCase) LinkCustomerToUser(ctx context.Context, customerID, userID uuidv7.UUID) error {
	if m.LinkCustomerToUserFunc != nil {
		return m.LinkCustomerToUserFunc(ctx, customerID, userID)
	}
	return nil
}

func (m *MockCustomerUseCase) AddTagToCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	if m.AddTagToCustomerFunc != nil {
		return m.AddTagToCustomerFunc(ctx, customerID, tag)
	}
	return nil
}

func (m *MockCustomerUseCase) RemoveTagFromCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	if m.RemoveTagFromCustomerFunc != nil {
		return m.RemoveTagFromCustomerFunc(ctx, customerID, tag)
	}
	return nil
}

func (m *MockCustomerUseCase) ListCustomersByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*customerAggregate.Customer, int, error) {
	if m.ListCustomersByAssignedToFunc != nil {
		return m.ListCustomersByAssignedToFunc(ctx, repID, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockCustomerUseCase) ListCustomersByStatus(ctx context.Context, status customerAggregate.CustomerStatus, limit, offset int) ([]*customerAggregate.Customer, int, error) {
	if m.ListCustomersByStatusFunc != nil {
		return m.ListCustomersByStatusFunc(ctx, status, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockCustomerUseCase) ListCustomersByTier(ctx context.Context, tier customerAggregate.CustomerTier, limit, offset int) ([]*customerAggregate.Customer, int, error) {
	if m.ListCustomersByTierFunc != nil {
		return m.ListCustomersByTierFunc(ctx, tier, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockCustomerUseCase) ListCustomers(ctx context.Context, limit, offset int) ([]*customerAggregate.Customer, int, error) {
	if m.ListCustomersFunc != nil {
		return m.ListCustomersFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockCustomerUseCase) GetCustomerStats(ctx context.Context) (*customerUseCase.CustomerStats, error) {
	if m.GetCustomerStatsFunc != nil {
		return m.GetCustomerStatsFunc(ctx)
	}
	return nil, nil
}

func (m *MockCustomerUseCase) DeleteCustomer(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteCustomerFunc != nil {
		return m.DeleteCustomerFunc(ctx, id)
	}
	return nil
}

// fakeCustomer creates a fake customer for testing
func fakeCustomer() *customerAggregate.Customer {
	c, _ := customerAggregate.NewCustomer("John Doe", "john@example.com", "Website", uuidv7.New())
	return c
}

// Test Create B2C Customer
func TestCustomerHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		CreateCustomerFunc: func(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*customerAggregate.Customer, error) {
			return fakeCustomer(), nil
		},
		SetCustomerPhoneFunc: func(ctx context.Context, customerID uuidv7.UUID, phone string) error {
			return nil
		},
		AddTagToCustomerFunc: func(ctx context.Context, customerID uuidv7.UUID, tag string) error {
			return nil
		},
		GetCustomerFunc: func(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error) {
			return fakeCustomer(), nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.POST("/customers", handler.Create)

	body := map[string]any{
		"name":        "John Doe",
		"email":       "john@example.com",
		"source":      "Website",
		"assigned_to": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/customers", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestCustomerHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{}
	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.POST("/customers", handler.Create)

	body := map[string]any{
		"name": "John Doe",
		// Missing required fields
	}

	w := smoke.MakeRequest(t, router, "POST", "/customers", body)
	smoke.AssertErrorResponse(t, w, 400, "BAD_REQUEST")
}

// Test Create B2B Customer
func TestCustomerHandler_CreateB2B_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		CreateB2BCustomerFunc: func(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*customerAggregate.Customer, error) {
			return fakeCustomer(), nil
		},
		SetCustomerPhoneFunc: func(ctx context.Context, customerID uuidv7.UUID, phone string) error {
			return nil
		},
		AddTagToCustomerFunc: func(ctx context.Context, customerID uuidv7.UUID, tag string) error {
			return nil
		},
		GetCustomerFunc: func(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error) {
			return fakeCustomer(), nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.POST("/customers/b2b", handler.CreateB2B)

	body := map[string]any{
		"name":        "Jane Smith",
		"email":       "jane@company.com",
		"source":      "Referral",
		"company_id":  smoke.FakeUUID(),
		"assigned_to": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/customers/b2b", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

// Test GetByID
func TestCustomerHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		GetCustomerFunc: func(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error) {
			return fakeCustomer(), nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.GET("/customers/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/customers/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCustomerHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		GetCustomerFunc: func(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error) {
			return nil, customer.ErrCustomerNotFound
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.GET("/customers/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/customers/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

// Test List
func TestCustomerHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		ListCustomersFunc: func(ctx context.Context, limit, offset int) ([]*customerAggregate.Customer, int, error) {
			return []*customerAggregate.Customer{fakeCustomer()}, 1, nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.GET("/customers", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/customers", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCustomerHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		ListCustomersFunc: func(ctx context.Context, limit, offset int) ([]*customerAggregate.Customer, int, error) {
			return []*customerAggregate.Customer{}, 0, nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.GET("/customers", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/customers", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// Test Delete
func TestCustomerHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		DeleteCustomerFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.DELETE("/customers/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/customers/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCustomerHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		DeleteCustomerFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return customer.ErrCustomerNotFound
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.DELETE("/customers/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/customers/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

// Test QualifyAsProspect (lifecycle transition)
func TestCustomerHandler_QualifyAsProspect_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		QualifyAsProspectFunc: func(ctx context.Context, customerID uuidv7.UUID) error {
			return nil
		},
		GetCustomerFunc: func(ctx context.Context, id uuidv7.UUID) (*customerAggregate.Customer, error) {
			return fakeCustomer(), nil
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.POST("/customers/:id/qualify", handler.QualifyAsProspect)

	w := smoke.MakeRequest(t, router, "POST", "/customers/"+smoke.FakeUUID()+"/qualify", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCustomerHandler_QualifyAsProspect_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCustomerUseCase{
		QualifyAsProspectFunc: func(ctx context.Context, customerID uuidv7.UUID) error {
			return customer.ErrCustomerNotFound
		},
	}

	handler := customerHTTP.NewCustomerHandler(mockUC)
	router.POST("/customers/:id/qualify", handler.QualifyAsProspect)

	w := smoke.MakeRequest(t, router, "POST", "/customers/"+smoke.FakeUUID()+"/qualify", nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}
