package smoke_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserContact_SmokeTest - fast end-to-end smoke test for user contacts
func TestUserContact_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	userRepo := postgres.NewUserRepository(testDB.DB)
	contactRepo := postgres.NewUserContactRepository(testDB.DB)

	// Create test user
	user := helpers.UserFixture()
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("✅ Contact_create_and_retrieve", func(t *testing.T) {
		// Create email contact
		label := "Work Email"
		contact := &entity.UserContact{
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "work@example.com",
			Label:        &label,
			IsPublic:     true,
			IsVerified:   false,
		}
		err := contactRepo.Create(ctx, contact)
		require.NoError(t, err)
		assert.NotEmpty(t, contact.ID)

		// Get by ID
		retrieved, err := contactRepo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.ID, retrieved.ID)
		assert.Equal(t, entity.ContactTypeEmail, retrieved.ContactType)

		// Update
		newValue := "updated@example.com"
		contact.ContactValue = newValue
		err = contactRepo.Update(ctx, contact)
		require.NoError(t, err)

		updated, err := contactRepo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, newValue, updated.ContactValue)

		// Delete
		err = contactRepo.Delete(ctx, contact.ID)
		require.NoError(t, err)

		_, err = contactRepo.GetByID(ctx, contact.ID)
		assert.Error(t, err)
	})

	t.Log("🎉 All user contact smoke tests passed!")
}
