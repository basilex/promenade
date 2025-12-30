package company

import (
	"context"
	"errors"
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of IRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, company *Company) error {
	args := m.Called(ctx, company)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Company, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Company), args.Error(1)
}

func (m *MockRepository) GetByName(ctx context.Context, name string) (*Company, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Company), args.Error(1)
}

func (m *MockRepository) GetByTaxID(ctx context.Context, taxID string) (*Company, error) {
	args := m.Called(ctx, taxID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Company), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*Company, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Company), args.Int(1), args.Error(2)
}

func (m *MockRepository) ListByIndustry(ctx context.Context, industry string, limit, offset int) ([]*Company, int, error) {
	args := m.Called(ctx, industry, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Company), args.Int(1), args.Error(2)
}

func (m *MockRepository) ListBySize(ctx context.Context, size string, limit, offset int) ([]*Company, int, error) {
	args := m.Called(ctx, size, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Company), args.Int(1), args.Error(2)
}

func (m *MockRepository) ListSubsidiaries(ctx context.Context, parentID uuidv7.UUID) ([]*Company, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Company), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, company *Company) error {
	args := m.Called(ctx, company)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Exists(ctx context.Context, id uuidv7.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ExistsByTaxID(ctx context.Context, taxID string) (bool, error) {
	args := m.Called(ctx, taxID)
	return args.Bool(0), args.Error(1)
}

// ============================================================================
// CreateCompany Tests
// ============================================================================

func TestUseCase_CreateCompany_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	legalName := "Acme Corporation LLC"
	taxID := "12-3456789"
	website := "https://acme.com"
	industry := "Technology"
	description := "Software company"

	email, _ := valueobject.NewEmail("info@acme.com")
	phone, _ := valueobject.NewPhone("+380501234567")
	address, _ := valueobject.NewAddress("123 Main St", "Kyiv", "01001", "UA")

	mockRepo.On("ExistsByName", ctx, "Acme Corp").Return(false, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*company.Company")).Return(nil)

	company, err := uc.CreateCompany(
		ctx,
		"Acme Corp",
		&legalName,
		"llc",
		&taxID,
		nil,
		&website,
		&email,
		&phone,
		&address,
		&industry,
		"large",
		500,
		10000000,
		"USD",
		&description,
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, "Acme Corp", company.Name)
	assert.Equal(t, "Acme Corporation LLC", *company.LegalName)
	assert.Equal(t, CompanyTypeLLC, company.Type)
	assert.Equal(t, CompanySizeLarge, company.Size)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_CreateCompany_DuplicateName(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("ExistsByName", ctx, "Acme Corp").Return(true, nil)

	company, err := uc.CreateCompany(
		ctx,
		"Acme Corp",
		nil,
		"llc",
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"small",
		0,
		0,
		"USD",
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompanyAlreadyExists))
	assert.Nil(t, company)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_CreateCompany_InvalidSize(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("ExistsByName", ctx, "Acme Corp").Return(false, nil)

	// Try to create company with invalid size
	company, err := uc.CreateCompany(
		ctx,
		"Acme Corp",
		nil,
		"llc",
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"invalid-size",
		0,
		0,
		"USD",
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, company)
	assert.Contains(t, err.Error(), "invalid company size")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// GetCompany Tests
// ============================================================================

func TestUseCase_GetCompany_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()
	expectedCompany, _ := NewCompany("Acme Corp", "llc")
	expectedCompany.ID = companyID

	mockRepo.On("GetByID", ctx, companyID).Return(expectedCompany, nil)

	company, err := uc.GetCompany(ctx, companyID)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, companyID, company.ID)
	assert.Equal(t, "Acme Corp", company.Name)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_GetCompany_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()

	mockRepo.On("GetByID", ctx, companyID).Return(nil, ErrCompanyNotFound)

	company, err := uc.GetCompany(ctx, companyID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompanyNotFound))
	assert.Nil(t, company)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// UpdateCompanyBasicInfo Tests
// ============================================================================

func TestUseCase_UpdateCompanyBasicInfo_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()
	existingCompany, _ := NewCompany("Old Name", "llc")
	existingCompany.ID = companyID

	legalName := "New Legal Name LLC"
	taxID := "98-7654321"

	mockRepo.On("GetByID", ctx, companyID).Return(existingCompany, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*company.Company")).Return(nil)

	company, err := uc.UpdateCompanyBasicInfo(
		ctx,
		companyID,
		"New Name",
		&legalName,
		"corporation",
		&taxID,
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, "New Name", company.Name)
	assert.Equal(t, "New Legal Name LLC", *company.LegalName)
	assert.Equal(t, CompanyTypeCorporation, company.Type)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_UpdateCompanyBasicInfo_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()

	mockRepo.On("GetByID", ctx, companyID).Return(nil, ErrCompanyNotFound)

	company, err := uc.UpdateCompanyBasicInfo(
		ctx,
		companyID,
		"New Name",
		nil,
		"llc",
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompanyNotFound))
	assert.Nil(t, company)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// UpdateCompanyContactInfo Tests
// ============================================================================

func TestUseCase_UpdateCompanyContactInfo_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()
	existingCompany, _ := NewCompany("Acme Corp", "llc")
	existingCompany.ID = companyID

	email, _ := valueobject.NewEmail("contact@acme.com")
	phone, _ := valueobject.NewPhone("+380501234567")
	address, _ := valueobject.NewAddress("456 Oak Ave", "Kyiv", "02002", "UA")
	website := "https://acme.com"

	mockRepo.On("GetByID", ctx, companyID).Return(existingCompany, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*company.Company")).Return(nil)

	company, err := uc.UpdateCompanyContactInfo(ctx, companyID, &website, &email, &phone, &address)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, "contact@acme.com", company.Email.Value())
	assert.Equal(t, "+380501234567", company.Phone.Value())
	assert.Equal(t, "456 Oak Ave", company.Address.Street)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// UpdateCompanyBusinessInfo Tests
// ============================================================================

func TestUseCase_UpdateCompanyBusinessInfo_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()
	existingCompany, _ := NewCompany("Acme Corp", "llc")
	existingCompany.ID = companyID

	industry := "Finance"

	mockRepo.On("GetByID", ctx, companyID).Return(existingCompany, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*company.Company")).Return(nil)

	company, err := uc.UpdateCompanyBusinessInfo(
		ctx,
		companyID,
		&industry,
		"enterprise",
		1000,
		50000000,
		"EUR",
	)

	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, "Finance", *company.Industry)
	assert.Equal(t, CompanySizeEnterprise, company.Size)
	assert.Equal(t, 1000, company.EmployeeCount)
	assert.Equal(t, int64(50000000), company.Revenue)
	assert.Equal(t, "EUR", company.Currency)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// DeleteCompany Tests
// ============================================================================

func TestUseCase_DeleteCompany_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()
	existingCompany, _ := NewCompany("Acme Corp", "llc")
	existingCompany.ID = companyID

	mockRepo.On("GetByID", ctx, companyID).Return(existingCompany, nil)
	mockRepo.On("Delete", ctx, companyID).Return(nil)

	err := uc.DeleteCompany(ctx, companyID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUseCase_DeleteCompany_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	companyID := uuidv7.New()

	mockRepo.On("GetByID", ctx, companyID).Return(nil, ErrCompanyNotFound)

	err := uc.DeleteCompany(ctx, companyID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompanyNotFound))
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// ListCompanies Tests
// ============================================================================

func TestUseCase_ListCompanies_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	company1, _ := NewCompany("Company A", "llc")
	company2, _ := NewCompany("Company B", "corporation")
	companies := []*Company{company1, company2}

	mockRepo.On("List", ctx, 1, 10).Return(companies, 2, nil)

	result, total, err := uc.ListCompanies(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// ListCompaniesByIndustry Tests
// ============================================================================

func TestUseCase_ListCompaniesByIndustry_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	company1, _ := NewCompany("Tech A", "llc")
	industry := "Technology"
	company1.Industry = &industry

	companies := []*Company{company1}

	mockRepo.On("ListByIndustry", ctx, "Technology", 1, 20).Return(companies, 1, nil)

	result, total, err := uc.ListCompaniesByIndustry(ctx, "Technology", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// ListCompaniesBySize Tests
// ============================================================================

func TestUseCase_ListCompaniesBySize_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	company1, _ := NewCompany("Large Corp", "llc")
	companies := []*Company{company1}

	mockRepo.On("ListBySize", ctx, "large", 1, 20).Return(companies, 1, nil)

	result, total, err := uc.ListCompaniesBySize(ctx, "large", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// ListSubsidiaries Tests
// ============================================================================

func TestUseCase_ListSubsidiaries_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	parentID := uuidv7.New()
	subsidiary1, _ := NewCompany("Subsidiary A", "llc")
	subsidiary2, _ := NewCompany("Subsidiary B", "llc")
	subsidiaries := []*Company{subsidiary1, subsidiary2}

	mockRepo.On("ListSubsidiaries", ctx, parentID).Return(subsidiaries, nil)

	result, err := uc.ListSubsidiaries(ctx, parentID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	mockRepo.AssertExpectations(t)
}

func TestUseCase_ListSubsidiaries_Empty(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	parentID := uuidv7.New()

	mockRepo.On("ListSubsidiaries", ctx, parentID).Return([]*Company{}, nil)

	result, err := uc.ListSubsidiaries(ctx, parentID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	mockRepo.AssertExpectations(t)
}
