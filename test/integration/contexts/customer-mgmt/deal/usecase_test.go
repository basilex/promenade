package deal_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	companyRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal"
	dealRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/adapter/repository/postgres"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	_ "github.com/basilex/promenade/internal/contexts/identity/role"
	roleRepo "github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/user"
	userRepo "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestDealUseCase_CreateDeal tests deal creation
func TestDealUseCase_CreateDeal(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	// Setup
	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer with unique email
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)

	// Create test sales rep (using customer as user placeholder)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deal
	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 100000, "USD", "2026-12-31")
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, d.ID)
	assert.Equal(t, "Test Deal", d.Name)
	assert.Equal(t, testCustomer.ID, d.CustomerID)
	assert.Equal(t, salesRep.ID, d.AssignedTo)
	assert.Equal(t, int64(100000), d.Value.Amount)
	assert.Equal(t, "USD", d.Value.Currency)
	assert.Equal(t, deal.DealStageLead, d.Stage)
	assert.Equal(t, 10, d.Probability) // Lead stage = 10%
	assert.NotNil(t, d.ExpectedCloseDate)
}

// TestDealUseCase_CreateDealWithFullData tests deal creation with all optional fields
func TestDealUseCase_CreateDealWithFullData(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deal with full data
	d, err := dealUC.CreateDeal(ctx, "Full Deal", testCustomer.ID, salesRep.ID, 250000, "EUR", "2026-06-30")
	require.NoError(t, err)

	// Update with additional info
	d, err = dealUC.UpdateDealBasicInfo(ctx, d.ID, "Full Deal Updated", "Test description")
	require.NoError(t, err)
	assert.Equal(t, "Full Deal Updated", d.Name)
	assert.Equal(t, "Test description", d.Description)

	d, err = dealUC.SetDealSource(ctx, d.ID, deal.DealSourceReferral)
	require.NoError(t, err)
	assert.Equal(t, deal.DealSourceReferral, d.Source)
}

// TestDealUseCase_GetDeal tests retrieving a deal by ID
func TestDealUseCase_GetDeal(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deals with unique email
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	created, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Get deal
	retrieved, err := dealUC.GetDeal(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)
	assert.Equal(t, created.CustomerID, retrieved.CustomerID)
}

// TestDealUseCase_GetDeal_NotFound tests retrieving non-existent deal
func TestDealUseCase_GetDeal_NotFound(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)

	ctx := context.Background()

	// Try to get non-existent deal
	_, err := dealUC.GetDeal(ctx, uuidv7.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deal not found")
}

// TestDealUseCase_UpdateDealBasicInfo tests updating deal name and description
func TestDealUseCase_UpdateDealBasicInfo(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Original Name", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Update basic info
	updated, err := dealUC.UpdateDealBasicInfo(ctx, d.ID, "Updated Name", "Updated description")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
}

// TestDealUseCase_UpdateDealValue tests updating deal value
func TestDealUseCase_UpdateDealValue(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Update value
	updated, err := dealUC.UpdateDealValue(ctx, d.ID, 75000, "EUR")
	require.NoError(t, err)
	assert.Equal(t, int64(75000), updated.Value.Amount)
	assert.Equal(t, "EUR", updated.Value.Currency)
}

// TestDealUseCase_MoveDealToStage tests moving deal through stages
func TestDealUseCase_MoveDealToStage(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageLead, d.Stage)
	assert.Equal(t, 10, d.Probability)

	// Move to qualified
	d, err = dealUC.MoveDealToStage(ctx, d.ID, deal.DealStageQualified)
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageQualified, d.Stage)
	assert.Equal(t, 25, d.Probability)

	// Move to proposal
	d, err = dealUC.MoveDealToStage(ctx, d.ID, deal.DealStageProposal)
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageProposal, d.Stage)
	assert.Equal(t, 50, d.Probability)

	// Move to negotiation
	d, err = dealUC.MoveDealToStage(ctx, d.ID, deal.DealStageNegotiation)
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageNegotiation, d.Stage)
	assert.Equal(t, 75, d.Probability)
}

// TestDealUseCase_MarkDealAsWon tests marking deal as won
func TestDealUseCase_MarkDealAsWon(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Mark as won
	won, err := dealUC.MarkDealAsWon(ctx, d.ID, "Customer accepted proposal")
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageClosedWon, won.Stage)
	assert.Equal(t, 100, won.Probability)
	assert.NotNil(t, won.ActualCloseDate)
	assert.Equal(t, "Customer accepted proposal", won.CloseReason)
}

// TestDealUseCase_MarkDealAsLost tests marking deal as lost
func TestDealUseCase_MarkDealAsLost(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Mark as lost
	lost, err := dealUC.MarkDealAsLost(ctx, d.ID, "Chose competitor")
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageClosedLost, lost.Stage)
	assert.Equal(t, 0, lost.Probability)
	assert.NotNil(t, lost.ActualCloseDate)
	assert.Equal(t, "Chose competitor", lost.CloseReason)
}

// TestDealUseCase_UpdateDealProbability tests updating win probability
func TestDealUseCase_UpdateDealProbability(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)
	assert.Equal(t, 10, d.Probability)

	// Update probability
	updated, err := dealUC.UpdateDealProbability(ctx, d.ID, 40)
	require.NoError(t, err)
	assert.Equal(t, 40, updated.Probability)
}

// TestDealUseCase_UpdateDealExpectedCloseDate tests updating expected close date
func TestDealUseCase_UpdateDealExpectedCloseDate(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Update expected close date
	updated, err := dealUC.UpdateDealExpectedCloseDate(ctx, d.ID, "2026-09-30")
	require.NoError(t, err)
	assert.NotNil(t, updated.ExpectedCloseDate)
}

// TestDealUseCase_AssignDealToSalesRep tests assigning deal to sales rep
func TestDealUseCase_AssignDealToSalesRep(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales reps
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep1, err := userUC.Register(ctx, "salesrep1@example.com", "Sales Rep 1", "password123")
	require.NoError(t, err)
	salesRep2, err := userUC.Register(ctx, "salesrep2@example.com", "Sales Rep 2", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep1.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)
	assert.Equal(t, salesRep1.ID, d.AssignedTo)

	// Reassign to different rep
	updated, err := dealUC.AssignDealToSalesRep(ctx, d.ID, salesRep2.ID)
	require.NoError(t, err)
	assert.Equal(t, salesRep2.ID, updated.AssignedTo)
}

// TestDealUseCase_LinkDealToCompany tests linking deal to company
func TestDealUseCase_LinkDealToCompany(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	companyRepository := companyRepo.NewCompanyRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)
	companyUC := company.NewUseCase(companyRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)
	assert.Nil(t, d.CompanyID)

	// Create company (minimal parameters - only required fields)
	// Use unique name to avoid conflicts when running multiple tests
	companyName := fmt.Sprintf("Test Company %s", uuidv7.New().String()[:8])
	testCompany, err := companyUC.CreateCompany(ctx,
		companyName,        // name (unique)
		nil,                // legalName
		"llc",              // companyType (valid: llc, corporation, sole_proprietor, partnership, non_profit, other)
		nil,                // taxID
		nil,                // registrationNumber
		nil,                // website
		nil,                // email
		nil,                // phone
		nil,                // address
		nil,                // industry
		"small",            // size
		0,                  // employeeCount
		0,                  // revenue
		"USD",              // currency
		nil,                // description
		nil,                // parentCompanyID
	)
	require.NoError(t, err)

	// Link deal to company
	updated, err := dealUC.LinkDealToCompany(ctx, d.ID, testCompany.ID)
	require.NoError(t, err)
	assert.NotNil(t, updated.CompanyID)
	assert.Equal(t, testCompany.ID, *updated.CompanyID)
}

// TestDealUseCase_SetDealSource tests setting deal source
func TestDealUseCase_SetDealSource(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Set source
	updated, err := dealUC.SetDealSource(ctx, d.ID, deal.DealSourceInbound)
	require.NoError(t, err)
	assert.Equal(t, deal.DealSourceInbound, updated.Source)
}

// TestDealUseCase_DeleteDeal tests soft deleting a deal
func TestDealUseCase_DeleteDeal(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and deal
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	d, err := dealUC.CreateDeal(ctx, "Test Deal", testCustomer.ID, salesRep.ID, 50000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Delete deal
	err = dealUC.DeleteDeal(ctx, d.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = dealUC.GetDeal(ctx, d.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deal not found")
}

// TestDealUseCase_ListDeals tests listing deals with pagination
func TestDealUseCase_ListDeals(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create multiple deals
	for i := 0; i < 5; i++ {
		_, err := dealUC.CreateDeal(ctx, "Deal "+string(rune(i+65)), testCustomer.ID, salesRep.ID, int64(10000*(i+1)), "USD", "2026-12-31")
		require.NoError(t, err)
	}

	// List deals
	deals, total, err := dealUC.ListDeals(ctx, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(deals), 5)
	assert.GreaterOrEqual(t, total, int64(5))
}

// TestDealUseCase_ListDealsByStage tests listing deals by stage
func TestDealUseCase_ListDealsByStage(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deals in different stages
	_, err = dealUC.CreateDeal(ctx, "Deal Lead", testCustomer.ID, salesRep.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)

	d2, err := dealUC.CreateDeal(ctx, "Deal Qualified", testCustomer.ID, salesRep.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.MoveDealToStage(ctx, d2.ID, deal.DealStageQualified)
	require.NoError(t, err)

	d3, err := dealUC.CreateDeal(ctx, "Deal Proposal", testCustomer.ID, salesRep.ID, 30000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.MoveDealToStage(ctx, d3.ID, deal.DealStageQualified)
	require.NoError(t, err)
	_, err = dealUC.MoveDealToStage(ctx, d3.ID, deal.DealStageProposal)
	require.NoError(t, err)

	// List deals by stage
	leadDeals, total, err := dealUC.ListDealsByStage(ctx, deal.DealStageLead, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(leadDeals), 1)
	assert.GreaterOrEqual(t, total, int64(1))

	qualifiedDeals, total, err := dealUC.ListDealsByStage(ctx, deal.DealStageQualified, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(qualifiedDeals), 1)
	assert.GreaterOrEqual(t, total, int64(1))

	proposalDeals, total, err := dealUC.ListDealsByStage(ctx, deal.DealStageProposal, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(proposalDeals), 1)
	assert.GreaterOrEqual(t, total, int64(1))
}

// TestDealUseCase_ListDealsByCustomer tests listing deals by customer
func TestDealUseCase_ListDealsByCustomer(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customers and sales rep
	customer1, err := customerUC.CreateCustomer(ctx, "Customer 1", "customer1@example.com", "website", uuidv7.New())
	require.NoError(t, err)
	customer2, err := customerUC.CreateCustomer(ctx, "Customer 2", "customer2@example.com", "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deals for each customer
	_, err = dealUC.CreateDeal(ctx, "Deal C1-1", customer1.ID, salesRep.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.CreateDeal(ctx, "Deal C1-2", customer1.ID, salesRep.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.CreateDeal(ctx, "Deal C2-1", customer2.ID, salesRep.ID, 30000, "USD", "2026-12-31")
	require.NoError(t, err)

	// List deals by customer
	customer1Deals, total, err := dealUC.ListDealsByCustomer(ctx, customer1.ID, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(customer1Deals), 2)
	assert.GreaterOrEqual(t, total, int64(2))

	customer2Deals, total, err := dealUC.ListDealsByCustomer(ctx, customer2.ID, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(customer2Deals), 1)
	assert.GreaterOrEqual(t, total, int64(1))
}

// TestDealUseCase_ListDealsByAssignedTo tests listing deals by sales rep
func TestDealUseCase_ListDealsByAssignedTo(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales reps
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep1, err := userUC.Register(ctx, "salesrep1@example.com", "Sales Rep 1", "password123")
	require.NoError(t, err)
	salesRep2, err := userUC.Register(ctx, "salesrep2@example.com", "Sales Rep 2", "password123")
	require.NoError(t, err)

	// Create deals for each sales rep
	_, err = dealUC.CreateDeal(ctx, "Deal SR1-1", testCustomer.ID, salesRep1.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.CreateDeal(ctx, "Deal SR1-2", testCustomer.ID, salesRep1.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.CreateDeal(ctx, "Deal SR2-1", testCustomer.ID, salesRep2.ID, 30000, "USD", "2026-12-31")
	require.NoError(t, err)

	// List deals by sales rep
	rep1Deals, total, err := dealUC.ListDealsByAssignedTo(ctx, salesRep1.ID, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rep1Deals), 2)
	assert.GreaterOrEqual(t, total, int64(2))

	rep2Deals, total, err := dealUC.ListDealsByAssignedTo(ctx, salesRep2.ID, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rep2Deals), 1)
	assert.GreaterOrEqual(t, total, int64(1))
}

// TestDealUseCase_ListDealsBySource tests listing deals by source
func TestDealUseCase_ListDealsBySource(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deals with different sources
	d1, err := dealUC.CreateDeal(ctx, "Deal Website", testCustomer.ID, salesRep.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.SetDealSource(ctx, d1.ID, deal.DealSourceInbound)
	require.NoError(t, err)

	d2, err := dealUC.CreateDeal(ctx, "Deal Referral", testCustomer.ID, salesRep.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.SetDealSource(ctx, d2.ID, deal.DealSourceReferral)
	require.NoError(t, err)

	// List deals by source
	websiteDeals, total, err := dealUC.ListDealsBySource(ctx, deal.DealSourceInbound, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(websiteDeals), 1)
	assert.GreaterOrEqual(t, total, int64(1))

	referralDeals, total, err := dealUC.ListDealsBySource(ctx, deal.DealSourceReferral, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(referralDeals), 1)
	assert.GreaterOrEqual(t, total, int64(1))
}

// TestDealUseCase_GetPipelineStats tests getting deal counts by stage
func TestDealUseCase_GetPipelineStats(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deals in different stages
	d1, err := dealUC.CreateDeal(ctx, "Deal Lead 1", testCustomer.ID, salesRep.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.CreateDeal(ctx, "Deal Lead 2", testCustomer.ID, salesRep.ID, 15000, "USD", "2026-12-31")
	require.NoError(t, err)

	d2, err := dealUC.CreateDeal(ctx, "Deal Qualified", testCustomer.ID, salesRep.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.MoveDealToStage(ctx, d2.ID, deal.DealStageQualified)
	require.NoError(t, err)

	_, err = dealUC.MarkDealAsWon(ctx, d1.ID, "Won")
	require.NoError(t, err)

	// Get pipeline stats
	stats, err := dealUC.GetPipelineStats(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats[deal.DealStageLead], int64(1))
	assert.GreaterOrEqual(t, stats[deal.DealStageQualified], int64(1))
	assert.GreaterOrEqual(t, stats[deal.DealStageClosedWon], int64(1))
}

// TestDealUseCase_GetTotalValue tests getting total value of active deals
func TestDealUseCase_GetTotalValue(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deals
	_, err = dealUC.CreateDeal(ctx, "Deal 1", testCustomer.ID, salesRep.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.CreateDeal(ctx, "Deal 2", testCustomer.ID, salesRep.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Get total value
	totalValue, err := dealUC.GetTotalValue(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, totalValue, int64(30000))
}

// TestDealUseCase_GetWonDeals tests getting won deals count and value
func TestDealUseCase_GetWonDeals(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()[:8]), "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "salesrep@example.com", "Sales Rep", "password123")
	require.NoError(t, err)

	// Create deals and mark some as won
	d1, err := dealUC.CreateDeal(ctx, "Deal Won 1", testCustomer.ID, salesRep.ID, 10000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.MarkDealAsWon(ctx, d1.ID, "Won")
	require.NoError(t, err)

	d2, err := dealUC.CreateDeal(ctx, "Deal Won 2", testCustomer.ID, salesRep.ID, 20000, "USD", "2026-12-31")
	require.NoError(t, err)
	_, err = dealUC.MarkDealAsWon(ctx, d2.ID, "Won")
	require.NoError(t, err)

	_, err = dealUC.CreateDeal(ctx, "Deal Lead", testCustomer.ID, salesRep.ID, 30000, "USD", "2026-12-31")
	require.NoError(t, err)

	// Get won deals
	count, value, err := dealUC.GetWonDeals(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(2))
	assert.GreaterOrEqual(t, value, int64(30000))
}

// TestDealUseCase_CompleteWorkflow tests complete deal lifecycle
func TestDealUseCase_CompleteWorkflow(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	dealRepository := dealRepo.NewDealRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)
	dealUC := deal.NewUseCase(dealRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// 1. Create customer and sales rep
	testCustomer, err := customerUC.CreateCustomer(ctx, "Acme Corp", "contact@acme.com", "website", uuidv7.New())
	require.NoError(t, err)
	salesRep, err := userUC.Register(ctx, "john@company.com", "John Sales", "password123")
	require.NoError(t, err)

	// 2. Create deal
	d, err := dealUC.CreateDeal(ctx, "Q1 Enterprise Deal", testCustomer.ID, salesRep.ID, 500000, "USD", "2026-03-31")
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageLead, d.Stage)
	assert.Equal(t, 10, d.Probability)

	// 3. Update basic info
	d, err = dealUC.UpdateDealBasicInfo(ctx, d.ID, "Q1 Enterprise Deal - Updated", "Major software purchase")
	require.NoError(t, err)

	// 4. Set source
	d, err = dealUC.SetDealSource(ctx, d.ID, deal.DealSourceInbound)
	require.NoError(t, err)

	// 5. Move through pipeline
	d, err = dealUC.MoveDealToStage(ctx, d.ID, deal.DealStageQualified)
	require.NoError(t, err)
	assert.Equal(t, 25, d.Probability)

	d, err = dealUC.MoveDealToStage(ctx, d.ID, deal.DealStageProposal)
	require.NoError(t, err)
	assert.Equal(t, 50, d.Probability)

	// 6. Update value after negotiation
	d, err = dealUC.UpdateDealValue(ctx, d.ID, 450000, "USD")
	require.NoError(t, err)

	// 7. Move to negotiation
	d, err = dealUC.MoveDealToStage(ctx, d.ID, deal.DealStageNegotiation)
	require.NoError(t, err)
	assert.Equal(t, 75, d.Probability)

	// 8. Update probability
	d, err = dealUC.UpdateDealProbability(ctx, d.ID, 90)
	require.NoError(t, err)

	// 9. Mark as won
	d, err = dealUC.MarkDealAsWon(ctx, d.ID, "Contract signed")
	require.NoError(t, err)
	assert.Equal(t, deal.DealStageClosedWon, d.Stage)
	assert.Equal(t, 100, d.Probability)
	assert.NotNil(t, d.ActualCloseDate)

	// 10. Verify stats
	stats, err := dealUC.GetPipelineStats(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats[deal.DealStageClosedWon], int64(1))

	count, value, err := dealUC.GetWonDeals(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
	assert.GreaterOrEqual(t, value, int64(450000))
}
