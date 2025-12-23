//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/profiles/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestUserProfileRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	repo := postgres.NewUserProfileRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user first
	user := fixtures.CreateUser(t, "profile-user@example.com", "password123")

	t.Run("Create and GetByUserID", func(t *testing.T) {
		firstName := "John"
		lastName := "Doe"
		bio := "Test bio"
		gender := "male"
		profile := &entity.UserProfile{
			ID:        uuidv7.New(),
			UserID:    user.ID,
			FirstName: &firstName,
			LastName:  &lastName,
			Bio:       &bio,
			Gender:    &gender,
			Timezone:  "UTC",
			Locale:    "en-US",
			IsPublic:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, profile)
		require.NoError(t, err)

		retrieved, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, profile.UserID, retrieved.UserID)
		assert.Equal(t, firstName, *retrieved.FirstName)
		assert.Equal(t, lastName, *retrieved.LastName)
		assert.Equal(t, bio, *retrieved.Bio)
	})

	t.Run("Update", func(t *testing.T) {
		// Create a new user for this test to avoid conflicts
		updateUser := fixtures.CreateUser(t, "profile-update@example.com", "password123")
		
		firstName := "Jane"
		gender := "female"
		profile := &entity.UserProfile{
			ID:        uuidv7.New(),
			UserID:    updateUser.ID,
			FirstName: &firstName,
			Gender:    &gender,
			Timezone:  "UTC",
			Locale:    "en-US",
			IsPublic:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, profile))

		newBio := "Updated bio"
		profile.Bio = &newBio
		profile.IsPublic = true
		err := repo.Update(ctx, profile)
		require.NoError(t, err)

		retrieved, err := repo.GetByUserID(ctx, updateUser.ID)
		require.NoError(t, err)
		assert.Equal(t, newBio, *retrieved.Bio)
		assert.True(t, retrieved.IsPublic)
	})

	t.Run("GetByUserID", func(t *testing.T) {
		// Check if profile exists
		retrieved, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)

		// Non-existent user should return error
		nonExistentUserID := uuidv7.New()
		_, err = repo.GetByUserID(ctx, nonExistentUserID)
		assert.Error(t, err)
	})

	t.Run("IncrementProfileViews", func(t *testing.T) {
		// Create a new user and profile for this test
		viewUser := fixtures.CreateUser(t, "profile-views@example.com", "password123")
		
		firstName := "ViewTest"
		viewProfile := &entity.UserProfile{
			ID:        uuidv7.New(),
			UserID:    viewUser.ID,
			FirstName: &firstName,
			Timezone:  "UTC",
			Locale:    "en-US",
			IsPublic:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, viewProfile))
		
		initialViews := viewProfile.ProfileViewsCount

		err := repo.IncrementProfileViews(ctx, viewProfile.ID)
		require.NoError(t, err)

		updated, err := repo.GetByUserID(ctx, viewUser.ID)
		require.NoError(t, err)
		assert.Equal(t, initialViews+1, updated.ProfileViewsCount)
	})
}

func TestUserContactRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	repo := postgres.NewUserContactRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user
	user := fixtures.CreateUser(t, "contact-user@example.com", "password123")

	t.Run("Create and GetUserContacts", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypePhone,
			ContactValue: "+1234567890",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)

		contacts, err := repo.GetUserContacts(ctx, user.ID, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contacts), 1)

		found := false
		for _, c := range contacts {
			if c.ContactValue == "+1234567890" {
				found = true
				assert.Equal(t, entity.ContactTypePhone, c.ContactType)
			}
		}
		assert.True(t, found)
	})

	t.Run("Update", func(t *testing.T) {
		contacts, err := repo.GetUserContacts(ctx, user.ID, false)
		require.NoError(t, err)
		require.Greater(t, len(contacts), 0)

		contact := contacts[0]
		label := "Primary Phone"
		contact.Label = &label
		err = repo.Update(ctx, contact)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, label, *updated.Label)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create new user and contact for delete test
		deleteUser := fixtures.CreateUser(t, "delete-contact@example.com", "password123")
		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       deleteUser.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "delete@test.com",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, contact))

		err := repo.Delete(ctx, contact.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, contact.ID)
		assert.Error(t, err)
	})

	t.Run("GetUserContacts with inactive", func(t *testing.T) {
		contacts, err := repo.GetUserContacts(ctx, user.ID, true)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contacts), 0)

		contactsActive, err := repo.GetUserContacts(ctx, user.ID, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contactsActive), 0)
	})
}
