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

// TestUserProfile_SmokeTest - fast end-to-end smoke test for user profiles
func TestUserProfile_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	userRepo := postgres.NewUserRepository(testDB.DB)
	profileRepo := postgres.NewUserProfileRepository(testDB.DB)

	// Create test user
	user := helpers.UserFixture()
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("✅ Profile_create_and_retrieve", func(t *testing.T) {
		// Create
		bio := "Test bio"
		profile := &entity.UserProfile{
			ID:       helpers.UserFixture().ID, // Generate ID
			UserID:   user.ID,
			Bio:      &bio,
			Timezone: "UTC",
			Locale:   "en",
			IsPublic: true,
		}
		err := profileRepo.Create(ctx, profile)
		require.NoError(t, err)
		assert.NotEmpty(t, profile.ID)

		// Get by UserID
		retrieved, err := profileRepo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.UserID)

		// Update
		newBio := "Updated bio"
		profile.Bio = &newBio
		err = profileRepo.Update(ctx, profile)
		require.NoError(t, err)

		updated, err := profileRepo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, newBio, *updated.Bio)
	})

	t.Log("🎉 All user profile smoke tests passed!")
}
