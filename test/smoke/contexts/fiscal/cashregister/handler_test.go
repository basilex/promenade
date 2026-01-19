package cashregister_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	cashregisterHTTP "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/http"
	cashregisterAggregate "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockCashRegisterUseCase implements IUseCase interface for smoke testing
type MockCashRegisterUseCase struct {
	CreateCashRegisterFunc     func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error)
	GetCashRegisterFunc        func(ctx context.Context, id uuidv7.UUID) (*cashregisterAggregate.CashRegister, error)
	GetByFiscalNumberFunc      func(ctx context.Context, fiscalNumber string) (*cashregisterAggregate.CashRegister, error)
	ListCashRegistersFunc      func(ctx context.Context, organizationID *uuidv7.UUID) ([]*cashregisterAggregate.CashRegister, error)
	ActivateCashRegisterFunc   func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error)
	DeactivateCashRegisterFunc func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error)
	UpdateCashRegisterFunc     func(ctx context.Context, cashRegister *cashregisterAggregate.CashRegister) error
	DeleteCashRegisterFunc     func(ctx context.Context, id uuidv7.UUID) error
	SyncCashRegisterFunc       func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error)
}

// Implement IUseCase methods with nil checks

func (m *MockCashRegisterUseCase) CreateCashRegister(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
	if m.CreateCashRegisterFunc != nil {
		return m.CreateCashRegisterFunc(ctx, organizationID, fiscalNumber, model, createdBy)
	}
	return nil, fmt.Errorf("CreateCashRegisterFunc not implemented")
}

func (m *MockCashRegisterUseCase) GetCashRegister(ctx context.Context, id uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
	if m.GetCashRegisterFunc != nil {
		return m.GetCashRegisterFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetCashRegisterFunc not implemented")
}

func (m *MockCashRegisterUseCase) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*cashregisterAggregate.CashRegister, error) {
	if m.GetByFiscalNumberFunc != nil {
		return m.GetByFiscalNumberFunc(ctx, fiscalNumber)
	}
	return nil, fmt.Errorf("GetByFiscalNumberFunc not implemented")
}

func (m *MockCashRegisterUseCase) ListCashRegisters(ctx context.Context, organizationID *uuidv7.UUID) ([]*cashregisterAggregate.CashRegister, error) {
	if m.ListCashRegistersFunc != nil {
		return m.ListCashRegistersFunc(ctx, organizationID)
	}
	return nil, fmt.Errorf("ListCashRegistersFunc not implemented")
}

func (m *MockCashRegisterUseCase) ActivateCashRegister(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
	if m.ActivateCashRegisterFunc != nil {
		return m.ActivateCashRegisterFunc(ctx, id, licenseKey, activatedBy)
	}
	return nil, fmt.Errorf("ActivateCashRegisterFunc not implemented")
}

func (m *MockCashRegisterUseCase) DeactivateCashRegister(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
	if m.DeactivateCashRegisterFunc != nil {
		return m.DeactivateCashRegisterFunc(ctx, id, deactivatedBy)
	}
	return nil, fmt.Errorf("DeactivateCashRegisterFunc not implemented")
}

func (m *MockCashRegisterUseCase) UpdateCashRegister(ctx context.Context, cashRegister *cashregisterAggregate.CashRegister) error {
	if m.UpdateCashRegisterFunc != nil {
		return m.UpdateCashRegisterFunc(ctx, cashRegister)
	}
	return fmt.Errorf("UpdateCashRegisterFunc not implemented")
}

func (m *MockCashRegisterUseCase) DeleteCashRegister(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteCashRegisterFunc != nil {
		return m.DeleteCashRegisterFunc(ctx, id)
	}
	return fmt.Errorf("DeleteCashRegisterFunc not implemented")
}

func (m *MockCashRegisterUseCase) SyncCashRegister(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
	if m.SyncCashRegisterFunc != nil {
		return m.SyncCashRegisterFunc(ctx, id, syncedBy)
	}
	return nil, fmt.Errorf("SyncCashRegisterFunc not implemented")
}

// fakeCashRegister creates a fake cash register for testing
func fakeCashRegister() *cashregisterAggregate.CashRegister {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	cr, _ := cashregisterAggregate.NewCashRegister(orgID, "FN-123456", "Model X", userID)
	return cr
}

// ============================================================================
// Smoke Tests
// ============================================================================

func TestCashRegisterHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		CreateCashRegisterFunc: func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return fakeCashRegister(), nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"organization_id": smoke.FakeUUID(),
		"fiscal_number":   "FN-123456",
		"model":           "Model X",
		"created_by":      smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)

	smoke.AssertSuccessResponse(t, w, http.StatusCreated)
}

func TestCashRegisterHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{}
	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"fiscal_number": "FN-123456",
		// Missing required fields
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)

	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		GetCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return fakeCashRegister(), nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.GET("/fiscal/cash-registers/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestCashRegisterHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		GetCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return nil, cashregister.ErrCashRegisterNotFound
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.GET("/fiscal/cash-registers/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		ListCashRegistersFunc: func(ctx context.Context, organizationID *uuidv7.UUID) ([]*cashregisterAggregate.CashRegister, error) {
			return []*cashregisterAggregate.CashRegister{fakeCashRegister(), fakeCashRegister()}, nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.GET("/fiscal/cash-registers", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestCashRegisterHandler_Activate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		ActivateCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			cr := fakeCashRegister()
			_ = cr.Activate(licenseKey, activatedBy)
			return cr, nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LICENSE-KEY-123",
		"activated_by": smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestCashRegisterHandler_Deactivate_Success(t *testing.T) {
	router := smoke.SetupRouter()
	userID := smoke.FakeUUID()

	mockUC := &MockCashRegisterUseCase{
		GetCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			cr := fakeCashRegister()
			_ = cr.Activate("LICENSE-KEY", uuidv7.New())
			return cr, nil
		},
		DeactivateCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			cr := fakeCashRegister()
			_ = cr.Activate("LICENSE-KEY", deactivatedBy)
			_ = cr.Deactivate(deactivatedBy)
			return cr, nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}, handler.Deactivate)

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestCashRegisterHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		DeleteCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.DELETE("/fiscal/cash-registers/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestCashRegisterHandler_Sync_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		SyncCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			cr := fakeCashRegister()
			_ = cr.UpdateLastSync(syncedBy)
			return cr, nil
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/sync", handler.Sync)

	body := map[string]any{
		"synced_by": smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/sync", body)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestCashRegisterHandler_Activate_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		ActivateCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return nil, cashregister.ErrCashRegisterNotFound
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LICENSE-KEY-404",
		"activated_by": smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)

	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Activate_AlreadyActive(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		ActivateCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return nil, cashregister.ErrCashRegisterAlreadyActive
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LICENSE-KEY-ACTIVE",
		"activated_by": smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)

	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Deactivate_NotFound(t *testing.T) {
	router := smoke.SetupRouter()
	userID := smoke.FakeUUID()

	mockUC := &MockCashRegisterUseCase{
		DeactivateCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return nil, cashregister.ErrCashRegisterNotFound
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}, handler.Deactivate)

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)

	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Deactivate_AlreadyInactive(t *testing.T) {
	router := smoke.SetupRouter()
	userID := smoke.FakeUUID()

	mockUC := &MockCashRegisterUseCase{
		DeactivateCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return nil, cashregister.ErrCashRegisterAlreadyInactive
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}, handler.Deactivate)

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)

	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		DeleteCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return cashregister.ErrCashRegisterNotFound
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.DELETE("/fiscal/cash-registers/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Sync_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCashRegisterUseCase{
		SyncCashRegisterFunc: func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*cashregisterAggregate.CashRegister, error) {
			return nil, cashregister.ErrCashRegisterNotFound
		},
	}

	handler := cashregisterHTTP.NewCashRegisterHandler(mockUC)
	router.POST("/fiscal/cash-registers/:id/sync", handler.Sync)

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/sync", nil)

	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}
