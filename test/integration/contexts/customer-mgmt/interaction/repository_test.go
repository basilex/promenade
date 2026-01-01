package interaction_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestUser creates a user in identity_users table for FK constraints
func createTestUser(ctx context.Context, tx *sqlx.Tx) (uuidv7.UUID, error) {
	userID := uuidv7.New()
	_, err := tx.ExecContext(ctx, `INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
		userID, fmt.Sprintf("test_%s@example.com", userID.String()[:8]), "hash", "active")
	return userID, err
}

func TestInteractionRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInteractionRepository(testDB.DB)
		custRepo := customerRepo.NewCustomerRepository(testDB.DB)

		// Create user first (FK requirement for created_by)
		userID, err := createTestUser(ctx, tx)
		require.NoError(t, err)

		// Create customer (FK requirement)
		cust, err := customer.NewCustomer("Test Customer", "test@example.com", "website", userID)
		require.NoError(t, err)
		require.NoError(t, custRepo.Create(ctx, cust))
		customerID := cust.ID

		createdBy := userID
		startedAt := time.Now()

		inter, err := interaction.NewInteraction(
			customerID,
			nil,
			interaction.InteractionTypeCall,
			interaction.InteractionDirectionOutbound,
			"Test call",
			"Discussed requirements",
			createdBy,
			startedAt,
		)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, inter))
		assert.NotEqual(t, uuidv7.UUID{}, inter.ID)

		found, err := repo.GetByID(ctx, inter.ID)
		require.NoError(t, err)
		assert.Equal(t, "Test call", found.Subject)
		assert.Equal(t, interaction.InteractionTypeCall, found.Type)

		require.NoError(t, found.SetOutcome(interaction.InteractionOutcomeSuccessful))
		require.NoError(t, repo.Update(ctx, found))
		updated, _ := repo.GetByID(ctx, found.ID)
		assert.Equal(t, interaction.InteractionOutcomeSuccessful, *updated.Outcome)

		require.NoError(t, repo.Delete(ctx, inter.ID))
		_, err = repo.GetByID(ctx, inter.ID)
		assert.Error(t, err)
	})
}

func TestInteractionRepository_ListByCustomer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInteractionRepository(testDB.DB)
		custRepo := customerRepo.NewCustomerRepository(testDB.DB)

		// Create user first (FK requirement for created_by)
		userID, err := createTestUser(ctx, tx)
		require.NoError(t, err)

		// Create customer (FK requirement)
		cust, err := customer.NewCustomer("Test Customer", "test@example.com", "website", userID)
		require.NoError(t, err)
		require.NoError(t, custRepo.Create(ctx, cust))
		customerID := cust.ID

		createdBy := userID
		startedAt := time.Now()

		for i := range 3 {
			inter, _ := interaction.NewInteraction(
				customerID,
				nil,
				interaction.InteractionTypeCall,
				interaction.InteractionDirectionOutbound,
				fmt.Sprintf("Call %d", i+1),
				"Description",
				createdBy,
				startedAt.Add(time.Duration(i)*time.Hour),
			)
			require.NoError(t, repo.Create(ctx, inter))
		}

		interactions, total, err := repo.ListByCustomer(ctx, customerID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, interactions, 3)
		assert.Equal(t, int64(3), total)
	})
}

func TestInteractionRepository_ListByType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInteractionRepository(testDB.DB)
		custRepo := customerRepo.NewCustomerRepository(testDB.DB)

		// Create user first (FK requirement for created_by)
		userID, err := createTestUser(ctx, tx)
		require.NoError(t, err)

		// Create customer (FK requirement)
		cust, err := customer.NewCustomer("Test Customer", "test@example.com", "website", userID)
		require.NoError(t, err)
		require.NoError(t, custRepo.Create(ctx, cust))
		customerID := cust.ID

		createdBy := userID
		startedAt := time.Now()

		types := []interaction.InteractionType{
			interaction.InteractionTypeCall,
			interaction.InteractionTypeCall,
			interaction.InteractionTypeEmail,
		}

		for i, iType := range types {
			inter, _ := interaction.NewInteraction(
				customerID,
				nil,
				iType,
				interaction.InteractionDirectionOutbound,
				fmt.Sprintf("Interaction %d", i+1),
				"Description",
				createdBy,
				startedAt,
			)
			require.NoError(t, repo.Create(ctx, inter))
		}

		calls, total, err := repo.ListByType(ctx, string(interaction.InteractionTypeCall), 1, 10)
		require.NoError(t, err)
		assert.Len(t, calls, 2)
		assert.Equal(t, int64(2), total)
	})
}

func TestInteractionRepository_ListPendingFollowUps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInteractionRepository(testDB.DB)
		custRepo := customerRepo.NewCustomerRepository(testDB.DB)

		// Create user first (FK requirement for created_by)
		userID, err := createTestUser(ctx, tx)
		require.NoError(t, err)

		// Create customer (FK requirement)
		cust, err := customer.NewCustomer("Test Customer", "test@example.com", "website", userID)
		require.NoError(t, err)
		require.NoError(t, custRepo.Create(ctx, cust))
		customerID := cust.ID

		createdBy := userID
		startedAt := time.Now()

		inter, _ := interaction.NewInteraction(
			customerID,
			nil,
			interaction.InteractionTypeCall,
			interaction.InteractionDirectionOutbound,
			"Call with follow-up",
			"Description",
			createdBy,
			startedAt,
		)
		// Set follow-up date in the past (pending)
		followUpDate := time.Now().Add(-24 * time.Hour)
		require.NoError(t, inter.SetFollowUp(true, &followUpDate, "Follow up notes"))
		require.NoError(t, repo.Create(ctx, inter))

		pending, total, err := repo.ListPendingFollowUps(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, pending, 1)
		assert.Equal(t, int64(1), total)
		assert.True(t, pending[0].FollowUpRequired)
	})
}

func TestInteractionRepository_Attendees(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInteractionRepository(testDB.DB)
		custRepo := customerRepo.NewCustomerRepository(testDB.DB)

		// Create user first (FK requirement for created_by)
		userID, err := createTestUser(ctx, tx)
		require.NoError(t, err)

		// Create customer (FK requirement)
		cust, err := customer.NewCustomer("Test Customer", "test@example.com", "website", userID)
		require.NoError(t, err)
		require.NoError(t, custRepo.Create(ctx, cust))
		customerID := cust.ID

		createdBy := userID
		startedAt := time.Now()

		inter, _ := interaction.NewInteraction(
			customerID,
			nil,
			interaction.InteractionTypeMeeting,
			interaction.InteractionDirectionInbound,
			"Meeting with attendees",
			"Description",
			createdBy,
			startedAt,
		)

		attendee1 := uuidv7.New()
		attendee2 := uuidv7.New()
		inter.AddAttendee(attendee1)
		inter.AddAttendee(attendee2)

		require.NoError(t, repo.Create(ctx, inter))

		found, err := repo.GetByID(ctx, inter.ID)
		require.NoError(t, err)
		assert.Len(t, found.Attendees, 2)
		assert.Contains(t, found.Attendees, attendee1)
	})
}
