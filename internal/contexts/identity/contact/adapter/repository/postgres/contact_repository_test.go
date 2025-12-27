package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func setupTestDB(t *testing.T) contact.IRepository {
	t.Helper()

	// Setup test database
	db := integration.SetupTestDB(t)

	// Create identity_contacts table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS identity_contacts (
			id UUID PRIMARY KEY DEFAULT uuid_v7(),
			user_id UUID NOT NULL,
			contact_type VARCHAR(20) NOT NULL CHECK (contact_type IN ('email', 'phone', 'address')),
			label VARCHAR(100) NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			address_street VARCHAR(255),
			address_street2 VARCHAR(255),
			address_city VARCHAR(100),
			address_state VARCHAR(100),
			address_postal_code VARCHAR(20),
			address_country CHAR(2),
			is_primary BOOLEAN NOT NULL DEFAULT FALSE,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			is_public BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Create partial unique index for primary contacts
	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_contacts_user_type_primary 
		ON identity_contacts(user_id, contact_type) 
		WHERE is_primary = true
	`)
	require.NoError(t, err)

	return postgres.NewContactRepository(db)
}

func TestContactRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("create email contact", func(t *testing.T) {
		c, err := contact.NewEmailContact(userID, "test@example.com", "Work")
		require.NoError(t, err)

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		assert.Equal(t, c.UserID, retrieved.UserID)
		assert.Equal(t, contact.ContactTypeEmail, retrieved.Type)
		assert.Equal(t, "Work", retrieved.Label)
		assert.NotNil(t, retrieved.Email)
		assert.Equal(t, "test@example.com", retrieved.Email.Value())
	})

	t.Run("create phone contact", func(t *testing.T) {
		c, err := contact.NewPhoneContact(userID, "+380991234567", "Personal")
		require.NoError(t, err)

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.ContactTypePhone, retrieved.Type)
		assert.NotNil(t, retrieved.Phone)
		assert.Equal(t, "+380991234567", retrieved.Phone.Value())
	})

	t.Run("create address contact", func(t *testing.T) {
		c, err := contact.NewAddressContact(userID, "123 Main St", "Kyiv", "UA", "01001", "Home")
		require.NoError(t, err)

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.ContactTypeAddress, retrieved.Type)
		assert.NotNil(t, retrieved.Address)
		assert.Equal(t, "123 Main St", retrieved.Address.Street)
		assert.Equal(t, "Kyiv", retrieved.Address.City)
		assert.Equal(t, "UA", retrieved.Address.Country)
	})
}

func TestContactRepository_GetByUserID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	// Create multiple contacts for user
	email, _ := contact.NewEmailContact(userID, "work@example.com", "Work")
	phone, _ := contact.NewPhoneContact(userID, "+380991234567", "Personal")
	address, _ := contact.NewAddressContact(userID, "Main St", "Kyiv", "UA", "01001", "Home")

	require.NoError(t, repo.Create(ctx, email))
	require.NoError(t, repo.Create(ctx, phone))
	require.NoError(t, repo.Create(ctx, address))

	// Retrieve all contacts
	contacts, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, contacts, 3)
}

func TestContactRepository_GetByUserIDAndType(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	// Create contacts of different types
	email1, _ := contact.NewEmailContact(userID, "work@example.com", "Work")
	email2, _ := contact.NewEmailContact(userID, "personal@example.com", "Personal")
	phone, _ := contact.NewPhoneContact(userID, "+380991234567", "Phone")

	require.NoError(t, repo.Create(ctx, email1))
	require.NoError(t, repo.Create(ctx, email2))
	require.NoError(t, repo.Create(ctx, phone))

	// Get only email contacts
	emails, err := repo.GetByUserIDAndType(ctx, userID, contact.ContactTypeEmail)
	require.NoError(t, err)
	assert.Len(t, emails, 2)
	for _, c := range emails {
		assert.Equal(t, contact.ContactTypeEmail, c.Type)
	}
}

func TestContactRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("update email contact", func(t *testing.T) {
		c, err := contact.NewEmailContact(userID, "old@example.com", "Work")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, c))

		// Update email
		require.NoError(t, c.SetEmail("new@example.com"))
		require.NoError(t, repo.Update(ctx, c))

		// Verify update
		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "new@example.com", retrieved.Email.Value())
	})

	t.Run("update label", func(t *testing.T) {
		c, err := contact.NewPhoneContact(userID, "+380991234567", "Work")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, c))

		// Update label
		c.UpdateLabel("Personal")
		require.NoError(t, repo.Update(ctx, c))

		// Verify update
		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "Personal", retrieved.Label)
	})
}

func TestContactRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	c, err := contact.NewEmailContact(userID, "delete@example.com", "Temp")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, c))

	// Delete contact
	err = repo.Delete(ctx, c.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, c.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestContactRepository_SetPrimary(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	// Create two email contacts
	email1, _ := contact.NewEmailContact(userID, "first@example.com", "First")
	email2, _ := contact.NewEmailContact(userID, "second@example.com", "Second")
	require.NoError(t, repo.Create(ctx, email1))
	require.NoError(t, repo.Create(ctx, email2))

	// Set first as primary
	email1.SetAsPrimary()
	require.NoError(t, repo.Update(ctx, email1))

	// Now set second as primary (should unset first)
	err := repo.SetPrimary(ctx, email2.ID)
	require.NoError(t, err)

	// Verify only second is primary
	retrieved1, _ := repo.GetByID(ctx, email1.ID)
	retrieved2, _ := repo.GetByID(ctx, email2.ID)
	assert.False(t, retrieved1.IsPrimary)
	assert.True(t, retrieved2.IsPrimary)
}

func TestContactRepository_GetPrimaryByUserIDAndType(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("returns primary contact", func(t *testing.T) {
		email1, _ := contact.NewEmailContact(userID, "first@example.com", "First")
		email2, _ := contact.NewEmailContact(userID, "second@example.com", "Second")
		email2.SetAsPrimary()

		require.NoError(t, repo.Create(ctx, email1))
		require.NoError(t, repo.Create(ctx, email2))

		primary, err := repo.GetPrimaryByUserIDAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		require.NotNil(t, primary)
		assert.Equal(t, email2.ID, primary.ID)
		assert.True(t, primary.IsPrimary)
	})

	t.Run("returns nil when no primary exists", func(t *testing.T) {
		newUserID := uuidv7.New()
		primary, err := repo.GetPrimaryByUserIDAndType(ctx, newUserID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.Nil(t, primary)
	})
}

func TestContactRepository_ExistsPrimaryForUserAndType(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("returns false when no primary exists", func(t *testing.T) {
		exists, err := repo.ExistsPrimaryForUserAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns true when primary exists", func(t *testing.T) {
		email, _ := contact.NewEmailContact(userID, "primary@example.com", "Primary")
		email.SetAsPrimary()
		require.NoError(t, repo.Create(ctx, email))

		exists, err := repo.ExistsPrimaryForUserAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestContactRepository_WithTransaction(t *testing.T) {
	integration.SetupTestDB(t)
	ctx := context.Background()

	// This test demonstrates transaction support via context
	// The actual transaction would be managed by TransactionManager
	t.Run("transaction context propagates correctly", func(t *testing.T) {
		// In real usage:
		// tm := database.NewTransactionManager(db)
		// tm.WithTransaction(ctx, func(txCtx context.Context) error {
		//     repo.Create(txCtx, contact) // Uses transaction from context
		//     return nil
		// })

		// For this test, we just verify repo operations work with regular context
		repo := setupTestDB(t)
		userID := uuidv7.New()
		
		c, err := contact.NewEmailContact(userID, "tx@example.com", "Transaction Test")
		require.NoError(t, err)
		
		// Create within "transaction context" (regular context for this test)
		txCtx := database.SetTx(ctx, nil) // Simulates transaction context
		err = repo.Create(txCtx, c)
		// Should work even with nil transaction in context (falls back to db)
		require.NoError(t, err)
	})
}
