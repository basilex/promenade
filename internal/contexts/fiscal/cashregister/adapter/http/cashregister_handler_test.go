package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	cashregistererrors "github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type mockCashRegisterUseCase struct {
	createFunc     func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error)
	getFunc        func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error)
	listFunc       func(ctx context.Context, organizationID *uuidv7.UUID) ([]*aggregate.CashRegister, error)
	activateFunc   func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error)
	deactivateFunc func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error)
	deleteFunc     func(ctx context.Context, id uuidv7.UUID) error
	syncFunc       func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error)
}

func (m *mockCashRegisterUseCase) CreateCashRegister(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, organizationID, fiscalNumber, model, createdBy)
	}
	return nil, fmt.Errorf("create not implemented")
}

func (m *mockCashRegisterUseCase) GetCashRegister(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, fmt.Errorf("get not implemented")
}

func (m *mockCashRegisterUseCase) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCashRegisterUseCase) ListCashRegisters(ctx context.Context, organizationID *uuidv7.UUID) ([]*aggregate.CashRegister, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, organizationID)
	}
	return nil, fmt.Errorf("list not implemented")
}

func (m *mockCashRegisterUseCase) ActivateCashRegister(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	if m.activateFunc != nil {
		return m.activateFunc(ctx, id, licenseKey, activatedBy)
	}
	return nil, fmt.Errorf("activate not implemented")
}

func (m *mockCashRegisterUseCase) DeactivateCashRegister(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	if m.deactivateFunc != nil {
		return m.deactivateFunc(ctx, id, deactivatedBy)
	}
	return nil, fmt.Errorf("deactivate not implemented")
}

func (m *mockCashRegisterUseCase) UpdateCashRegister(ctx context.Context, cashRegister *aggregate.CashRegister) error {
	return fmt.Errorf("not implemented")
}

func (m *mockCashRegisterUseCase) DeleteCashRegister(ctx context.Context, id uuidv7.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return fmt.Errorf("delete not implemented")
}

func (m *mockCashRegisterUseCase) SyncCashRegister(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	if m.syncFunc != nil {
		return m.syncFunc(ctx, id, syncedBy)
	}
	return nil, fmt.Errorf("sync not implemented")
}

func TestCashRegisterHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		createFunc: func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return aggregate.NewCashRegister(organizationID, fiscalNumber, model, createdBy)
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"organization_id": smoke.FakeUUID(),
		"fiscal_number":   "FN-123",
		"model":           "Checkbox",
		"created_by":      smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)
	smoke.AssertSuccessResponse(t, resp, http.StatusCreated)
}

func TestCashRegisterHandler_Create_BindError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}
	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", map[string]any{})
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Create_InvalidOrganizationID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"organization_id": "invalid",
		"fiscal_number":   "FN-124",
		"model":           "Checkbox",
		"created_by":      smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Create_InvalidCreatedBy(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"organization_id": smoke.FakeUUID(),
		"fiscal_number":   "FN-126",
		"model":           "Checkbox",
		"created_by":      "invalid",
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Create_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		createFunc: func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, fmt.Errorf("create error")
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"organization_id": smoke.FakeUUID(),
		"fiscal_number":   "FN-127",
		"model":           "Checkbox",
		"created_by":      smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestCashRegisterHandler_Create_ErrorMappings(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "organization required", err: cashregistererrors.ErrOrganizationIDRequired},
		{name: "fiscal number required", err: cashregistererrors.ErrFiscalNumberRequired},
		{name: "model required", err: cashregistererrors.ErrModelRequired},
		{name: "fiscal number exists", err: cashregistererrors.ErrCashRegisterFiscalNumberExists},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := smoke.SetupRouter()

			usecase := &mockCashRegisterUseCase{
				createFunc: func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error) {
					return nil, tc.err
				},
			}

			handler := NewCashRegisterHandler(usecase)
			router.POST("/fiscal/cash-registers", handler.Create)

			body := map[string]any{
				"organization_id": smoke.FakeUUID(),
				"fiscal_number":   "FN-128",
				"model":           "Checkbox",
				"created_by":      smoke.FakeUUID(),
			}

			resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)
			smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
		})
	}
}

func TestCashRegisterHandler_Create_DuplicateFiscalNumber(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		createFunc: func(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterFiscalNumberExists
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers", handler.Create)

	body := map[string]any{
		"organization_id": smoke.FakeUUID(),
		"fiscal_number":   "FN-125",
		"model":           "Checkbox",
		"created_by":      smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		getFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.GET("/fiscal/cash-registers/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_GetByID_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}

	handler := NewCashRegisterHandler(usecase)
	router.GET("/fiscal/cash-registers/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers/invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		getFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return aggregate.NewCashRegister(uuidv7.New(), "FN-200", "Test", uuidv7.New())
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.GET("/fiscal/cash-registers/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestCashRegisterHandler_List_InvalidOrganizationID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}
	handler := NewCashRegisterHandler(usecase)
	router.GET("/fiscal/cash-registers", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers?organization_id=invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		listFunc: func(ctx context.Context, organizationID *uuidv7.UUID) ([]*aggregate.CashRegister, error) {
			cr, _ := aggregate.NewCashRegister(uuidv7.New(), "FN-201", "Model", uuidv7.New())
			return []*aggregate.CashRegister{cr}, nil
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.GET("/fiscal/cash-registers", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers", nil)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestCashRegisterHandler_List_Error(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		listFunc: func(ctx context.Context, organizationID *uuidv7.UUID) ([]*aggregate.CashRegister, error) {
			return nil, fmt.Errorf("list error")
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.GET("/fiscal/cash-registers", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/cash-registers", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestCashRegisterHandler_Activate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		activateFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			cr, _ := aggregate.NewCashRegister(uuidv7.New(), "FN-202", "Model", uuidv7.New())
			_ = cr.Activate(licenseKey, activatedBy)
			return cr, nil
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LIC-202",
		"activated_by": smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestCashRegisterHandler_Activate_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}
	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LIC-203",
		"activated_by": smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/invalid/activate", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Activate_LicenseRequired(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		activateFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrLicenseKeyRequired
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "",
		"activated_by": smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Activate_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		activateFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LIC-204",
		"activated_by": smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Activate_AlreadyActive(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		activateFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterAlreadyActive
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LIC-205",
		"activated_by": smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Activate_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		activateFunc: func(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, fmt.Errorf("activate error")
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/activate", handler.Activate)

	body := map[string]any{
		"license_key":  "LIC-206",
		"activated_by": smoke.FakeUUID(),
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/activate", body)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestCashRegisterHandler_Deactivate_Unauthorized(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}
	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/deactivate", handler.Deactivate)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Deactivate_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		deactivateFunc: func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", smoke.FakeUUID())
		handler.Deactivate(c)
	})

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Deactivate_AlreadyInactive(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		deactivateFunc: func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterAlreadyInactive
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", smoke.FakeUUID())
		handler.Deactivate(c)
	})

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Deactivate_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", smoke.FakeUUID())
		handler.Deactivate(c)
	})

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/invalid/deactivate", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Deactivate_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		deactivateFunc: func(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, fmt.Errorf("deactivate error")
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/deactivate", func(c *gin.Context) {
		c.Set("user_id", smoke.FakeUUID())
		handler.Deactivate(c)
	})

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/deactivate", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestCashRegisterHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		deleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return cashregistererrors.ErrCashRegisterNotFound
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.DELETE("/fiscal/cash-registers/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Delete_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		deleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return fmt.Errorf("delete failed")
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.DELETE("/fiscal/cash-registers/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestCashRegisterHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		deleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.DELETE("/fiscal/cash-registers/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/cash-registers/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestCashRegisterHandler_Delete_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}
	handler := NewCashRegisterHandler(usecase)
	router.DELETE("/fiscal/cash-registers/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/cash-registers/invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Sync_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		syncFunc: func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/sync", handler.Sync)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/sync", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestCashRegisterHandler_Sync_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{}
	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/sync", handler.Sync)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/invalid/sync", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestCashRegisterHandler_Sync_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		syncFunc: func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return aggregate.NewCashRegister(uuidv7.New(), "FN-208", "Model", uuidv7.New())
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/sync", handler.Sync)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/sync", nil)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestCashRegisterHandler_Sync_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockCashRegisterUseCase{
		syncFunc: func(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, fmt.Errorf("sync error")
		},
	}

	handler := NewCashRegisterHandler(usecase)
	router.POST("/fiscal/cash-registers/:id/sync", handler.Sync)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/cash-registers/"+smoke.FakeUUID()+"/sync", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}
