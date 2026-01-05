package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// Helper function to create customer for tests
func createCustomer(t *testing.T, ctx context.Context, tx *sqlx.Tx) uuidv7.UUID {
	t.Helper()
	customerID := uuidv7.New()
	assignedTo := uuidv7.New() // Mock sales rep ID
	_, err := tx.ExecContext(ctx, `
		INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
	require.NoError(t, err)
	return customerID
}

func TestSubscriptionRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create subscription with metadata
		startDate := time.Now()
		sub, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)

		// Set metadata
		metadata := map[string]string{
			"promo_code":   "SAVE20",
			"affiliate_id": "AFF123",
		}
		sub.Metadata.Set(metadata)

		// Create in database
		require.NoError(t, repo.Create(ctx, sub))
		assert.NotEqual(t, uuidv7.UUID{}, sub.ID)
		assert.NotEmpty(t, sub.SubscriptionNo)
		assert.Contains(t, sub.SubscriptionNo, "SUB-")

		// Verify retrieval
		retrieved, err := repo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.Equal(t, sub.ID, retrieved.ID)
		assert.Equal(t, sub.SubscriptionNo, retrieved.SubscriptionNo)
		assert.Equal(t, customerID, retrieved.CustomerID)
		assert.Equal(t, subscription.BillingPeriodMonthly, retrieved.BillingPeriod)
		assert.Equal(t, subscription.SubscriptionStatusTrial, retrieved.Status)
		assert.Equal(t, int64(2999), retrieved.Amount.Amount)
		assert.Equal(t, "USD", retrieved.Amount.Currency)

		// Verify metadata
		retrievedMetadata := retrieved.Metadata.Get()
		assert.Equal(t, "SAVE20", retrievedMetadata["promo_code"])
		assert.Equal(t, "AFF123", retrievedMetadata["affiliate_id"])
	})
}

func TestSubscriptionRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("existing subscription", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewSubscriptionRepository(testDB.DB)
			customerID := createCustomer(t, ctx, tx)

			// Create with trial period
			startDate := time.Now()
			trialEnd := startDate.Add(14 * 24 * time.Hour)
			sub, err := subscription.NewSubscription(
				customerID,
				"plan_basic",
				subscription.BillingPeriodMonthly,
				"USD",
				2999,
				startDate,
				14,
			)
			require.NoError(t, err)
			sub.TrialEndDate = &trialEnd
			require.NoError(t, repo.Create(ctx, sub))

			// Retrieve
			retrieved, err := repo.GetByID(ctx, sub.ID)
			require.NoError(t, err)
			assert.Equal(t, sub.ID, retrieved.ID)
			assert.NotNil(t, retrieved.TrialEndDate)
			// Compare timestamps with some tolerance (avoid timezone issues)
			assert.WithinDuration(t, trialEnd, *retrieved.TrialEndDate, 2*time.Hour)
		})
	})

	t.Run("non-existing subscription", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewSubscriptionRepository(testDB.DB)
			nonExistingID := uuidv7.New()

			_, err := repo.GetByID(ctx, nonExistingID)
			assert.ErrorIs(t, err, subscription.ErrSubscriptionNotFound)
		})
	})
}

func TestSubscriptionRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create subscription
		startDate := time.Now()
		sub, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, sub))

		// Update: activate and change metadata
		require.NoError(t, sub.Activate())
		metadata := map[string]string{
			"promo_code": "NEWCODE",
			"notes":      "Upgraded plan",
		}
		sub.Metadata.Set(metadata)

		// Update in database
		require.NoError(t, repo.Update(ctx, sub))

		// Verify update
		retrieved, err := repo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		assert.Equal(t, subscription.SubscriptionStatusActive, retrieved.Status)

		retrievedMetadata := retrieved.Metadata.Get()
		assert.Equal(t, "NEWCODE", retrievedMetadata["promo_code"])
		assert.Equal(t, "Upgraded plan", retrievedMetadata["notes"])
	})
}

func TestSubscriptionRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create subscription
		startDate := time.Now()
		sub, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, sub))

		// Delete (soft delete)
		require.NoError(t, repo.Delete(ctx, sub.ID))

		// Verify soft delete
		_, err = repo.GetByID(ctx, sub.ID)
		assert.ErrorIs(t, err, subscription.ErrSubscriptionNotFound)
	})
}

func TestSubscriptionRepository_ListSubscriptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t) //  Clean tables for each test
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create 3 subscriptions
		startDate := time.Now()
		for i := 0; i < 3; i++ {
			sub, err := subscription.NewSubscription(
				customerID,
				"plan_basic",
				subscription.BillingPeriodMonthly,
				"USD",
				2999,
				startDate,
				14,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, sub))
			time.Sleep(10 * time.Millisecond) // Ensure different creation times
		}

		// List with pagination
		subs, err := repo.ListSubscriptions(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, subs, 3)

		// Verify ordering (newest first - DESC)
		assert.True(t, subs[0].CreatedAt.After(subs[1].CreatedAt) || subs[0].CreatedAt.Equal(subs[1].CreatedAt))
		assert.True(t, subs[1].CreatedAt.After(subs[2].CreatedAt) || subs[1].CreatedAt.Equal(subs[2].CreatedAt))
	})
}

func TestSubscriptionRepository_ListByCustomer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t) //  Clean tables
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customer1ID := createCustomer(t, ctx, tx)
		customer2ID := createCustomer(t, ctx, tx)

		// Create 2 subscriptions for customer1
		startDate := time.Now()
		for i := 0; i < 2; i++ {
			sub, err := subscription.NewSubscription(
				customer1ID,
				"plan_basic",
				subscription.BillingPeriodMonthly,
				"USD",
				2999,
				startDate,
				14,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, sub))
		}

		// Create 1 subscription for customer2
		sub3, err := subscription.NewSubscription(
			customer2ID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, sub3))

		// List customer1 subscriptions
		customer1Subs, err := repo.ListByCustomer(ctx, customer1ID)
		require.NoError(t, err)
		assert.Len(t, customer1Subs, 2)
		for _, sub := range customer1Subs {
			assert.Equal(t, customer1ID, sub.CustomerID)
		}

		// List customer2 subscriptions
		customer2Subs, err := repo.ListByCustomer(ctx, customer2ID)
		require.NoError(t, err)
		assert.Len(t, customer2Subs, 1)
		assert.Equal(t, customer2ID, customer2Subs[0].CustomerID)
	})
}

func TestSubscriptionRepository_ListByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t) //  Clean tables
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create 2 trial subscriptions
		startDate := time.Now()
		for i := 0; i < 2; i++ {
			sub, err := subscription.NewSubscription(
				customerID,
				"plan_basic",
				subscription.BillingPeriodMonthly,
				"USD",
				2999,
				startDate,
				14,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, sub))
		}

		// Create 1 active subscription
		subActive, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, subActive.Activate())
		require.NoError(t, repo.Create(ctx, subActive))

		// List trial subscriptions
		trialSubs, err := repo.ListByStatus(ctx, subscription.SubscriptionStatusTrial)
		require.NoError(t, err)
		assert.Len(t, trialSubs, 2)
		for _, sub := range trialSubs {
			assert.Equal(t, subscription.SubscriptionStatusTrial, sub.Status)
		}

		// List active subscriptions
		activeSubs, err := repo.ListByStatus(ctx, subscription.SubscriptionStatusActive)
		require.NoError(t, err)
		assert.Len(t, activeSubs, 1)
		assert.Equal(t, subscription.SubscriptionStatusActive, activeSubs[0].Status)
	})
}

func TestSubscriptionRepository_CountByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t) //  Clean tables
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create 3 trial, 2 active subscriptions
		startDate := time.Now()

		for i := 0; i < 3; i++ {
			sub, err := subscription.NewSubscription(
				customerID,
				"plan_basic",
				subscription.BillingPeriodMonthly,
				"USD",
				2999,
				startDate,
				14,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, sub))
		}

		for i := 0; i < 2; i++ {
			sub, err := subscription.NewSubscription(
				customerID,
				"plan_basic",
				subscription.BillingPeriodMonthly,
				"USD",
				2999,
				startDate,
				14,
			)
			require.NoError(t, err)
			require.NoError(t, sub.Activate())
			require.NoError(t, repo.Create(ctx, sub))
		}

		// Count by status
		trialCount, err := repo.CountByStatus(ctx, subscription.SubscriptionStatusTrial)
		require.NoError(t, err)
		assert.Equal(t, int64(3), trialCount)

		activeCount, err := repo.CountByStatus(ctx, subscription.SubscriptionStatusActive)
		require.NoError(t, err)
		assert.Equal(t, int64(2), activeCount)

		pausedCount, err := repo.CountByStatus(ctx, subscription.SubscriptionStatusPaused)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pausedCount)
	})
}

func TestSubscriptionRepository_GetTotalRevenue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t) //  Clean tables
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		startDate := time.Now()

		// Monthly: $29.99 (2999 cents)
		subMonthly, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, subMonthly.Activate())
		require.NoError(t, repo.Create(ctx, subMonthly))

		// Quarterly: $89.97 (8997 cents) → MRR = 8997/3 = 2999 cents
		subQuarterly, err := subscription.NewSubscription(
			customerID,
			"plan_pro",
			subscription.BillingPeriodQuarterly,
			"USD",
			8997,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, subQuarterly.Activate())
		require.NoError(t, repo.Create(ctx, subQuarterly))

		// Yearly: $359.88 (35988 cents) → MRR = 35988/12 = 2999 cents
		subYearly, err := subscription.NewSubscription(
			customerID,
			"plan_enterprise",
			subscription.BillingPeriodYearly,
			"USD",
			35988,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, subYearly.Activate())
		require.NoError(t, repo.Create(ctx, subYearly))

		// Cancelled subscription (should not be counted)
		subCancelled, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			1999,
			startDate,
			14,
		)
		require.NoError(t, err)
		require.NoError(t, subCancelled.Activate())
		require.NoError(t, subCancelled.Cancel("Test cancellation", time.Now()))
		require.NoError(t, repo.Create(ctx, subCancelled))

		// Calculate total MRR: 2999 + 2999 + 2999 = 8997 cents
		totalMRR, err := repo.GetTotalRevenue(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(8997), totalMRR)
	})
}

func TestSubscriptionRepository_MetadataJSONSerialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewSubscriptionRepository(testDB.DB)
		customerID := createCustomer(t, ctx, tx)

		// Create subscription with complex metadata
		startDate := time.Now()
		sub, err := subscription.NewSubscription(
			customerID,
			"plan_basic",
			subscription.BillingPeriodMonthly,
			"USD",
			2999,
			startDate,
			14,
		)
		require.NoError(t, err)

		// Set metadata with special characters
		metadata := map[string]string{
			"promo_code":     "SAVE20",
			"notes":          "Customer requested: \"premium\" plan with 'special' features",
			"affiliate":      "partner@example.com",
			"custom_field_1": "Value with spaces",
			"emoji":          "",
			"json_like":      `{"key": "value"}`,
		}
		sub.Metadata.Set(metadata)

		require.NoError(t, repo.Create(ctx, sub))

		// Retrieve and verify metadata
		retrieved, err := repo.GetByID(ctx, sub.ID)
		require.NoError(t, err)

		retrievedMetadata := retrieved.Metadata.Get()
		assert.Equal(t, "SAVE20", retrievedMetadata["promo_code"])
		assert.Equal(t, "Customer requested: \"premium\" plan with 'special' features", retrievedMetadata["notes"])
		assert.Equal(t, "partner@example.com", retrievedMetadata["affiliate"])
		assert.Equal(t, "Value with spaces", retrievedMetadata["custom_field_1"])
		assert.Equal(t, "", retrievedMetadata["emoji"])
		assert.Equal(t, `{"key": "value"}`, retrievedMetadata["json_like"])

		// Update metadata
		newMetadata := map[string]string{
			"promo_code": "NEWCODE",
			"updated":    "true",
			"timestamp":  "2026-01-05T10:00:00Z",
		}
		retrieved.Metadata.Set(newMetadata)
		require.NoError(t, repo.Update(ctx, retrieved))

		// Verify update
		final, err := repo.GetByID(ctx, sub.ID)
		require.NoError(t, err)
		finalMetadata := final.Metadata.Get()
		assert.Equal(t, "NEWCODE", finalMetadata["promo_code"])
		assert.Equal(t, "true", finalMetadata["updated"])
		assert.Equal(t, "2026-01-05T10:00:00Z", finalMetadata["timestamp"])
		assert.NotContains(t, finalMetadata, "notes") // Old keys should be replaced
	})
}
