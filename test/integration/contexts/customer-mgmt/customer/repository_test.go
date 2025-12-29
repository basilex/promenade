package customer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestCustomerRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCustomerRepository(testDB.DB)

		// Create
		assignedTo := uuidv7.New()
		c, err := customer.NewCustomer("Test Customer", "test@example.com", "website", assignedTo)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, c))
		assert.NotEqual(t, uuidv7.UUID{}, c.ID)

		// GetByID
		found, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "Test Customer", found.Name)
		assert.Equal(t, "test@example.com", found.Email.Value())

		// GetByEmail
		foundByEmail, err := repo.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, c.ID, foundByEmail.ID)

		// Update
		c.Name = "Updated Customer"
		require.NoError(t, c.QualifyAsProspect())
		require.NoError(t, repo.Update(ctx, c))
		updated, _ := repo.GetByID(ctx, c.ID)
		assert.Equal(t, "Updated Customer", updated.Name)
		assert.Equal(t, customer.CustomerStatusProspect, updated.Status)

		// Delete
		require.NoError(t, repo.Delete(ctx, c.ID))
		_, err = repo.GetByID(ctx, c.ID)
		assert.Error(t, err)
	})
}

func TestCustomerRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCustomerRepository(testDB.DB)
		assignedTo := uuidv7.New()

		// Create 3 customers with unique emails
		uuid := uuidv7.New().String()[:8]
		for i := range 3 {
			c, _ := customer.NewCustomer("Customer "+string(rune('A'+i)), fmt.Sprintf("cust%c_%s@test.com", rune('a'+i), uuid), "web", assignedTo)
			require.NoError(t, repo.Create(ctx, c))
		}

		// ExistsByEmail
		exists, err := repo.ExistsByEmail(ctx, "custa@test.com")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByEmail(ctx, "nonexistent@test.com")
		require.NoError(t, err)
		assert.False(t, exists)

		// List with pagination
		customers, total, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 3)
		assert.GreaterOrEqual(t, len(customers), 3)

		// GetByUserID (requires user_id assignment)
		userID := uuidv7.New()
		uuid = uuidv7.New().String()[:8]
		// Create user first
		_, err = tx.ExecContext(ctx, `INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
			userID, fmt.Sprintf("user_%s@test.com", uuid), "hash", "active")
		require.NoError(t, err)
		c4, _ := customer.NewCustomer("User Customer", fmt.Sprintf("user_%s@test.com", uuid), "web", assignedTo)
		c4.UserID = &userID
		require.NoError(t, repo.Create(ctx, c4))
		
		byUser, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, c4.ID, byUser.ID)
	})
}

func TestCustomerRepository_StatusAndTier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCustomerRepository(testDB.DB)
		assignedTo := uuidv7.New()

		// Create customers with different statuses and tiers
		uuid := uuidv7.New().String()[:8]
		c1, _ := customer.NewCustomer("Lead1", fmt.Sprintf("lead1_%s@test.com", uuid), "web", assignedTo)
		c2, _ := customer.NewCustomer("Qualified1", fmt.Sprintf("qual1_%s@test.com", uuid), "web", assignedTo)
		c2.QualifyAsProspect()
		c3, _ := customer.NewCustomer("Active1", fmt.Sprintf("active1_%s@test.com", uuid), "web", assignedTo)
		c3.ConvertToCustomer()
		c3.UpgradeTier(customer.CustomerTierPro)
		
		require.NoError(t, repo.Create(ctx, c1))
		require.NoError(t, repo.Create(ctx, c2))
		require.NoError(t, repo.Create(ctx, c3))

		// ListByStatus
		leads, total, err := repo.ListByStatus(ctx, customer.CustomerStatusLead, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.GreaterOrEqual(t, len(leads), 1)

		// CountByStatus
		leadCount, err := repo.CountByStatus(ctx, customer.CustomerStatusLead)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, leadCount, 1)

		// ListByTier
		proCustomers, total, err := repo.ListByTier(ctx, customer.CustomerTierPro, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.GreaterOrEqual(t, len(proCustomers), 1)

		// CountByTier
		proCount, err := repo.CountByTier(ctx, customer.CustomerTierPro)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, proCount, 1)
	})
}

func TestCustomerRepository_Relations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCustomerRepository(testDB.DB)

		// Create customers assigned to specific person
		assignedTo1 := uuidv7.New()
		assignedTo2 := uuidv7.New()
		
		for i := range 2 {
			c, _ := customer.NewCustomer("Assigned1", fmt.Sprintf("assigned1_%d@test.com", i), "web", assignedTo1)
			require.NoError(t, repo.Create(ctx, c))
		}
		uuid := uuidv7.New().String()[:8]
		c3, _ := customer.NewCustomer("Assigned2", fmt.Sprintf("assigned2_%s@test.com", uuid), "web", assignedTo2)
		require.NoError(t, repo.Create(ctx, c3))

		// ListByAssignedTo
		assigned, total, err := repo.ListByAssignedTo(ctx, assignedTo1, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
		assert.GreaterOrEqual(t, len(assigned), 2)

		// ListByCompanyID (B2B customers)
		companyID := uuidv7.New()
		for i := range 2 {
			c, _ := customer.NewB2BCustomer("B2B Customer", fmt.Sprintf("b2b_%d@test.com", i), "web", companyID, assignedTo1)
			require.NoError(t, repo.Create(ctx, c))
		}

		byCompany, err := repo.ListByCompanyID(ctx, companyID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(byCompany), 2)
	})
}
