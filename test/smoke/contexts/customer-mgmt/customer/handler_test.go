package customer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	httpHandler "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCustomerUseCase is a mock implementation of ICustomerUseCase
type MockCustomerUseCase struct {
	mock.Mock
}

func (m *MockCustomerUseCase) CreateCustomer(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*customer.Customer, error) {
	args := m.Called(ctx, name, email, source, assignedTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.Customer), args.Error(1)
}

func (m *MockCustomerUseCase) CreateB2BCustomer(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*customer.Customer, error) {
	args := m.Called(ctx, name, email, source, companyID, assignedTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.Customer), args.Error(1)
}

func (m *MockCustomerUseCase) GetCustomer(ctx context.Context, id uuidv7.UUID) (*customer.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.Customer), args.Error(1)
}

func (m *MockCustomerUseCase) GetCustomerByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.Customer), args.Error(1)
}

func (m *MockCustomerUseCase) SetCustomerPhone(ctx context.Context, customerID uuidv7.UUID, phone string) error {
	args := m.Called(ctx, customerID, phone)
	return args.Error(0)
}

func (m *MockCustomerUseCase) UpdateCustomer(ctx context.Context, customer *customer.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerUseCase) DeleteCustomer(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerUseCase) QualifyAsProspect(ctx context.Context, customerID uuidv7.UUID) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockCustomerUseCase) ConvertToCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockCustomerUseCase) ChurnCustomer(ctx context.Context, customerID uuidv7.UUID, reason string) error {
	args := m.Called(ctx, customerID, reason)
	return args.Error(0)
}

func (m *MockCustomerUseCase) ReactivateCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockCustomerUseCase) LinkCustomerToUser(ctx context.Context, customerID, userID uuidv7.UUID) error {
	args := m.Called(ctx, customerID, userID)
	return args.Error(0)
}

func (m *MockCustomerUseCase) ReassignCustomer(ctx context.Context, customerID, newRepID uuidv7.UUID) error {
	args := m.Called(ctx, customerID, newRepID)
	return args.Error(0)
}

func (m *MockCustomerUseCase) UpgradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier customer.CustomerTier) error {
	args := m.Called(ctx, customerID, newTier)
	return args.Error(0)
}

func (m *MockCustomerUseCase) DowngradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier customer.CustomerTier) error {
	args := m.Called(ctx, customerID, newTier)
	return args.Error(0)
}

func (m *MockCustomerUseCase) AddTagToCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	args := m.Called(ctx, customerID, tag)
	return args.Error(0)
}

func (m *MockCustomerUseCase) RemoveTagFromCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	args := m.Called(ctx, customerID, tag)
	return args.Error(0)
}

func (m *MockCustomerUseCase) ListCustomersByStatus(ctx context.Context, status customer.CustomerStatus, limit, offset int) ([]*customer.Customer, int, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*customer.Customer), args.Int(1), args.Error(2)
}

func (m *MockCustomerUseCase) ListCustomersByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*customer.Customer, int, error) {
	args := m.Called(ctx, repID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*customer.Customer), args.Int(1), args.Error(2)
}

func (m *MockCustomerUseCase) ListCustomersByTier(ctx context.Context, tier customer.CustomerTier, limit, offset int) ([]*customer.Customer, int, error) {
	args := m.Called(ctx, tier, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*customer.Customer), args.Int(1), args.Error(2)
}

func (m *MockCustomerUseCase) ListCustomers(ctx context.Context, limit, offset int) ([]*customer.Customer, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*customer.Customer), args.Int(1), args.Error(2)
}

func (m *MockCustomerUseCase) GetCustomerStats(ctx context.Context) (*customer.CustomerStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.CustomerStats), args.Error(1)
}

func setupTestRouter(mockUC *MockCustomerUseCase) (*gin.Engine, *httpHandler.CustomerHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := httpHandler.NewCustomerHandler(mockUC)
	return router, h
}

func TestCustomerHandler_Create(t *testing.T) {
	t.Run("create B2C customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.POST("/customers", h.Create)

		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		mockUC.On("CreateCustomer", mock.Anything, "John Doe", "john@example.com", "website", assignedTo).Return(c, nil)
		mockUC.On("GetCustomer", mock.Anything, c.ID).Return(c, nil)

		body := map[string]interface{}{
			"name":        "John Doe",
			"email":       "john@example.com",
			"source":      "website",
			"assigned_to": assignedTo.String(),
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("create B2B customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.POST("/customers", h.Create)

		assignedTo := uuidv7.New()
		companyID := uuidv7.New()
		// Note: Handler currently calls CreateCustomer for all customers, not CreateB2BCustomer
		c, _ := customer.NewCustomer("Jane Smith", "jane@company.com", "referral", assignedTo)
		mockUC.On("CreateCustomer", mock.Anything, "Jane Smith", "jane@company.com", "referral", assignedTo).Return(c, nil)
		mockUC.On("GetCustomer", mock.Anything, c.ID).Return(c, nil)

		body := map[string]interface{}{
			"name":        "Jane Smith",
			"email":       "jane@company.com",
			"source":      "referral",
			"company_id":  companyID.String(),
			"assigned_to": assignedTo.String(),
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("create customer invalid JSON", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.POST("/customers", h.Create)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomerHandler_GetByID(t *testing.T) {
	t.Run("get customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/:id", h.GetByID)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		req := httptest.NewRequest(http.MethodGet, "/customers/"+id.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("get customer not found", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/:id", h.GetByID)

		id := uuidv7.New()
		mockUC.On("GetCustomer", mock.Anything, id).Return(nil, customer.ErrCustomerNotFound)

		req := httptest.NewRequest(http.MethodGet, "/customers/"+id.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("get customer invalid UUID", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/:id", h.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/customers/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomerHandler_GetByEmail(t *testing.T) {
	t.Run("get by email success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/by-email", h.GetByEmail)

		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		mockUC.On("GetCustomerByEmail", mock.Anything, "test@example.com").Return(c, nil)

		req := httptest.NewRequest(http.MethodGet, "/customers/by-email?email=test@example.com", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("get by email not found", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/by-email", h.GetByEmail)

		mockUC.On("GetCustomerByEmail", mock.Anything, "notfound@example.com").Return(nil, customer.ErrCustomerNotFound)

		req := httptest.NewRequest(http.MethodGet, "/customers/by-email?email=notfound@example.com", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_Update(t *testing.T) {
	t.Run("update customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PUT("/customers/:id", h.Update)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Original Name", "test@example.com", "website", assignedTo)
		c.ID = id
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)
		c.Name = "Updated Name"
		mockUC.On("UpdateCustomer", mock.Anything, c).Return(nil)

		body := map[string]interface{}{
			"name": "Updated Name",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/customers/"+id.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("update customer not found", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PUT("/customers/:id", h.Update)

		id := uuidv7.New()
		mockUC.On("GetCustomer", mock.Anything, id).Return(nil, customer.ErrCustomerNotFound)

		body := map[string]interface{}{
			"name": "Updated Name",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/customers/"+id.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_Delete(t *testing.T) {
	t.Run("delete customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.DELETE("/customers/:id", h.Delete)

		id := uuidv7.New()
		mockUC.On("DeleteCustomer", mock.Anything, id).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/customers/"+id.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("delete customer not found", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.DELETE("/customers/:id", h.Delete)

		id := uuidv7.New()
		mockUC.On("DeleteCustomer", mock.Anything, id).Return(customer.ErrCustomerNotFound)

		req := httptest.NewRequest(http.MethodDelete, "/customers/"+id.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_QualifyAsProspect(t *testing.T) {
	t.Run("qualify as prospect success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PATCH("/customers/:id/qualify-prospect", h.QualifyAsProspect)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		mockUC.On("QualifyAsProspect", mock.Anything, id).Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		req := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String()+"/qualify-prospect", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_ConvertToCustomer(t *testing.T) {
	t.Run("convert to customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PATCH("/customers/:id/convert", h.ConvertToCustomer)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		mockUC.On("ConvertToCustomer", mock.Anything, id).Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		req := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String()+"/convert", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_MarkAsChurned(t *testing.T) {
	t.Run("mark as churned success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PATCH("/customers/:id/churn", h.Churn)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		mockUC.On("ChurnCustomer", mock.Anything, id, "Not satisfied").Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		body := map[string]interface{}{
			"reason": "Not satisfied",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String()+"/churn", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_Reactivate(t *testing.T) {
	t.Run("reactivate customer success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PATCH("/customers/:id/reactivate", h.Reactivate)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		mockUC.On("ReactivateCustomer", mock.Anything, id).Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		req := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String()+"/reactivate", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_LinkToUser(t *testing.T) {
	t.Run("link to user success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.POST("/customers/:id/link-user", h.LinkToUser)

		customerID := uuidv7.New()
		userID := uuidv7.New()

		mockUC.On("LinkCustomerToUser", mock.Anything, customerID, userID).Return(nil)

		reqBody := map[string]interface{}{
			"user_id": userID.String(),
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/customers/"+customerID.String()+"/link-user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_AssignTo(t *testing.T) {
	t.Run("assign to user success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.POST("/customers/:id/assign", h.AssignTo)

		customerID := uuidv7.New()
		newRepID := uuidv7.New()

		mockUC.On("ReassignCustomer", mock.Anything, customerID, newRepID).Return(nil)

		reqBody := map[string]interface{}{
			"new_rep_id": newRepID.String(),
		}
		jsonData, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/customers/"+customerID.String()+"/assign", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_UpgradeTier(t *testing.T) {
	t.Run("upgrade tier success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.PATCH("/customers/:id/upgrade/:tier", h.UpgradeTier)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		_ = c.UpgradeTier(customer.CustomerTierPro)
		mockUC.On("UpgradeCustomerTier", mock.Anything, id, customer.CustomerTierPro).Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		req := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String()+"/upgrade/pro", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_AddTag(t *testing.T) {
	t.Run("add tag success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.POST("/customers/:id/tags", h.AddTag)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		c.AddTag("vip")
		mockUC.On("AddTagToCustomer", mock.Anything, id, "vip").Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		body := map[string]interface{}{
			"tag": "vip",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/customers/"+id.String()+"/tags", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_RemoveTag(t *testing.T) {
	t.Run("remove tag success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.DELETE("/customers/:id/tags/:tag", h.RemoveTag)

		id := uuidv7.New()
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = id
		mockUC.On("RemoveTagFromCustomer", mock.Anything, id, "vip").Return(nil)
		mockUC.On("GetCustomer", mock.Anything, id).Return(c, nil)

		req := httptest.NewRequest(http.MethodDelete, "/customers/"+id.String()+"/tags/vip", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_ListByStatus(t *testing.T) {
	t.Run("list by status success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/status/:status", h.ListByStatus)

		assignedTo := uuidv7.New()
		c1, _ := customer.NewCustomer("Lead 1", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
		c2, _ := customer.NewCustomer("Lead 2", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
		customers := []*customer.Customer{c1, c2}
		mockUC.On("ListCustomersByStatus", mock.Anything, customer.CustomerStatusLead, 20, 0).Return(customers, 2, nil)

		req := httptest.NewRequest(http.MethodGet, "/customers/status/lead", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_ListByTier(t *testing.T) {
	t.Run("list by tier success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/tier/:tier", h.ListByTier)

		assignedTo := uuidv7.New()
		c1, _ := customer.NewCustomer("Free 1", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
		c2, _ := customer.NewCustomer("Free 2", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
		customers := []*customer.Customer{c1, c2}
		mockUC.On("ListCustomersByTier", mock.Anything, customer.CustomerTierFree, 20, 0).Return(customers, 2, nil)

		req := httptest.NewRequest(http.MethodGet, "/customers/tier/free", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_List(t *testing.T) {
	t.Run("list customers success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers", h.List)

		assignedTo := uuidv7.New()
		c1, _ := customer.NewCustomer("Customer 1", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
		c2, _ := customer.NewCustomer("Customer 2", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
		customers := []*customer.Customer{c1, c2}
		mockUC.On("ListCustomers", mock.Anything, 20, 0).Return(customers, 2, nil)

		req := httptest.NewRequest(http.MethodGet, "/customers", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestCustomerHandler_GetStats(t *testing.T) {
	t.Run("get stats success", func(t *testing.T) {
		mockUC := new(MockCustomerUseCase)
		router, h := setupTestRouter(mockUC)
		router.GET("/customers/stats", h.GetStats)

		stats := &customer.CustomerStats{
			TotalCustomers: 100,
			ByStatus: map[customer.CustomerStatus]int{
				customer.CustomerStatusLead:     30,
				customer.CustomerStatusProspect: 25,
				customer.CustomerStatusCustomer: 40,
				customer.CustomerStatusChurned:  5,
			},
			ByTier: map[customer.CustomerTier]int{
				customer.CustomerTierFree:       60,
				customer.CustomerTierBasic:      25,
				customer.CustomerTierPro:        10,
				customer.CustomerTierEnterprise: 5,
			},
			ActiveCustomers:  40,
			ChurnedCustomers: 5,
		}
		mockUC.On("GetCustomerStats", mock.Anything).Return(stats, nil)

		req := httptest.NewRequest(http.MethodGet, "/customers/stats", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
}
