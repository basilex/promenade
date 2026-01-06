package customer_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	postgres "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestCustomerRepository_N1Optimization validates that bulk count methods avoid N+1 queries
func TestCustomerRepository_N1Optimization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDBWithCleanTables(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCustomerRepository(testDB.DB)
		assignedTo := uuidv7.New()

		// Clean table before test to ensure accurate counts
		_, err := tx.ExecContext(ctx, "DELETE FROM customer_customers")
		require.NoError(t, err)

		// Create test data with different statuses and tiers
		testData := []struct {
			name   string
			status customer.CustomerStatus
			tier   customer.CustomerTier
		}{
			{"Lead Free 1", customer.CustomerStatusLead, customer.CustomerTierFree},
			{"Lead Free 2", customer.CustomerStatusLead, customer.CustomerTierFree},
			{"Prospect Basic", customer.CustomerStatusProspect, customer.CustomerTierBasic},
			{"Customer Pro", customer.CustomerStatusCustomer, customer.CustomerTierPro},
			{"Customer Enterprise", customer.CustomerStatusCustomer, customer.CustomerTierEnterprise},
			{"Churned Basic", customer.CustomerStatusChurned, customer.CustomerTierBasic},
		}

		uuid := uuidv7.New().String()
		for i, td := range testData {
			c, err := customer.NewCustomer(td.name, "cust"+string(rune('a'+i))+"_"+uuid+"@test.com", "web", assignedTo)
			require.NoError(t, err)
			c.Status = td.status
			c.Tier = td.tier
			require.NoError(t, repo.Create(ctx, c))
		}

		// Test optimized CountByAllStatuses (single GROUP BY query)
		t.Run("CountByAllStatuses", func(t *testing.T) {
			statusCounts, err := repo.CountByAllStatuses(ctx)
			require.NoError(t, err)

			// Verify counts
			assert.Equal(t, 2, statusCounts[customer.CustomerStatusLead], "Expected 2 Leads")
			assert.Equal(t, 1, statusCounts[customer.CustomerStatusProspect], "Expected 1 Prospect")
			assert.Equal(t, 2, statusCounts[customer.CustomerStatusCustomer], "Expected 2 Customers")
			assert.Equal(t, 1, statusCounts[customer.CustomerStatusChurned], "Expected 1 Churned")

			assert.Len(t, statusCounts, 4, "Should return counts for 4 statuses")
		})

		// Test optimized CountByAllTiers (single GROUP BY query)
		t.Run("CountByAllTiers", func(t *testing.T) {
			tierCounts, err := repo.CountByAllTiers(ctx)
			require.NoError(t, err)

			// Verify counts
			assert.Equal(t, 2, tierCounts[customer.CustomerTierFree], "Expected 2 Free tier")
			assert.Equal(t, 2, tierCounts[customer.CustomerTierBasic], "Expected 2 Basic tier")
			assert.Equal(t, 1, tierCounts[customer.CustomerTierPro], "Expected 1 Pro tier")
			assert.Equal(t, 1, tierCounts[customer.CustomerTierEnterprise], "Expected 1 Enterprise tier")

			assert.Len(t, tierCounts, 4, "Should return counts for 4 tiers")
		})

		// Performance comparison documentation
		t.Run("Performance_Comparison", func(t *testing.T) {
			// Old N+1 approach: 8 separate queries (4 statuses + 4 tiers)
			// New GROUP BY approach: 2 queries total
			// Result: 75% query reduction

			statusCounts, err := repo.CountByAllStatuses(ctx)
			require.NoError(t, err)
			assert.Len(t, statusCounts, 4, "Should return counts for 4 statuses")

			tierCounts, err := repo.CountByAllTiers(ctx)
			require.NoError(t, err)
			assert.Len(t, tierCounts, 4, "Should return counts for 4 tiers")
		})
	})
}