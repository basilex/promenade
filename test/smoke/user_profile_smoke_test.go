package smoke

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserProfile_SmokeTest - comprehensive end-to-end smoke test for user profiles
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

	// Create test users
	user1 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "profile1@test.com"
		u.Name = "Profile User 1"
	})
	require.NoError(t, userRepo.Create(ctx, user1))

	user2 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "profile2@test.com"
		u.Name = "Profile User 2"
	})
	require.NoError(t, userRepo.Create(ctx, user2))

	var profile1ID, profile2ID uuidv7.UUID

	t.Run("[+] Create_profile", func(t *testing.T) {
		bio := "Software developer"
		profile := &entity.UserProfile{
			ID:       uuidv7.New(),
			UserID:   user1.ID,
			Bio:      &bio,
			Timezone: "UTC",
			Locale:   "en",
			IsPublic: true,
		}
		err := profileRepo.Create(ctx, profile)
		require.NoError(t, err)
		assert.NotEmpty(t, profile.ID)
		assert.False(t, profile.IsBanned)
		assert.False(t, profile.IsVerified)
		profile1ID = profile.ID
	})

	t.Run("[+] Get_profile_by_ID", func(t *testing.T) {
		profile, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.Equal(t, profile1ID, profile.ID)
	})

	t.Run("[+] Get_profile_by_user_ID", func(t *testing.T) {
		profile, err := profileRepo.GetByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, user1.ID, profile.UserID)
	})

	t.Run("[+] Update_profile", func(t *testing.T) {
		profile, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)

		newBio := "Senior software developer"
		profile.Bio = &newBio
		err = profileRepo.Update(ctx, profile)
		require.NoError(t, err)

		updated, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.Equal(t, "Senior software developer", *updated.Bio)
	})

	t.Run("[+] Toggle_privacy_settings", func(t *testing.T) {
		profile, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.True(t, profile.IsPublic, "profile should be public initially")

		// Make private
		profile.IsPublic = false
		err = profileRepo.Update(ctx, profile)
		require.NoError(t, err)

		updated, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.False(t, updated.IsPublic, "profile should be private")

		// Make public again
		updated.IsPublic = true
		err = profileRepo.Update(ctx, updated)
		require.NoError(t, err)
	})

	t.Run("[+] Verify_profile", func(t *testing.T) {
		err := profileRepo.SetVerified(ctx, profile1ID, true)
		require.NoError(t, err)

		profile, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.True(t, profile.IsVerified, "profile should be verified")

		// Unverify
		err = profileRepo.SetVerified(ctx, profile1ID, false)
		require.NoError(t, err)

		profile, err = profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.False(t, profile.IsVerified, "profile should be unverified")
	})

	t.Run("[+] Ban_and_unban_profile", func(t *testing.T) {
		reason := "Violating community guidelines"
		// Ban requires: profileID, reason string, and bannedBy UUID
		// Use user2 as admin who performs the ban
		err := profileRepo.Ban(ctx, profile1ID, reason, user2.ID)
		require.NoError(t, err)

		profile, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.True(t, profile.IsBanned)
		assert.NotNil(t, profile.BanReason)
		assert.Equal(t, reason, *profile.BanReason)

		// Unban
		err = profileRepo.Unban(ctx, profile1ID)
		require.NoError(t, err)

		unbanned, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.False(t, unbanned.IsBanned)
		assert.Nil(t, unbanned.BanReason)
	})

	t.Run("[+] Update_last_seen", func(t *testing.T) {
		before := time.Now()
		err := profileRepo.UpdateLastSeen(ctx, profile1ID)
		require.NoError(t, err)
		after := time.Now()

		profile, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.NotNil(t, profile.LastSeenAt)
		assert.True(t, profile.LastSeenAt.After(before) || profile.LastSeenAt.Equal(before))
		assert.True(t, profile.LastSeenAt.Before(after) || profile.LastSeenAt.Equal(after))
	})

	t.Run("[+] Increment_profile_views", func(t *testing.T) {
		initial, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		initialViews := initial.ProfileViewsCount

		err = profileRepo.IncrementProfileViews(ctx, profile1ID)
		require.NoError(t, err)

		updated, err := profileRepo.GetByID(ctx, profile1ID)
		require.NoError(t, err)
		assert.Equal(t, initialViews+1, updated.ProfileViewsCount)
	})

	t.Run("[+] List_profiles", func(t *testing.T) {
		// Create second profile
		bio2 := "Designer"
		profile2 := &entity.UserProfile{
			ID:       uuidv7.New(),
			UserID:   user2.ID,
			Bio:      &bio2,
			Timezone: "America/New_York",
			Locale:   "en",
			IsPublic: true,
		}
		err := profileRepo.Create(ctx, profile2)
		require.NoError(t, err)
		profile2ID = profile2.ID

		// List profiles (with isPublic filter)
		profiles, err := profileRepo.List(ctx, 10, 0, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(profiles), 2, "should have at least 2 profiles")
	})

	t.Run("[+] Search_profiles", func(t *testing.T) {
		profiles, err := profileRepo.Search(ctx, "developer", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(profiles), 1, "should find profiles matching 'developer'")

		// Verify search results
		foundMatch := false
		for _, profile := range profiles {
			if profile.Bio != nil && *profile.Bio == "Senior software developer" {
				foundMatch = true
			}
		}
		assert.True(t, foundMatch, "should find profile with 'developer' in bio")
	})

	t.Run("[+] Delete_profile", func(t *testing.T) {
		err := profileRepo.Delete(ctx, profile2ID)
		require.NoError(t, err)

		_, err = profileRepo.GetByID(ctx, profile2ID)
		assert.Error(t, err, "deleted profile should not be found")
	})

	t.Logf("[SUCCESS] All user profile smoke tests passed!")
}
