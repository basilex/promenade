package customer_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	customerPostgres "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestCustomerUseCase_CreateCustomer tests creating a new B2C customer
func TestCustomerUseCase_CreateCustomer(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customer
		repID := uuidv7.New()
		email := fmt.Sprintf("john_%s@example.com", uuidv7.New().String())
		cust, err := uc.CreateCustomer(ctx, "John Doe", email, "website", repID)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, cust.GetID())
		assert.Equal(t, "John Doe", cust.Name)
		assert.Equal(t, email, cust.Email.Value())
		assert.Equal(t, customer.CustomerStatusLead, cust.Status)
		assert.Equal(t, customer.CustomerTierFree, cust.Tier)

		// Test - Duplicate email should fail
		_, err = uc.CreateCustomer(ctx, "Jane Doe", email, "referral", repID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, customer.ErrCustomerAlreadyExists))
	})
}

// TestCustomerUseCase_CreateB2BCustomer tests creating a B2B customer linked to company
func TestCustomerUseCase_CreateB2BCustomer(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create company first (use tx for transaction support)
		companyID := uuidv7.New()
		_, err := tx.ExecContext(ctx,
			`INSERT INTO customer_companies (id, name, legal_name, type, size, industry, country)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		companyID, "Acme Corp", "Acme Corporation", "llc", "small", "Technology", "US",
	)
	require.NoError(t, err)

	// Create B2B customer
	repID := uuidv7.New()
		cust, err := uc.CreateB2BCustomer(ctx, "Alice Smith", fmt.Sprintf("alice_%s@acme.com", uuidv7.New().String()), "partnership", companyID, repID)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, cust.GetID())
		assert.Equal(t, "Alice Smith", cust.Name)
		assert.Equal(t, companyID, *cust.CompanyID)
		assert.Equal(t, customer.CustomerStatusLead, cust.Status)
	})
}

// TestCustomerUseCase_GetCustomer tests retrieving a customer by ID
func TestCustomerUseCase_GetCustomer(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customer
		repID := uuidv7.New()
		created, err := uc.CreateCustomer(ctx, "Bob Wilson", fmt.Sprintf("bob_%s@example.com", uuidv7.New().String()), "cold_call", repID)
		require.NoError(t, err)

		// Test - Get by ID
		retrieved, err := uc.GetCustomer(ctx, created.GetID())
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, "Bob Wilson", retrieved.Name)

		// Test - Non-existent ID
		_, err = uc.GetCustomer(ctx, uuidv7.New())
		assert.Error(t, err)
	})
}

// TestCustomerUseCase_GetCustomerByEmail tests retrieving a customer by email
func TestCustomerUseCase_GetCustomerByEmail(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customer
		repID := uuidv7.New()
		email := fmt.Sprintf("carol_%s@example.com", uuidv7.New().String())
		created, err := uc.CreateCustomer(ctx, "Carol Davis", email, "webinar", repID)
		require.NoError(t, err)

		// Test - Get by email
		retrieved, err := uc.GetCustomerByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, email, retrieved.Email.Value())

		// Test - Non-existent email
		nonExistentEmail := fmt.Sprintf("nonexistent_%s@example.com", uuidv7.New().String())
		_, err = uc.GetCustomerByEmail(ctx, nonExistentEmail)
		assert.Error(t, err)
	})
}

// TestCustomerUseCase_QualifyAsProspect tests lifecycle transition Lead → Prospect
func TestCustomerUseCase_QualifyAsProspect(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customer (starts as Lead)
		repID := uuidv7.New()
		cust, err := uc.CreateCustomer(ctx, "David Lee", fmt.Sprintf("david_%s@example.com", uuidv7.New().String()), "referral", repID)
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerStatusLead, cust.Status)

		// Test - Qualify as prospect
		err = uc.QualifyAsProspect(ctx, cust.GetID())
		require.NoError(t, err)

		// Verify status changed
		updated, err := uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerStatusProspect, updated.Status)
	})
}

// TestCustomerUseCase_ConvertToCustomer tests lifecycle transition Prospect → Customer
func TestCustomerUseCase_ConvertToCustomer(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create and qualify customer
		repID := uuidv7.New()
		cust, err := uc.CreateCustomer(ctx, "Emma Brown", fmt.Sprintf("emma_%s@example.com", uuidv7.New().String()), "event", repID)
		require.NoError(t, err)
		err = uc.QualifyAsProspect(ctx, cust.GetID())
		require.NoError(t, err)

		// Test - Convert to paying customer
		err = uc.ConvertToCustomer(ctx, cust.GetID())
		require.NoError(t, err)

		// Verify status changed
		updated, err := uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerStatusCustomer, updated.Status)
	})
}

// TestCustomerUseCase_TierManagement tests tier upgrade and downgrade
func TestCustomerUseCase_TierManagement(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customer (starts as Free tier)
		repID := uuidv7.New()
		cust, err := uc.CreateCustomer(ctx, "Frank Miller", fmt.Sprintf("frank_%s@example.com", uuidv7.New().String()), "trial", repID)
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerTierFree, cust.Tier)

		// Test - Upgrade to Basic
		err = uc.UpgradeCustomerTier(ctx, cust.GetID(), customer.CustomerTierBasic)
		require.NoError(t, err)

		updated, err := uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerTierBasic, updated.Tier)

		// Test - Upgrade to Pro
		err = uc.UpgradeCustomerTier(ctx, cust.GetID(), customer.CustomerTierPro)
		require.NoError(t, err)

		updated, err = uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerTierPro, updated.Tier)

		// Test - Downgrade to Basic
		err = uc.DowngradeCustomerTier(ctx, cust.GetID(), customer.CustomerTierBasic)
		require.NoError(t, err)

		updated, err = uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerTierBasic, updated.Tier)
	})
}

// TestCustomerUseCase_TagManagement tests adding and removing tags
func TestCustomerUseCase_TagManagement(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customer
		repID := uuidv7.New()
		cust, err := uc.CreateCustomer(ctx, "Grace Wong", fmt.Sprintf("grace_%s@example.com", uuidv7.New().String()), "marketing", repID)
		require.NoError(t, err)

		// Test - Add tags
		err = uc.AddTagToCustomer(ctx, cust.GetID(), "vip")
		require.NoError(t, err)
		err = uc.AddTagToCustomer(ctx, cust.GetID(), "enterprise")
		require.NoError(t, err)

		updated, err := uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Len(t, updated.Tags, 2)
		assert.Contains(t, updated.Tags, "vip")
		assert.Contains(t, updated.Tags, "enterprise")

		// Test - Remove tag
		err = uc.RemoveTagFromCustomer(ctx, cust.GetID(), "vip")
		require.NoError(t, err)

		updated, err = uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Len(t, updated.Tags, 1)
		assert.Contains(t, updated.Tags, "enterprise")
		assert.NotContains(t, updated.Tags, "vip")
	})
}

// TestCustomerUseCase_ListCustomers tests listing with pagination
func TestCustomerUseCase_ListCustomers(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create multiple customers
		repID := uuidv7.New()
		for i := 1; i <= 5; i++ {
			_, err := uc.CreateCustomer(ctx, "Customer "+string(rune('A'+i-1)), 
				string(rune('a'+i-1))+"@example.com", "test", repID)
			require.NoError(t, err)
		}

		// Test - List all
		customers, total, err := uc.ListCustomers(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 5)
		assert.GreaterOrEqual(t, total, 5)

		// Test - Pagination
		customers, total, err = uc.ListCustomers(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(customers))
		assert.GreaterOrEqual(t, total, 5)
	})
}

// TestCustomerUseCase_ListByStatus tests listing customers by status
func TestCustomerUseCase_ListByStatus(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Create customers and qualify some
		repID := uuidv7.New()
		_, _ = uc.CreateCustomer(ctx, "Henry Kim", fmt.Sprintf("henry_%s@example.com", uuidv7.New().String()), "ad", repID)
		cust2, _ := uc.CreateCustomer(ctx, "Iris Chen", fmt.Sprintf("iris_%s@example.com", uuidv7.New().String()), "social", repID)
		_ = uc.QualifyAsProspect(ctx, cust2.GetID())

		// Test - List leads
		leads, total, err := uc.ListCustomersByStatus(ctx, customer.CustomerStatusLead, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(leads), 1)
		assert.GreaterOrEqual(t, total, 1)

		// Test - List prospects
		prospects, total, err := uc.ListCustomersByStatus(ctx, customer.CustomerStatusProspect, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(prospects), 1)
		assert.GreaterOrEqual(t, total, 1)
	})
}

// TestCustomerUseCase_CompleteWorkflow tests complete customer lifecycle
func TestCustomerUseCase_CompleteWorkflow(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := customerPostgres.NewCustomerRepository(testDB.DB)
		uc := customer.NewUseCase(repo)

		// Step 1: Create lead
		repID := uuidv7.New()
		cust, err := uc.CreateCustomer(ctx, "Jack Taylor", fmt.Sprintf("jack_%s@example.com", uuidv7.New().String()), "demo_request", repID)
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerStatusLead, cust.Status)
		assert.Equal(t, customer.CustomerTierFree, cust.Tier)

		// Step 2: Add phone number
		err = uc.SetCustomerPhone(ctx, cust.GetID(), "+1234567890")
		require.NoError(t, err)

		// Step 3: Add segmentation tags
		err = uc.AddTagToCustomer(ctx, cust.GetID(), "high-value")
		require.NoError(t, err)
		err = uc.AddTagToCustomer(ctx, cust.GetID(), "tech-startup")
		require.NoError(t, err)

		// Step 4: Qualify as prospect
		err = uc.QualifyAsProspect(ctx, cust.GetID())
		require.NoError(t, err)

		// Step 5: Convert to paying customer
		err = uc.ConvertToCustomer(ctx, cust.GetID())
		require.NoError(t, err)

		// Step 6: Upgrade to Pro tier
		err = uc.UpgradeCustomerTier(ctx, cust.GetID(), customer.CustomerTierPro)
		require.NoError(t, err)

		// Verify final state
		final, err := uc.GetCustomer(ctx, cust.GetID())
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerStatusCustomer, final.Status)
		assert.Equal(t, customer.CustomerTierPro, final.Tier)
		assert.Equal(t, "+1234567890", final.Phone.Value())
		assert.Len(t, final.Tags, 2)
		assert.Contains(t, final.Tags, "high-value")
		assert.Contains(t, final.Tags, "tech-startup")
	})
}
