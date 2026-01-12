package company

import (
	"errors"
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Factory Methods Tests
// ============================================================================

func TestNewCompany_Success(t *testing.T) {
	company, err := NewCompany("Acme Corp", "llc")

	assert.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, company.ID)
	assert.Equal(t, "Acme Corp", company.Name)
	assert.Nil(t, company.LegalName)
	assert.Equal(t, CompanyTypeLLC, company.Type)
	assert.Nil(t, company.TaxID)
	assert.Nil(t, company.RegistrationNumber)
	assert.Nil(t, company.Website)
	assert.Nil(t, company.Email)
	assert.Nil(t, company.Phone)
	assert.Nil(t, company.Address)
	assert.Nil(t, company.Industry)
	assert.Equal(t, CompanySizeMicro, company.Size)
	assert.Equal(t, 0, company.EmployeeCount)
	assert.Equal(t, int64(0), company.Revenue)
	assert.Equal(t, "USD", company.Currency)
	assert.Nil(t, company.Description)
	assert.Nil(t, company.ParentCompanyID)
	assert.False(t, company.CreatedAt.IsZero())
	assert.False(t, company.UpdatedAt.IsZero())
	assert.Nil(t, company.DeletedAt)
}

func TestNewCompany_EmptyName(t *testing.T) {
	company, err := NewCompany("", "llc")

	assert.Error(t, err)
	assert.Nil(t, company)
	assert.True(t, errors.Is(err, ErrCompanyNameRequired))
}

func TestNewCompany_InvalidType(t *testing.T) {
	tests := []struct {
		name         string
		companyType  string
		expectError  bool
		expectedType CompanyType
	}{
		{"valid LLC", "llc", false, CompanyTypeLLC},
		{"valid Corporation", "corporation", false, CompanyTypeCorporation},
		{"valid Sole Proprietor", "sole_proprietor", false, CompanyTypeSoleProprietor},
		{"valid Partnership", "partnership", false, CompanyTypePartnership},
		{"valid Non-Profit", "non_profit", false, CompanyTypeNonProfit},
		{"valid Other", "other", false, CompanyTypeOther},
		{"invalid type", "invalid", true, ""},
		{"empty type", "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company, err := NewCompany("Test Corp", tt.companyType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, company)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, company)
				assert.Equal(t, tt.expectedType, company.Type)
			}
		})
	}
}

// ============================================================================
// Business Logic Tests
// ============================================================================

func TestCompany_UpdateBasicInfo(t *testing.T) {
	company, _ := NewCompany("Original Name", "llc")
	legalName := "Original Legal Name LLC"
	taxID := "12-3456789"

	err := company.UpdateBasicInfo("Updated Name", &legalName, "corporation", &taxID, nil)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", company.Name)
	assert.Equal(t, "Original Legal Name LLC", *company.LegalName)
	assert.Equal(t, CompanyTypeCorporation, company.Type)
	assert.Equal(t, "12-3456789", *company.TaxID)
	assert.False(t, company.UpdatedAt.IsZero())
}

func TestCompany_UpdateBasicInfo_EmptyName(t *testing.T) {
	company, _ := NewCompany("Original Name", "llc")

	err := company.UpdateBasicInfo("", nil, "llc", nil, nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompanyNameRequired))
	assert.Equal(t, "Original Name", company.Name)
}

func TestCompany_UpdateContactInfo(t *testing.T) {
	company, _ := NewCompany("Acme Corp", "llc")

	// Create contact info
	email, _ := valueobject.NewEmail("info@acme.com")
	phone, _ := valueobject.NewPhone("+380501234567")
	address, _ := valueobject.NewAddress("123 Main St", "Kyiv", "01001", "UA")
	website := "https://acme.com"

	company.UpdateContactInfo(&website, &email, &phone, &address)

	assert.NotNil(t, company.Website)
	assert.Equal(t, "https://acme.com", *company.Website)
	assert.NotNil(t, company.Email)
	assert.Equal(t, "info@acme.com", company.Email.Value())
	assert.NotNil(t, company.Phone)
	assert.Equal(t, "+380501234567", company.Phone.Value())
	assert.NotNil(t, company.Address)
	assert.Equal(t, "123 Main St", company.Address.Street)
	assert.False(t, company.UpdatedAt.IsZero())
}

func TestCompany_UpdateBusinessInfo(t *testing.T) {
	company, _ := NewCompany("Acme Corp", "llc")
	industry := "Technology"

	err := company.UpdateBusinessInfo(&industry, "large", 500, 10000000, "USD")

	assert.NoError(t, err)
	assert.Equal(t, "Technology", *company.Industry)
	assert.Equal(t, CompanySizeLarge, company.Size)
	assert.Equal(t, 500, company.EmployeeCount)
	assert.Equal(t, int64(10000000), company.Revenue)
	assert.Equal(t, "USD", company.Currency)
	assert.False(t, company.UpdatedAt.IsZero())
}

func TestCompany_UpdateBusinessInfo_InvalidSize(t *testing.T) {
	company, _ := NewCompany("Acme Corp", "llc")

	err := company.UpdateBusinessInfo(nil, "invalid", 10, 100000, "USD")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompanySizeInvalid))
	assert.Equal(t, CompanySizeMicro, company.Size)
}

func TestCompany_UpdateBusinessInfo_NegativeValues(t *testing.T) {
	company, _ := NewCompany("Acme Corp", "llc")

	tests := []struct {
		name          string
		employeeCount int
		revenue       int64
		expectError   bool
	}{
		{"negative employees", -1, 100000, true},
		{"negative revenue", 10, -1000, true},
		{"both negative", -5, -5000, true},
		{"zero values", 0, 0, false},
		{"positive values", 100, 1000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := company.UpdateBusinessInfo(nil, "small", tt.employeeCount, tt.revenue, "USD")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompany_SetParentCompany(t *testing.T) {
	company, _ := NewCompany("Subsidiary Corp", "llc")
	parentID := uuidv7.New()

	_ = company.SetParentCompany(&parentID)

	assert.NotNil(t, company.ParentCompanyID)
	assert.Equal(t, parentID, *company.ParentCompanyID)
	assert.False(t, company.UpdatedAt.IsZero())
}

// ============================================================================
// CompanyType Tests
// ============================================================================

func TestCompanyType_String(t *testing.T) {
	tests := []struct {
		ctype    CompanyType
		expected string
	}{
		{CompanyTypeLLC, "llc"},
		{CompanyTypeCorporation, "corporation"},
		{CompanyTypeSoleProprietor, "sole_proprietor"},
		{CompanyTypePartnership, "partnership"},
		{CompanyTypeNonProfit, "non_profit"},
		{CompanyTypeOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.ctype))
		})
	}
}

// ============================================================================
// CompanySize Tests
// ============================================================================

func TestCompanySize_String(t *testing.T) {
	tests := []struct {
		size     CompanySize
		expected string
	}{
		{CompanySizeMicro, "micro"},
		{CompanySizeSmall, "small"},
		{CompanySizeMedium, "medium"},
		{CompanySizeLarge, "large"},
		{CompanySizeEnterprise, "enterprise"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.size))
		})
	}
}
