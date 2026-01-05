package company_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	companyRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func setupCompanyUseCase(db *sqlx.DB) company.IUseCase {
	repo := companyRepo.NewCompanyRepository(db)
	return company.NewUseCase(repo)
}

// TestCompanyUseCase_CreateCompany tests basic company creation
func TestCompanyUseCase_CreateCompany(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company with minimal fields
		comp, err := uc.CreateCompany(ctx, "Acme Corp", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "medium", 50, 1000000, "USD", nil, nil)
		require.NoError(t, err)
		require.NotNil(t, comp)
		assert.NotEqual(t, uuidv7.UUID{}, comp.ID)
		assert.Equal(t, "Acme Corp", comp.Name)
		assert.Equal(t, company.CompanyType("llc"), comp.Type)
		assert.Equal(t, company.CompanySize("medium"), comp.Size)
		assert.Equal(t, 50, comp.EmployeeCount)
		assert.Equal(t, int64(1000000), comp.Revenue)
		assert.Equal(t, "USD", comp.Currency)

		// Test duplicate name error
		_, err = uc.CreateCompany(ctx, "Acme Corp", nil, "corporation", nil, nil, nil, nil, nil, nil, nil, "large", 200, 5000000, "USD", nil, nil)
		assert.ErrorIs(t, err, company.ErrCompanyAlreadyExists)
	})
}

// TestCompanyUseCase_CreateCompanyWithFullData tests company creation with all fields
func TestCompanyUseCase_CreateCompanyWithFullData(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		legalName := "Tech Solutions Limited"
		taxID := "US-12345678"
		regNumber := "REG-2024-001"
		website := "https://techsolutions.com"
		email, _ := valueobject.NewEmail("info@techsolutions.com")
		phone, _ := valueobject.NewPhone("+1234567890")
		address, _ := valueobject.NewAddress("123 Tech St", "San Francisco", "94105", "US")
		industry := "technology"
		description := "Leading provider of cloud solutions"

		comp, err := uc.CreateCompany(ctx, "Tech Solutions Inc", &legalName, "corporation", &taxID, &regNumber, &website, &email, &phone, &address, &industry, "large", 500, 50000000, "USD", &description, nil)
		require.NoError(t, err)
		require.NotNil(t, comp)

		assert.Equal(t, "Tech Solutions Inc", comp.Name)
		assert.Equal(t, "Tech Solutions Limited", *comp.LegalName)
		assert.Equal(t, company.CompanyType("corporation"), comp.Type)
		assert.Equal(t, "US-12345678", *comp.TaxID)
		assert.Equal(t, "REG-2024-001", *comp.RegistrationNumber)
		assert.Equal(t, "https://techsolutions.com", *comp.Website)
		assert.Equal(t, "info@techsolutions.com", comp.Email.Value())
		assert.Equal(t, "+1234567890", comp.Phone.Value())
		assert.Equal(t, "123 Tech St", comp.Address.Street)
		assert.Equal(t, "San Francisco", comp.Address.City)
		assert.Equal(t, "94105", comp.Address.PostalCode)
		assert.Equal(t, "US", comp.Address.Country)
		assert.Equal(t, "technology", *comp.Industry)
		assert.Equal(t, company.CompanySize("large"), comp.Size)
		assert.Equal(t, 500, comp.EmployeeCount)
		assert.Equal(t, int64(50000000), comp.Revenue)
		assert.Equal(t, "Leading provider of cloud solutions", *comp.Description)
	})
}

// TestCompanyUseCase_GetCompany tests company retrieval by ID
func TestCompanyUseCase_GetCompany(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company
		created, err := uc.CreateCompany(ctx, "Global Services", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 20, 500000, "USD", nil, nil)
		require.NoError(t, err)

		// Get by ID
		fetched, err := uc.GetCompany(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, "Global Services", fetched.Name)

		// Test not found error
		nonExistentID := uuidv7.New()
		_, err = uc.GetCompany(ctx, nonExistentID)
		assert.ErrorIs(t, err, company.ErrCompanyNotFound)
	})
}

// TestCompanyUseCase_GetCompanyByName tests company retrieval by name
func TestCompanyUseCase_GetCompanyByName(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company
		_, err := uc.CreateCompany(ctx, "Innovative Systems", nil, "corporation", nil, nil, nil, nil, nil, nil, nil, "medium", 100, 2000000, "USD", nil, nil)
		require.NoError(t, err)

		// Get by name
		fetched, err := uc.GetCompanyByName(ctx, "Innovative Systems")
		require.NoError(t, err)
		assert.Equal(t, "Innovative Systems", fetched.Name)

		// Test not found error
		_, err = uc.GetCompanyByName(ctx, "Non-existent Company")
		assert.ErrorIs(t, err, company.ErrCompanyNotFound)
	})
}

// TestCompanyUseCase_GetCompanyByTaxID tests company retrieval by tax ID
func TestCompanyUseCase_GetCompanyByTaxID(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		taxID := "TAX-987654"

		// Create company with tax ID
		_, err := uc.CreateCompany(ctx, "Taxable Enterprises", nil, "llc", &taxID, nil, nil, nil, nil, nil, nil, "small", 15, 300000, "USD", nil, nil)
		require.NoError(t, err)

		// Get by tax ID
		fetched, err := uc.GetCompanyByTaxID(ctx, "TAX-987654")
		require.NoError(t, err)
		assert.Equal(t, "TAX-987654", *fetched.TaxID)
		assert.Equal(t, "Taxable Enterprises", fetched.Name)

		// Test not found error
		_, err = uc.GetCompanyByTaxID(ctx, "NON-EXISTENT-TAX")
		assert.ErrorIs(t, err, company.ErrCompanyNotFound)
	})
}

// TestCompanyUseCase_UpdateCompanyBasicInfo tests updating basic company information
func TestCompanyUseCase_UpdateCompanyBasicInfo(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company
		comp, err := uc.CreateCompany(ctx, "Old Name Corp", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 10, 100000, "USD", nil, nil)
		require.NoError(t, err)

		// Update basic info
		newLegalName := "New Legal Name LLC"
		newTaxID := "NEW-TAX-123"
		newRegNumber := "NEW-REG-456"

		updated, err := uc.UpdateCompanyBasicInfo(ctx, comp.ID, "New Name Corp", &newLegalName, "corporation", &newTaxID, &newRegNumber)
		require.NoError(t, err)
		assert.Equal(t, "New Name Corp", updated.Name)
		assert.Equal(t, "New Legal Name LLC", *updated.LegalName)
		assert.Equal(t, company.CompanyType("corporation"), updated.Type)
		assert.Equal(t, "NEW-TAX-123", *updated.TaxID)
		assert.Equal(t, "NEW-REG-456", *updated.RegistrationNumber)
	})
}

// TestCompanyUseCase_UpdateCompanyContactInfo tests updating contact information
func TestCompanyUseCase_UpdateCompanyContactInfo(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company
		comp, err := uc.CreateCompany(ctx, "Contact Test Corp", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 5, 50000, "USD", nil, nil)
		require.NoError(t, err)

		// Update contact info
		website := "https://newwebsite.com"
		email, _ := valueobject.NewEmail("contact@newwebsite.com")
		phone, _ := valueobject.NewPhone("+9876543210")
		address, _ := valueobject.NewAddress("456 New St", "New York", "10001", "US")

		updated, err := uc.UpdateCompanyContactInfo(ctx, comp.ID, &website, &email, &phone, &address)
		require.NoError(t, err)
		assert.Equal(t, "https://newwebsite.com", *updated.Website)
		assert.Equal(t, "contact@newwebsite.com", updated.Email.Value())
		assert.Equal(t, "+9876543210", updated.Phone.Value())
		assert.Equal(t, "456 New St", updated.Address.Street)
		assert.Equal(t, "New York", updated.Address.City)
		assert.Equal(t, "10001", updated.Address.PostalCode)
		assert.Equal(t, "US", updated.Address.Country)
	})
}

// TestCompanyUseCase_UpdateCompanyBusinessInfo tests updating business metrics
func TestCompanyUseCase_UpdateCompanyBusinessInfo(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company
		industry := "retail"
		comp, err := uc.CreateCompany(ctx, "Growing Business", nil, "llc", nil, nil, nil, nil, nil, nil, &industry, "small", 10, 100000, "USD", nil, nil)
		require.NoError(t, err)

		// Update business info - scale up
		newIndustry := "e-commerce"
		updated, err := uc.UpdateCompanyBusinessInfo(ctx, comp.ID, &newIndustry, "medium", 50, 2000000, "USD")
		require.NoError(t, err)
		assert.Equal(t, "e-commerce", *updated.Industry)
		assert.Equal(t, company.CompanySize("medium"), updated.Size)
		assert.Equal(t, 50, updated.EmployeeCount)
		assert.Equal(t, int64(2000000), updated.Revenue)
	})
}

// TestCompanyUseCase_ParentChildRelationship tests parent-child company relationships
func TestCompanyUseCase_ParentChildRelationship(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create parent company
		parent, err := uc.CreateCompany(ctx, "Parent Holdings Inc", nil, "corporation", nil, nil, nil, nil, nil, nil, nil, "large", 1000, 100000000, "USD", nil, nil)
		require.NoError(t, err)

		// Create subsidiary with parent
		subsidiary, err := uc.CreateCompany(ctx, "Subsidiary LLC", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "medium", 100, 5000000, "USD", nil, &parent.ID)
		require.NoError(t, err)
		require.NotNil(t, subsidiary.ParentCompanyID)
		assert.Equal(t, parent.ID, *subsidiary.ParentCompanyID)

		// Create another company and link to parent
		independent, err := uc.CreateCompany(ctx, "Independent Corp", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 20, 500000, "USD", nil, nil)
		require.NoError(t, err)
		assert.Nil(t, independent.ParentCompanyID)

		linked, err := uc.SetParentCompany(ctx, independent.ID, &parent.ID)
		require.NoError(t, err)
		require.NotNil(t, linked.ParentCompanyID)
		assert.Equal(t, parent.ID, *linked.ParentCompanyID)

		// List subsidiaries
		subsidiaries, err := uc.ListSubsidiaries(ctx, parent.ID)
		require.NoError(t, err)
		assert.Len(t, subsidiaries, 2)

		// Unlink parent
		unlinked, err := uc.SetParentCompany(ctx, independent.ID, nil)
		require.NoError(t, err)
		assert.Nil(t, unlinked.ParentCompanyID)
	})
}

// TestCompanyUseCase_ListCompanies tests listing all companies with pagination
func TestCompanyUseCase_ListCompanies(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create 5 companies
		for i := 1; i <= 5; i++ {
			_, err := uc.CreateCompany(ctx, fmt.Sprintf("Company %d", i), nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 10, 100000, "USD", nil, nil)
			require.NoError(t, err)
		}

		// List first page
		companies, total, err := uc.ListCompanies(ctx, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, len(companies))
		assert.GreaterOrEqual(t, total, 5)

		// List second page
		companies, total, err = uc.ListCompanies(ctx, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, len(companies))
		assert.GreaterOrEqual(t, total, 5)

		// List all
		companies, total, err = uc.ListCompanies(ctx, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(companies), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
}

// TestCompanyUseCase_ListCompaniesByIndustry tests filtering by industry
func TestCompanyUseCase_ListCompaniesByIndustry(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		tech := "technology"
		retail := "retail"

		// Create tech companies
		_, err := uc.CreateCompany(ctx, "Tech Corp 1", nil, "llc", nil, nil, nil, nil, nil, nil, &tech, "small", 10, 100000, "USD", nil, nil)
		require.NoError(t, err)
		_, err = uc.CreateCompany(ctx, "Tech Corp 2", nil, "corporation", nil, nil, nil, nil, nil, nil, &tech, "medium", 50, 500000, "USD", nil, nil)
		require.NoError(t, err)

		// Create retail companies
		_, err = uc.CreateCompany(ctx, "Retail Store 1", nil, "llc", nil, nil, nil, nil, nil, nil, &retail, "small", 5, 50000, "USD", nil, nil)
		require.NoError(t, err)

		// List by technology industry
		techCompanies, total, err := uc.ListCompaniesByIndustry(ctx, "technology", 1, 10)
		require.NoError(t, err)
		assert.Len(t, techCompanies, 2)
		assert.Equal(t, 2, total)

		// List by retail industry
		retailCompanies, total, err := uc.ListCompaniesByIndustry(ctx, "retail", 1, 10)
		require.NoError(t, err)
		assert.Len(t, retailCompanies, 1)
		assert.Equal(t, 1, total)
	})
}

// TestCompanyUseCase_ListCompaniesBySize tests filtering by company size
func TestCompanyUseCase_ListCompaniesBySize(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create companies of different sizes
		_, err := uc.CreateCompany(ctx, "Small Co 1", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 5, 50000, "USD", nil, nil)
		require.NoError(t, err)
		_, err = uc.CreateCompany(ctx, "Small Co 2", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 10, 100000, "USD", nil, nil)
		require.NoError(t, err)
		_, err = uc.CreateCompany(ctx, "Medium Co", nil, "corporation", nil, nil, nil, nil, nil, nil, nil, "medium", 50, 500000, "USD", nil, nil)
		require.NoError(t, err)
		_, err = uc.CreateCompany(ctx, "Large Co", nil, "corporation", nil, nil, nil, nil, nil, nil, nil, "large", 500, 10000000, "USD", nil, nil)
		require.NoError(t, err)

		// List small companies
		companies, total, err := uc.ListCompaniesBySize(ctx, "small", 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(companies), 2)
		assert.GreaterOrEqual(t, total, 2)

		// List medium companies
		companies, total, err = uc.ListCompaniesBySize(ctx, "medium", 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(companies), 1)
		assert.GreaterOrEqual(t, total, 1)

		// List large companies
		companies, total, err = uc.ListCompaniesBySize(ctx, "large", 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(companies), 1)
		assert.GreaterOrEqual(t, total, 1)
	})
}

// TestCompanyUseCase_UpdateCompanyDescription tests updating company description
func TestCompanyUseCase_UpdateCompanyDescription(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company without description
		comp, err := uc.CreateCompany(ctx, "Description Test", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 5, 50000, "USD", nil, nil)
		require.NoError(t, err)
		assert.Nil(t, comp.Description)

		// Add description
		desc := "A leading provider of innovative solutions"
		updated, err := uc.UpdateCompanyDescription(ctx, comp.ID, &desc)
		require.NoError(t, err)
		require.NotNil(t, updated.Description)
		assert.Equal(t, "A leading provider of innovative solutions", *updated.Description)

		// Update description
		newDesc := "Updated company description with more details"
		updated, err = uc.UpdateCompanyDescription(ctx, comp.ID, &newDesc)
		require.NoError(t, err)
		assert.Equal(t, "Updated company description with more details", *updated.Description)

		// Clear description
		updated, err = uc.UpdateCompanyDescription(ctx, comp.ID, nil)
		require.NoError(t, err)
		assert.Nil(t, updated.Description)
	})
}

// TestCompanyUseCase_DeleteCompany tests company deletion
func TestCompanyUseCase_DeleteCompany(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create company
		comp, err := uc.CreateCompany(ctx, "To Be Deleted", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "small", 1, 10000, "USD", nil, nil)
		require.NoError(t, err)

		// Verify exists
		fetched, err := uc.GetCompany(ctx, comp.ID)
		require.NoError(t, err)
		assert.Equal(t, comp.ID, fetched.ID)

		// Delete
		err = uc.DeleteCompany(ctx, comp.ID)
		require.NoError(t, err)

		// Verify deleted
		_, err = uc.GetCompany(ctx, comp.ID)
		assert.ErrorIs(t, err, company.ErrCompanyNotFound)

		// Try to delete non-existent company
		err = uc.DeleteCompany(ctx, uuidv7.New())
		assert.ErrorIs(t, err, company.ErrCompanyNotFound)
	})
}

// TestCompanyUseCase_CompleteWorkflow tests end-to-end company lifecycle
func TestCompanyUseCase_CompleteWorkflow(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	uc := setupCompanyUseCase(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Step 1: Create company with minimal info
		comp, err := uc.CreateCompany(ctx, "Startup Inc", nil, "llc", nil, nil, nil, nil, nil, nil, nil, "micro", 5, 100000, "USD", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "Startup Inc", comp.Name)
		assert.Equal(t, company.CompanySize("micro"), comp.Size)
		assert.Equal(t, 5, comp.EmployeeCount)

		// Step 2: Update basic info as company formalizes
		legalName := "Startup Incorporated LLC"
		taxID := "STARTUP-TAX-001"
		regNumber := "REG-STARTUP-2024"
		comp, err = uc.UpdateCompanyBasicInfo(ctx, comp.ID, "Startup Inc", &legalName, "llc", &taxID, &regNumber)
		require.NoError(t, err)
		assert.Equal(t, "Startup Incorporated LLC", *comp.LegalName)
		assert.Equal(t, "STARTUP-TAX-001", *comp.TaxID)

		// Step 3: Add contact information
		website := "https://startup-inc.com"
		email, _ := valueobject.NewEmail("info@startup-inc.com")
		phone, _ := valueobject.NewPhone("+1555123456")
		address, _ := valueobject.NewAddress("100 Startup Blvd", "Palo Alto", "94301", "US")
		comp, err = uc.UpdateCompanyContactInfo(ctx, comp.ID, &website, &email, &phone, &address)
		require.NoError(t, err)
		assert.Equal(t, "https://startup-inc.com", *comp.Website)
		assert.Equal(t, "info@startup-inc.com", comp.Email.Value())

		// Step 4: Add description
		desc := "Innovative startup disrupting the market"
		comp, err = uc.UpdateCompanyDescription(ctx, comp.ID, &desc)
		require.NoError(t, err)
		assert.Equal(t, "Innovative startup disrupting the market", *comp.Description)

		// Step 5: Scale up business
		industry := "technology"
		comp, err = uc.UpdateCompanyBusinessInfo(ctx, comp.ID, &industry, "small", 25, 1000000, "USD")
		require.NoError(t, err)
		assert.Equal(t, company.CompanySize("small"), comp.Size)
		assert.Equal(t, 25, comp.EmployeeCount)
		assert.Equal(t, int64(1000000), comp.Revenue)

		// Step 6: Continue growth
		comp, err = uc.UpdateCompanyBusinessInfo(ctx, comp.ID, &industry, "medium", 100, 10000000, "USD")
		require.NoError(t, err)
		assert.Equal(t, company.CompanySize("medium"), comp.Size)
		assert.Equal(t, 100, comp.EmployeeCount)

		// Step 7: Verify all data persisted
		final, err := uc.GetCompany(ctx, comp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Startup Inc", final.Name)
		assert.Equal(t, company.CompanySize("medium"), final.Size)
		assert.Equal(t, 100, final.EmployeeCount)
		assert.Equal(t, "info@startup-inc.com", final.Email.Value())
		assert.Equal(t, "Innovative startup disrupting the market", *final.Description)
	})
}
