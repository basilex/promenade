package taxcode_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	taxcodeHTTP "github.com/basilex/promenade/internal/contexts/accounting/taxcode/adapter/http"
	taxcodeAggregate "github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockTaxCodeUseCase struct {
	CreateTaxCodeFunc              func(ctx context.Context, organizationID uuidv7.UUID, code, name string, taxType taxcodeAggregate.TaxType, rate int, createdBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	GetTaxCodeByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	GetTaxCodeByCodeFunc           func(ctx context.Context, organizationID uuidv7.UUID, code string) (*taxcodeAggregate.TaxCode, error)
	UpdateTaxCodeFunc              func(ctx context.Context, id uuidv7.UUID, rate int, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	SetTaxPayableAccountFunc       func(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	SetTaxReceivableAccountFunc    func(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	ActivateTaxCodeFunc            func(ctx context.Context, id, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	DeactivateTaxCodeFunc          func(ctx context.Context, id, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error)
	DeleteTaxCodeFunc              func(ctx context.Context, id uuidv7.UUID) error
	ListTaxCodesByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*taxcodeAggregate.TaxCode, error)
	ListTaxCodesByTypeFunc         func(ctx context.Context, organizationID uuidv7.UUID, taxType taxcodeAggregate.TaxType) ([]*taxcodeAggregate.TaxCode, error)
	ListActiveTaxCodesFunc         func(ctx context.Context, organizationID uuidv7.UUID) ([]*taxcodeAggregate.TaxCode, error)
}

func (m *MockTaxCodeUseCase) CreateTaxCode(ctx context.Context, organizationID uuidv7.UUID, code, name string, taxType taxcodeAggregate.TaxType, rate int, createdBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.CreateTaxCodeFunc != nil {
		return m.CreateTaxCodeFunc(ctx, organizationID, code, name, taxType, rate, createdBy)
	}
	return nil, fmt.Errorf("CreateTaxCodeFunc not implemented")
}

func (m *MockTaxCodeUseCase) GetTaxCodeByID(ctx context.Context, id uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.GetTaxCodeByIDFunc != nil {
		return m.GetTaxCodeByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetTaxCodeByIDFunc not implemented")
}

func (m *MockTaxCodeUseCase) GetTaxCodeByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*taxcodeAggregate.TaxCode, error) {
	if m.GetTaxCodeByCodeFunc != nil {
		return m.GetTaxCodeByCodeFunc(ctx, organizationID, code)
	}
	return nil, fmt.Errorf("GetTaxCodeByCodeFunc not implemented")
}

func (m *MockTaxCodeUseCase) UpdateTaxCode(ctx context.Context, id uuidv7.UUID, rate int, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.UpdateTaxCodeFunc != nil {
		return m.UpdateTaxCodeFunc(ctx, id, rate, updatedBy)
	}
	return nil, fmt.Errorf("UpdateTaxCodeFunc not implemented")
}

func (m *MockTaxCodeUseCase) SetTaxPayableAccount(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.SetTaxPayableAccountFunc != nil {
		return m.SetTaxPayableAccountFunc(ctx, id, accountID, updatedBy)
	}
	return nil, fmt.Errorf("SetTaxPayableAccountFunc not implemented")
}

func (m *MockTaxCodeUseCase) SetTaxReceivableAccount(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.SetTaxReceivableAccountFunc != nil {
		return m.SetTaxReceivableAccountFunc(ctx, id, accountID, updatedBy)
	}
	return nil, fmt.Errorf("SetTaxReceivableAccountFunc not implemented")
}

func (m *MockTaxCodeUseCase) ActivateTaxCode(ctx context.Context, id, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.ActivateTaxCodeFunc != nil {
		return m.ActivateTaxCodeFunc(ctx, id, updatedBy)
	}
	return nil, fmt.Errorf("ActivateTaxCodeFunc not implemented")
}

func (m *MockTaxCodeUseCase) DeactivateTaxCode(ctx context.Context, id, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
	if m.DeactivateTaxCodeFunc != nil {
		return m.DeactivateTaxCodeFunc(ctx, id, updatedBy)
	}
	return nil, fmt.Errorf("DeactivateTaxCodeFunc not implemented")
}

func (m *MockTaxCodeUseCase) DeleteTaxCode(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteTaxCodeFunc != nil {
		return m.DeleteTaxCodeFunc(ctx, id)
	}
	return fmt.Errorf("DeleteTaxCodeFunc not implemented")
}

func (m *MockTaxCodeUseCase) ListTaxCodesByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*taxcodeAggregate.TaxCode, error) {
	if m.ListTaxCodesByOrganizationFunc != nil {
		return m.ListTaxCodesByOrganizationFunc(ctx, organizationID, limit, offset)
	}
	return nil, fmt.Errorf("ListTaxCodesByOrganizationFunc not implemented")
}

func (m *MockTaxCodeUseCase) ListTaxCodesByType(ctx context.Context, organizationID uuidv7.UUID, taxType taxcodeAggregate.TaxType) ([]*taxcodeAggregate.TaxCode, error) {
	if m.ListTaxCodesByTypeFunc != nil {
		return m.ListTaxCodesByTypeFunc(ctx, organizationID, taxType)
	}
	return nil, fmt.Errorf("ListTaxCodesByTypeFunc not implemented")
}

func (m *MockTaxCodeUseCase) ListActiveTaxCodes(ctx context.Context, organizationID uuidv7.UUID) ([]*taxcodeAggregate.TaxCode, error) {
	if m.ListActiveTaxCodesFunc != nil {
		return m.ListActiveTaxCodesFunc(ctx, organizationID)
	}
	return nil, fmt.Errorf("ListActiveTaxCodesFunc not implemented")
}

func setupRouter(mockUC *MockTaxCodeUseCase) *gin.Engine {
	router := smoke.SetupRouter()
	handler := taxcodeHTTP.NewTaxCodeHandler(mockUC)

	router.Use(func(c *gin.Context) {
		c.Set("organization_id", uuidv7.New().String())
		c.Set("user_id", uuidv7.New().String())
		c.Next()
	})

	handler.RegisterRoutes(router.Group("/api/v1/accounting"))
	return router
}

func TestCreateTaxCode_Success(t *testing.T) {
	taxCodeID := uuidv7.New()

	mockUC := &MockTaxCodeUseCase{
		CreateTaxCodeFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, taxType taxcodeAggregate.TaxType, rate int, createdBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			tc, _ := taxcodeAggregate.NewTaxCode(organizationID, code, name, taxType, rate, createdBy)
			tc.ID = taxCodeID
			return tc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":     "VAT20",
		"name":     "VAT 20%",
		"tax_type": "vat",
		"rate":     2000,
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/tax-codes", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateTaxCode_ValidationError(t *testing.T) {
	mockUC := &MockTaxCodeUseCase{
		CreateTaxCodeFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, taxType taxcodeAggregate.TaxType, rate int, createdBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return nil, taxcode.ErrTaxCodeEmpty
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":     "",
		"name":     "VAT 20%",
		"tax_type": "vat",
		"rate":     2000,
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/tax-codes", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTaxCode_Success(t *testing.T) {
	taxCodeID := uuidv7.New()

	tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", taxcodeAggregate.TaxTypeVAT, 2000, uuidv7.New())
	tc.ID = taxCodeID

	mockUC := &MockTaxCodeUseCase{
		GetTaxCodeByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return tc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/tax-codes/"+taxCodeID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTaxCode_NotFound(t *testing.T) {
	taxCodeID := uuidv7.New()

	mockUC := &MockTaxCodeUseCase{
		GetTaxCodeByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return nil, taxcode.ErrTaxCodeNotFound
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/tax-codes/"+taxCodeID.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateTaxCode_Success(t *testing.T) {
	taxCodeID := uuidv7.New()

	tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", taxcodeAggregate.TaxTypeVAT, 1800, uuidv7.New())
	tc.ID = taxCodeID

	mockUC := &MockTaxCodeUseCase{
		UpdateTaxCodeFunc: func(ctx context.Context, id uuidv7.UUID, rate int, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return tc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"rate": 1800,
	}

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/tax-codes/"+taxCodeID.String(), body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetTaxPayableAccount_Success(t *testing.T) {
	taxCodeID := uuidv7.New()
	accountID := uuidv7.New()

	tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", taxcodeAggregate.TaxTypeVAT, 2000, uuidv7.New())
	tc.ID = taxCodeID

	mockUC := &MockTaxCodeUseCase{
		SetTaxPayableAccountFunc: func(ctx context.Context, id, accID uuidv7.UUID, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return tc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"account_id": accountID.String(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/tax-codes/"+taxCodeID.String()+"/payable-account", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestActivateTaxCode_Success(t *testing.T) {
	taxCodeID := uuidv7.New()

	tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", taxcodeAggregate.TaxTypeVAT, 2000, uuidv7.New())
	tc.ID = taxCodeID

	mockUC := &MockTaxCodeUseCase{
		ActivateTaxCodeFunc: func(ctx context.Context, id, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return tc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/tax-codes/"+taxCodeID.String()+"/activate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeactivateTaxCode_Success(t *testing.T) {
	taxCodeID := uuidv7.New()

	tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", taxcodeAggregate.TaxTypeVAT, 2000, uuidv7.New())
	tc.ID = taxCodeID

	mockUC := &MockTaxCodeUseCase{
		DeactivateTaxCodeFunc: func(ctx context.Context, id, updatedBy uuidv7.UUID) (*taxcodeAggregate.TaxCode, error) {
			return tc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/tax-codes/"+taxCodeID.String()+"/deactivate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListTaxCodes_Success(t *testing.T) {
	taxCodes := []*taxcodeAggregate.TaxCode{
		func() *taxcodeAggregate.TaxCode {
			tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", taxcodeAggregate.TaxTypeVAT, 2000, uuidv7.New())
			return tc
		}(),
		func() *taxcodeAggregate.TaxCode {
			tc, _ := taxcodeAggregate.NewTaxCode(uuidv7.New(), "IT15", "Income Tax 15%", taxcodeAggregate.TaxTypeIncomeTax, 1500, uuidv7.New())
			return tc
		}(),
	}

	mockUC := &MockTaxCodeUseCase{
		ListTaxCodesByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*taxcodeAggregate.TaxCode, error) {
			return taxCodes, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/tax-codes", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}
