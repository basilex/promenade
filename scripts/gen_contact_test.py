#!/usr/bin/env python3

contact_test = '''package contact_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
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

		// Create email contact
		c, err := contact.NewEmailContact(userID, "test@example.com", "Work")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, c))
		assert.NotEqual(t, uuidv7.UUID{}, c.ID)

		// GetByUserID
		contacts, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, "test@example.com", contacts[0].Email.Value())

		// GetByUserIDAndType
		emailContacts, err := repo.GetByUserIDAndType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.Len(t, emailContacts, 1)

		// Update
		c.UpdateLabel("Personal")
		c.SetVerified()
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

		// Create two contacts
		c1, _ := contact.NewEmailContact(userID, "first@example.com", "Work")
		c2, _ := contact.NewEmailContact(userID, "second@example.com", "Personal")
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
	email := valueobject.MustNewEmail("tx@example.com")

	// Transaction should rollback on error
	err := testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewContactRepository(testDB.DB)
		
		c, _ := contact.NewEmailContact(userID, email.Value(), "Work")
		require.NoError(t, repo.Create(ctx, c))
		
		// Force error to trigger rollback
		return assert.AnError
	})
	assert.Error(t, err)

	// Verify rollback - contact should not exist
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewContactRepository(testDB.DB)
		contacts, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, contacts, "Contact should have been rolled back")
	})
}
'''

with open('test/integration/contexts/identity/contact/repository_test.go', 'w') as f:
    f.write(contact_test)

print("✓ Generated test/integration/contexts/identity/contact/repository_test.go (122 lines)")
