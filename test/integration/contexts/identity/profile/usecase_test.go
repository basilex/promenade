package profile_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/profile"
	profilePostgres "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/user"
	userPostgres "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// Helper function to create a test user (same as Contact tests)
func createTestUser(t *testing.T, ctx context.Context, userRepo user.IRepository) uuidv7.UUID {
	u, err := user.NewUser(fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "password123")
	require.NoError(t, err)
	err = userRepo.Create(ctx, u)
	require.NoError(t, err)
	return u.ID
}

// Test 1: CreateProfile
func TestProfileUseCase_CreateProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		// Create user first (FK constraint)
		userID := createTestUser(t, ctx, userRepo)

		// Test - Create profile
		p, err := uc.CreateProfile(ctx, userID, "John Doe")
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, p.GetID())
		assert.Equal(t, userID, p.UserID)
		assert.Equal(t, "John Doe", p.DisplayName)
		assert.Equal(t, profile.GenderNotSpecify, p.Gender) // Default
		assert.True(t, p.IsPublic)                          // Default
		assert.True(t, p.IsActive)                          // Default

		// Verify persistence
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "John Doe", retrieved.DisplayName)
	})
}

// Test 2: GetProfile
func TestProfileUseCase_GetProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)

		// Create profile
		created, err := uc.CreateProfile(ctx, userID, "Jane Smith")
		require.NoError(t, err)

		// Test - Get profile by ID
		retrieved, err := uc.GetProfile(ctx, created.GetID())
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, "Jane Smith", retrieved.DisplayName)

		// Test - Non-existent profile
		_, err = uc.GetProfile(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, profile.ErrProfileNotFound))
	})
}

// Test 3: GetProfileByUserID
func TestProfileUseCase_GetProfileByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)

		// Create profile
		created, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Get profile by user ID
		retrieved, err := uc.GetProfileByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, userID, retrieved.UserID)
		assert.Equal(t, "Test User", retrieved.DisplayName)

		// Test - Non-existent user
		_, err = uc.GetProfileByUserID(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, profile.ErrProfileNotFound))
	})
}

// Test 4: UpdateDisplayName
func TestProfileUseCase_UpdateDisplayName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Old Name")
		require.NoError(t, err)

		// Test - Update display name
		err = uc.UpdateDisplayName(ctx, p.GetID(), "New Name")
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "New Name", retrieved.DisplayName)
	})
}

// Test 5: UpdateBio
func TestProfileUseCase_UpdateBio(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Update bio
		bio := "Software engineer passionate about Go and DDD"
		err = uc.UpdateBio(ctx, p.GetID(), bio)
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, bio, retrieved.Bio)
	})
}

// Test 6: UpdateAvatar
func TestProfileUseCase_UpdateAvatar(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Update avatar
		avatarURL := "https://example.com/avatar.jpg"
		err = uc.UpdateAvatar(ctx, p.GetID(), avatarURL)
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, avatarURL, retrieved.AvatarURL)
	})
}

// Test 7: UpdatePersonalInfo
func TestProfileUseCase_UpdatePersonalInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Update personal info
		err = uc.UpdatePersonalInfo(ctx, p.GetID(), "John", "Doe", "Michael")
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "John", retrieved.FirstName)
		assert.Equal(t, "Doe", retrieved.LastName)
		assert.Equal(t, "Michael", retrieved.MiddleName)
	})
}

// Test 8: UpdateGender
func TestProfileUseCase_UpdateGender(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)
		assert.Equal(t, profile.GenderNotSpecify, p.Gender) // Default

		// Test - Update to Male
		err = uc.UpdateGender(ctx, p.GetID(), profile.GenderMale)
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, profile.GenderMale, retrieved.Gender)

		// Test - Update to Female
		err = uc.UpdateGender(ctx, p.GetID(), profile.GenderFemale)
		require.NoError(t, err)

		retrieved, err = uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, profile.GenderFemale, retrieved.Gender)
	})
}

// Test 9: UpdateDateOfBirth
func TestProfileUseCase_UpdateDateOfBirth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)
		assert.Nil(t, p.DateOfBirth) // Default nil

		// Test - Update date of birth
		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		err = uc.UpdateDateOfBirth(ctx, p.GetID(), &dob)
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		require.NotNil(t, retrieved.DateOfBirth)
		assert.Equal(t, dob.Year(), retrieved.DateOfBirth.Year())
		assert.Equal(t, dob.Month(), retrieved.DateOfBirth.Month())
		assert.Equal(t, dob.Day(), retrieved.DateOfBirth.Day())
	})
}

// Test 10: UpdateLocalization
func TestProfileUseCase_UpdateLocalization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Update localization
		err = uc.UpdateLocalization(ctx, p.GetID(), "Europe/Kyiv", "uk", "UA")
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Europe/Kyiv", retrieved.Timezone)
		assert.Equal(t, "uk", retrieved.Language)
		assert.Equal(t, "UA", retrieved.Country)
	})
}

// Test 11: UpdateSocialLinks
func TestProfileUseCase_UpdateSocialLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Update social links
		err = uc.UpdateSocialLinks(ctx, p.GetID(),
			"https://example.com",
			"https://linkedin.com/in/johndoe",
			"https://twitter.com/johndoe",
			"https://github.com/johndoe",
			"https://facebook.com/johndoe",
			"https://instagram.com/johndoe",
		)
		require.NoError(t, err)

		// Verify
		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", retrieved.Website)
		assert.Equal(t, "https://linkedin.com/in/johndoe", retrieved.LinkedIn)
		assert.Equal(t, "https://twitter.com/johndoe", retrieved.Twitter)
		assert.Equal(t, "https://github.com/johndoe", retrieved.GitHub)
		assert.Equal(t, "https://facebook.com/johndoe", retrieved.Facebook)
		assert.Equal(t, "https://instagram.com/johndoe", retrieved.Instagram)
	})
}

// Test 12: SetPublic and SetPrivate (combined test)
func TestProfileUseCase_SetPublicAndSetPrivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)
		assert.True(t, p.IsPublic) // Default public

		// Test - Set private
		err = uc.SetPrivate(ctx, p.GetID())
		require.NoError(t, err)

		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.False(t, retrieved.IsPublic)

		// Test - Set public again
		err = uc.SetPublic(ctx, p.GetID())
		require.NoError(t, err)

		retrieved, err = uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.True(t, retrieved.IsPublic)
	})
}

// Test 13: Activate and Deactivate (combined test)
func TestProfileUseCase_ActivateAndDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)
		assert.True(t, p.IsActive) // Default active

		// Test - Deactivate
		err = uc.Deactivate(ctx, p.GetID())
		require.NoError(t, err)

		retrieved, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.False(t, retrieved.IsActive)

		// Test - Activate again
		err = uc.Activate(ctx, p.GetID())
		require.NoError(t, err)

		retrieved, err = uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.True(t, retrieved.IsActive)
	})
}

// Test 14: DeleteProfile
func TestProfileUseCase_DeleteProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)
		p, err := uc.CreateProfile(ctx, userID, "Test User")
		require.NoError(t, err)

		// Test - Delete profile
		err = uc.DeleteProfile(ctx, p.GetID())
		require.NoError(t, err)

		// Verify - Should not be found
		_, err = uc.GetProfile(ctx, p.GetID())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, profile.ErrProfileNotFound))
	})
}

// Test 15: CompleteWorkflow - Full lifecycle test
func TestProfileUseCase_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		profileRepo := profilePostgres.NewProfileRepository(testDB.DB)
		uc := profile.NewUseCase(profileRepo)

		userID := createTestUser(t, ctx, userRepo)

		// Step 1: Create profile
		p, err := uc.CreateProfile(ctx, userID, "John Doe")
		require.NoError(t, err)
		assert.Equal(t, "John Doe", p.DisplayName)

		// Step 2: Update display name
		err = uc.UpdateDisplayName(ctx, p.GetID(), "John Michael Doe")
		require.NoError(t, err)

		// Step 3: Update bio
		err = uc.UpdateBio(ctx, p.GetID(), "Software engineer from Ukraine")
		require.NoError(t, err)

		// Step 4: Update avatar
		err = uc.UpdateAvatar(ctx, p.GetID(), "https://example.com/john.jpg")
		require.NoError(t, err)

		// Step 5: Update personal info
		err = uc.UpdatePersonalInfo(ctx, p.GetID(), "John", "Doe", "Michael")
		require.NoError(t, err)

		// Step 6: Update gender
		err = uc.UpdateGender(ctx, p.GetID(), profile.GenderMale)
		require.NoError(t, err)

		// Step 7: Update date of birth
		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		err = uc.UpdateDateOfBirth(ctx, p.GetID(), &dob)
		require.NoError(t, err)

		// Step 8: Update localization
		err = uc.UpdateLocalization(ctx, p.GetID(), "Europe/Kyiv", "uk", "UA")
		require.NoError(t, err)

		// Step 9: Update social links
		err = uc.UpdateSocialLinks(ctx, p.GetID(),
			"https://johndoe.dev",
			"https://linkedin.com/in/johndoe",
			"https://twitter.com/johndoe",
			"https://github.com/johndoe",
			"",
			"",
		)
		require.NoError(t, err)

		// Step 10: Set private
		err = uc.SetPrivate(ctx, p.GetID())
		require.NoError(t, err)

		// Step 11: Deactivate
		err = uc.Deactivate(ctx, p.GetID())
		require.NoError(t, err)

		// Step 12: Verify all updates
		final, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)

		assert.Equal(t, "John Michael Doe", final.DisplayName)
		assert.Equal(t, "Software engineer from Ukraine", final.Bio)
		assert.Equal(t, "https://example.com/john.jpg", final.AvatarURL)
		assert.Equal(t, "John", final.FirstName)
		assert.Equal(t, "Doe", final.LastName)
		assert.Equal(t, "Michael", final.MiddleName)
		assert.Equal(t, profile.GenderMale, final.Gender)
		require.NotNil(t, final.DateOfBirth)
		assert.Equal(t, 1990, final.DateOfBirth.Year())
		assert.Equal(t, "Europe/Kyiv", final.Timezone)
		assert.Equal(t, "uk", final.Language)
		assert.Equal(t, "UA", final.Country)
		assert.Equal(t, "https://johndoe.dev", final.Website)
		assert.Equal(t, "https://linkedin.com/in/johndoe", final.LinkedIn)
		assert.Equal(t, "https://twitter.com/johndoe", final.Twitter)
		assert.Equal(t, "https://github.com/johndoe", final.GitHub)
		assert.False(t, final.IsPublic)
		assert.False(t, final.IsActive)

		// Step 13: Reactivate
		err = uc.Activate(ctx, p.GetID())
		require.NoError(t, err)

		reactivated, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.True(t, reactivated.IsActive)

		// Step 14: Set public
		err = uc.SetPublic(ctx, p.GetID())
		require.NoError(t, err)

		publicProfile, err := uc.GetProfile(ctx, p.GetID())
		require.NoError(t, err)
		assert.True(t, publicProfile.IsPublic)
	})
}
