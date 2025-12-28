package profile_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/profile"
	"github.com/basilex/promenade/internal/contexts/identity/profile/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestUser creates a test user in the database and returns its ID
func createTestUser(t *testing.T, db *integration.TestDB, userID uuidv7.UUID) {
	t.Helper()

	uniqueEmail := "test_" + userID.String() + "@example.com"

	_, err := db.DB.Exec(`
		INSERT INTO identity_users (id, email, name, password, status)
		VALUES ($1, $2, $3, $4, 'active')
	`, userID, uniqueEmail, "Test User", "password_hash")
	require.NoError(t, err)
}

func TestProfileRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	repo := postgres.NewProfileRepository(db.DB)
	ctx := context.Background()
	userID := uuidv7.New()

	createTestUser(t, db, userID)

	t.Run("create basic profile", func(t *testing.T) {
		p, err := profile.NewProfile(userID, "John Doe")
		require.NoError(t, err)

		err = repo.Create(ctx, p)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, p.UserID, retrieved.UserID)
		assert.Equal(t, "John Doe", retrieved.DisplayName)
		assert.True(t, retrieved.IsActive)
	})

	t.Run("create profile with full data", func(t *testing.T) {
		newUserID := uuidv7.New()
		createTestUser(t, db, newUserID)

		p, err := profile.NewProfile(newUserID, "Jane Smith")
		require.NoError(t, err)

		p.UpdateBio("Software Engineer")
		p.UpdatePersonalInfo("Jane", "Smith", "")
		p.UpdateGender(profile.GenderFemale)
		p.UpdateLocalization("Europe/Kyiv", "uk", "UA")
		p.UpdateSocialLinks("https://example.com", "jane", "janesmith", "jane", "", "jane.smith")
		p.SetPublic()

		err = repo.Create(ctx, p)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "Jane Smith", retrieved.DisplayName)
		assert.Equal(t, "Software Engineer", retrieved.Bio)
		assert.Equal(t, "Jane", retrieved.FirstName)
		assert.Equal(t, "Smith", retrieved.LastName)
		assert.Equal(t, profile.GenderFemale, retrieved.Gender)
		assert.Equal(t, "Europe/Kyiv", retrieved.Timezone)
		assert.Equal(t, "uk", retrieved.Language)
		assert.Equal(t, "UA", retrieved.Country)
		assert.True(t, retrieved.IsPublic)
	})
}

func TestProfileRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDB(t)
	repo := postgres.NewProfileRepository(db.DB)
	ctx := context.Background()
	userID := uuidv7.New()

	createTestUser(t, db, userID)

	t.Run("returns profile when exists", func(t *testing.T) {
		p, err := profile.NewProfile(userID, "Test Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, "Test Profile", retrieved.DisplayName)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		nonExistentID := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
	})
}

func TestProfileRepository_GetByUserID(t *testing.T) {
	db := integration.SetupTestDB(t)
	repo := postgres.NewProfileRepository(db.DB)
	ctx := context.Background()
	userID := uuidv7.New()

	createTestUser(t, db, userID)

	t.Run("returns profile for user", func(t *testing.T) {
		p, err := profile.NewProfile(userID, "User Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		retrieved, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, userID, retrieved.UserID)
		assert.Equal(t, "User Profile", retrieved.DisplayName)
	})

	t.Run("returns error when user has no profile", func(t *testing.T) {
		nonExistentUserID := uuidv7.New()
		_, err := repo.GetByUserID(ctx, nonExistentUserID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, profile.ErrNotFound)
	})
}

func TestProfileRepository_Update(t *testing.T) {
	db := integration.SetupTestDB(t)
	repo := postgres.NewProfileRepository(db.DB)
	ctx := context.Background()

	t.Run("update display name", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Old Name")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateDisplayName("New Name")
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", retrieved.DisplayName)
	})

	t.Run("update bio", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Profile With Bio")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateBio("New bio text")
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "New bio text", retrieved.Bio)
	})

	t.Run("update avatar", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Profile With Avatar")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateAvatar("https://example.com/avatar.jpg")
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/avatar.jpg", retrieved.AvatarURL)
	})

	t.Run("update personal info", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Full Name Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdatePersonalInfo("John", "Doe", "Michael")
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "John", retrieved.FirstName)
		assert.Equal(t, "Doe", retrieved.LastName)
		assert.Equal(t, "Michael", retrieved.MiddleName)
	})

	t.Run("update gender", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Gender Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateGender(profile.GenderMale)
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, profile.GenderMale, retrieved.Gender)
	})

	t.Run("update date of birth", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "DOB Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		p.UpdateDateOfBirth(&dob)
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved.DateOfBirth)
		assert.Equal(t, dob.Format("2006-01-02"), retrieved.DateOfBirth.Format("2006-01-02"))
	})

	t.Run("update localization", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Localized Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateLocalization("Europe/London", "en", "GB")
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "Europe/London", retrieved.Timezone)
		assert.Equal(t, "en", retrieved.Language)
		assert.Equal(t, "GB", retrieved.Country)
	})

	t.Run("update social links", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Social Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateSocialLinks(
			"https://mywebsite.com",
			"https://linkedin.com/in/linkedin-user",
			"https://twitter.com/twitter-user",
			"https://github.com/github-user",
			"https://facebook.com/facebook-user",
			"https://instagram.com/instagram-user",
		)
		require.NoError(t, repo.Update(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "https://mywebsite.com", retrieved.Website)
		assert.Equal(t, "https://linkedin.com/in/linkedin-user", retrieved.LinkedIn)
		assert.Equal(t, "https://twitter.com/twitter-user", retrieved.Twitter)
		assert.Equal(t, "https://github.com/github-user", retrieved.GitHub)
		assert.Equal(t, "https://facebook.com/facebook-user", retrieved.Facebook)
		assert.Equal(t, "https://instagram.com/instagram-user", retrieved.Instagram)
	})

	t.Run("toggle visibility", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Visibility Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.SetPublic()
		require.NoError(t, repo.Update(ctx, p))
		retrieved, _ := repo.GetByID(ctx, p.ID)
		assert.True(t, retrieved.IsPublic)

		p.SetPrivate()
		require.NoError(t, repo.Update(ctx, p))
		retrieved, _ = repo.GetByID(ctx, p.ID)
		assert.False(t, retrieved.IsPublic)
	})

	t.Run("toggle active status", func(t *testing.T) {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, "Active Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		p.Deactivate()
		require.NoError(t, repo.Update(ctx, p))
		retrieved, _ := repo.GetByID(ctx, p.ID)
		assert.False(t, retrieved.IsActive)

		p.Activate()
		require.NoError(t, repo.Update(ctx, p))
		retrieved, _ = repo.GetByID(ctx, p.ID)
		assert.True(t, retrieved.IsActive)
	})
}

func TestProfileRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	repo := postgres.NewProfileRepository(db.DB)
	ctx := context.Background()
	userID := uuidv7.New()

	createTestUser(t, db, userID)

	t.Run("soft delete profile", func(t *testing.T) {
		p, err := profile.NewProfile(userID, "Delete Profile")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		err = repo.Delete(ctx, p.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, p.ID)
		assert.Error(t, err)
	})
}

func TestProfileRepository_ListPublicProfiles(t *testing.T) {
	db := integration.SetupTestDB(t)
	repo := postgres.NewProfileRepository(db.DB)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		userID := uuidv7.New()
		createTestUser(t, db, userID)

		p, err := profile.NewProfile(userID, fmt.Sprintf("Public User %d", i))
		require.NoError(t, err)

		if i%2 == 0 {
			p.SetPublic()
		}

		require.NoError(t, repo.Create(ctx, p))
	}

	t.Run("returns only public profiles", func(t *testing.T) {
		profiles, err := repo.ListPublicProfiles(ctx, 10, 0)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(profiles), 3)

		for _, p := range profiles {
			assert.True(t, p.IsPublic, "Profile %s should be public", p.DisplayName)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		profiles, err := repo.ListPublicProfiles(ctx, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(profiles), 2)
	})

	t.Run("respects offset", func(t *testing.T) {
		firstBatch, err := repo.ListPublicProfiles(ctx, 2, 0)
		require.NoError(t, err)

		secondBatch, err := repo.ListPublicProfiles(ctx, 2, 2)
		require.NoError(t, err)

		if len(firstBatch) > 0 && len(secondBatch) > 0 {
			assert.NotEqual(t, firstBatch[0].ID, secondBatch[0].ID)
		}
	})
}
