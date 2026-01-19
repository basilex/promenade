package company_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	companyHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/http"
	companyAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/smoke"
)

// MockCompanyUseCase implements company.IUseCase for smoke testing
type MockCompanyUseCase struct {
	CreateCompanyFunc             func(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*companyAggregate.Company, error)
	GetCompanyFunc                func(ctx context.Context, id uuidv7.UUID) (*companyAggregate.Company, error)
	GetCompanyByNameFunc          func(ctx context.Context, name string) (*companyAggregate.Company, error)
	GetCompanyByTaxIDFunc         func(ctx context.Context, taxID string) (*companyAggregate.Company, error)
	ListCompaniesFunc             func(ctx context.Context, page, pageSize int) ([]*companyAggregate.Company, int, error)
	ListCompaniesByIndustryFunc   func(ctx context.Context, industry string, page, pageSize int) ([]*companyAggregate.Company, int, error)
	ListCompaniesBySizeFunc       func(ctx context.Context, size string, page, pageSize int) ([]*companyAggregate.Company, int, error)
	ListSubsidiariesFunc          func(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*companyAggregate.Company, error)
	UpdateCompanyBasicInfoFunc    func(ctx context.Context, id uuidv7.UUID, name string, legalName *string, companyType string, taxID *string, registrationNumber *string) (*companyAggregate.Company, error)
	UpdateCompanyContactInfoFunc  func(ctx context.Context, id uuidv7.UUID, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) (*companyAggregate.Company, error)
	UpdateCompanyBusinessInfoFunc func(ctx context.Context, id uuidv7.UUID, industry *string, size string, employeeCount int, revenue int64, currency string) (*companyAggregate.Company, error)
	SetParentCompanyFunc          func(ctx context.Context, id uuidv7.UUID, parentCompanyID *uuidv7.UUID) (*companyAggregate.Company, error)
	UpdateCompanyDescriptionFunc  func(ctx context.Context, id uuidv7.UUID, description *string) (*companyAggregate.Company, error)
	DeleteCompanyFunc             func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockCompanyUseCase) CreateCompany(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*companyAggregate.Company, error) {
	return m.CreateCompanyFunc(ctx, name, legalName, companyType, taxID, registrationNumber, website, email, phone, address, industry, size, employeeCount, revenue, currency, description, parentCompanyID)
}

func (m *MockCompanyUseCase) GetCompany(ctx context.Context, id uuidv7.UUID) (*companyAggregate.Company, error) {
	return m.GetCompanyFunc(ctx, id)
}

func (m *MockCompanyUseCase) GetCompanyByName(ctx context.Context, name string) (*companyAggregate.Company, error) {
	return m.GetCompanyByNameFunc(ctx, name)
}

func (m *MockCompanyUseCase) GetCompanyByTaxID(ctx context.Context, taxID string) (*companyAggregate.Company, error) {
	return m.GetCompanyByTaxIDFunc(ctx, taxID)
}

func (m *MockCompanyUseCase) ListCompanies(ctx context.Context, page, pageSize int) ([]*companyAggregate.Company, int, error) {
	return m.ListCompaniesFunc(ctx, page, pageSize)
}

func (m *MockCompanyUseCase) ListCompaniesByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*companyAggregate.Company, int, error) {
	return m.ListCompaniesByIndustryFunc(ctx, industry, page, pageSize)
}

func (m *MockCompanyUseCase) ListCompaniesBySize(ctx context.Context, size string, page, pageSize int) ([]*companyAggregate.Company, int, error) {
	return m.ListCompaniesBySizeFunc(ctx, size, page, pageSize)
}

func (m *MockCompanyUseCase) ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*companyAggregate.Company, error) {
	return m.ListSubsidiariesFunc(ctx, parentCompanyID)
}

func (m *MockCompanyUseCase) UpdateCompanyBasicInfo(ctx context.Context, id uuidv7.UUID, name string, legalName *string, companyType string, taxID *string, registrationNumber *string) (*companyAggregate.Company, error) {
	return m.UpdateCompanyBasicInfoFunc(ctx, id, name, legalName, companyType, taxID, registrationNumber)
}

func (m *MockCompanyUseCase) UpdateCompanyContactInfo(ctx context.Context, id uuidv7.UUID, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) (*companyAggregate.Company, error) {
	return m.UpdateCompanyContactInfoFunc(ctx, id, website, email, phone, address)
}

func (m *MockCompanyUseCase) UpdateCompanyBusinessInfo(ctx context.Context, id uuidv7.UUID, industry *string, size string, employeeCount int, revenue int64, currency string) (*companyAggregate.Company, error) {
	return m.UpdateCompanyBusinessInfoFunc(ctx, id, industry, size, employeeCount, revenue, currency)
}

func (m *MockCompanyUseCase) SetParentCompany(ctx context.Context, id uuidv7.UUID, parentCompanyID *uuidv7.UUID) (*companyAggregate.Company, error) {
	return m.SetParentCompanyFunc(ctx, id, parentCompanyID)
}

func (m *MockCompanyUseCase) UpdateCompanyDescription(ctx context.Context, id uuidv7.UUID, description *string) (*companyAggregate.Company, error) {
	return m.UpdateCompanyDescriptionFunc(ctx, id, description)
}

func (m *MockCompanyUseCase) DeleteCompany(ctx context.Context, id uuidv7.UUID) error {
	return m.DeleteCompanyFunc(ctx, id)
}

// fakeCompany creates a fake company for testing
func fakeCompany() *companyAggregate.Company {
	legalName := "Acme Corporation LLC"
	comp, _ := companyAggregate.NewCompany("Acme Corp", "corporation")
	comp.LegalName = &legalName
	return comp
}

// Test Create - Success
func TestCompanyHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		CreateCompanyFunc: func(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*companyAggregate.Company, error) {
			return fakeCompany(), nil
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.POST("/companies", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/companies", map[string]any{
		"name":           "Acme Corp",
		"legal_name":     "Acme Corporation LLC",
		"type":           "corporation",
		"size":           "medium",
		"currency":       "USD",
		"employee_count": 100,
		"revenue":        1000000,
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

// Test Create - Validation Error
func TestCompanyHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{}
	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.POST("/companies", handler.Create)

	// Missing required field 'name'
	resp := smoke.MakeRequest(t, router, "POST", "/companies", map[string]any{
		"type": "corporation",
	})

	smoke.AssertErrorResponse(t, resp, 400, "VALIDATION_ERROR")
}

// Test GetByID - Success
func TestCompanyHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		GetCompanyFunc: func(ctx context.Context, id uuidv7.UUID) (*companyAggregate.Company, error) {
			return fakeCompany(), nil
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.GET("/companies/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/companies/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

// Test GetByID - Not Found
func TestCompanyHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		GetCompanyFunc: func(ctx context.Context, id uuidv7.UUID) (*companyAggregate.Company, error) {
			return nil, company.ErrCompanyNotFound
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.GET("/companies/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/companies/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, 404, "COMPANY_NOT_FOUND")
}

// Test List - Success
func TestCompanyHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		ListCompaniesFunc: func(ctx context.Context, page, pageSize int) ([]*companyAggregate.Company, int, error) {
			return []*companyAggregate.Company{fakeCompany()}, 1, nil
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.GET("/companies", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/companies", nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

// Test List - Empty Result
func TestCompanyHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		ListCompaniesFunc: func(ctx context.Context, page, pageSize int) ([]*companyAggregate.Company, int, error) {
			return []*companyAggregate.Company{}, 0, nil
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.GET("/companies", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/companies", nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

// Test Delete - Success
func TestCompanyHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		DeleteCompanyFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.DELETE("/companies/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/companies/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

// Test Delete - Not Found
func TestCompanyHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCompanyUseCase{
		DeleteCompanyFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return company.ErrCompanyNotFound
		},
	}

	handler := companyHTTP.NewCompanyHandler(mockUC)
	router.DELETE("/companies/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/companies/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, 404, "COMPANY_NOT_FOUND")
}
