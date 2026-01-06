package contact_test

import (
	"context"
	"testing"
"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestContactRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewContactRepository(testDB.DB)
		userID := uuidv7.New()

		// Create user first (required for foreign key)
		_, err := tx.ExecContext(ctx, `INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
			userID, "user_"+userID.String()+"@test.com", "hash", "active")
		require.NoError(t, err)

		// Create email contact
		testEmail := fmt.Sprintf("test_%s@example.com", uuidv7.New().String())
		c, err := contact.NewEmailContact(userID, testEmail, "Work")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, c))
		assert.NotEqual(t, uuidv7.UUID{}, c.ID)

		// GetByUserID
		contacts, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, testEmail, contacts[0].Email.Value())

		// GetByUserIDAndType
		emailContacts, err := repo.GetByUserIDAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.Len(t, emailContacts, 1)

		// Update
		_ = c.UpdateLabel("Personal")
		c.Verify()
		require.NoError(t, repo.Update(ctx, c))
		updated, _ := repo.GetByUserID(ctx, userID)
		assert.Equal(t, "Personal", updated[0].Label)
		assert.True(t, updated[0].IsVerified)

		// Delete
		require.NoError(t, repo.Delete(ctx, c.ID))
		deleted, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, deleted)
	})
}

func TestContactRepository_Primary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewContactRepository(testDB.DB)
		userID := uuidv7.New()

		// Create user first (required for foreign key)
		_, err := tx.ExecContext(ctx, `INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
			userID, "user_"+userID.String()+"@test.com", "hash", "active")
		require.NoError(t, err)

		// Create two contacts
		c1, _ := contact.NewEmailContact(userID, fmt.Sprintf("first_%s@example.com", uuidv7.New().String()), "Work")
		c2, _ := contact.NewEmailContact(userID, fmt.Sprintf("second_%s@example.com", uuidv7.New().String()), "Personal")
		require.NoError(t, repo.Create(ctx, c1))
		require.NoError(t, repo.Create(ctx, c2))

		// ExistsPrimaryForUserAndType (should be false initially)
		exists, err := repo.ExistsPrimaryForUserAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.False(t, exists)

		// SetPrimary
		c1.SetAsPrimary()
		require.NoError(t, repo.Update(ctx, c1))

		// ExistsPrimaryForUserAndType (should be true now)
		exists, err = repo.ExistsPrimaryForUserAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.True(t, exists)

		// GetPrimaryByUserIDAndType
		primary, err := repo.GetPrimaryByUserIDAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.Equal(t, c1.ID, primary.ID)
		assert.True(t, primary.IsPrimary)
	})
}

func TestContactRepository_WithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	userID := uuidv7.New()
	uuid := uuidv7.New().String()
	email := "tx_" + uuid + "@example.com"

	// Create user outside transaction for FK constraint
	_, err := testDB.DB.Exec(`INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
		userID, "user_"+uuid+"@test.com", "hash", "active")
	require.NoError(t, err)

	// Test transaction rollback
	tx, err := testDB.DB.BeginTxx(context.Background(), nil)
	require.NoError(t, err)
	
	ctx := database.SetTxToContext(context.Background(), tx)
	repo := postgres.NewContactRepository(testDB.DB)
	
	c, _ := contact.NewEmailContact(userID, email, "Work")
	err = repo.Create(ctx, c)
	require.NoError(t, err)
	
	// Verify contact was created in transaction
	found, err := repo.GetByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.ID, found.ID)
	
	// Rollback
	_ = tx.Rollback()

	// Verify rollback - contact should not exist
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewContactRepository(testDB.DB)
		contacts, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, contacts, "Contact should have been rolled back")
	})
}
