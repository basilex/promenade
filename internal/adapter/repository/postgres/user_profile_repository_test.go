package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserProfileRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	displayName := "John Doe"
	nickname := "johndoe"
	bio := "Software developer"
	gender := string(entity.GenderMale)

	profile := &entity.UserProfile{
		ID:          uuidv7.New(),
		UserID:      user.ID,
		DisplayName: &displayName,
		Nickname:    &nickname,
		Bio:         &bio,
		Gender:      &gender,
		IsPublic:    true,
		Timezone:    "UTC",
		Locale:      "en",
		SocialLinks: entity.SocialLinks{
			"twitter": "https://twitter.com/johndoe",
		},
		Preferences: entity.Preferences{
			"theme": "dark",
		},
	}

	err := repo.Create(ctx, profile)
	require.NoError(t, err)
	assert.NotZero(t, profile.CreatedAt)

	// Verify profile was created
	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.Equal(t, profile.ID, retrieved.ID)
	assert.Equal(t, *profile.DisplayName, *retrieved.DisplayName)
}

func TestUserProfileRepository_GetByID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.Equal(t, profile.ID, retrieved.ID)

	_, err = repo.GetByID(ctx, uuidv7.New())
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserProfileRepository_GetByUserID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	retrieved, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, profile.ID, retrieved.ID)
}

func TestUserProfileRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	newBio := "Updated bio"
	profile.Bio = &newBio

	err := repo.Update(ctx, profile)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.Equal(t, newBio, *retrieved.Bio)
}

func TestUserProfileRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	err := repo.Delete(ctx, profile.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, profile.ID)
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserProfileRepository_List(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user1 := helpers.CreateTestUser(t, testDB.DB, "user1@example.com", "User 1")
	helpers.CreateTestProfile(t, testDB.DB, user1.ID, "user1")

	user2 := helpers.CreateTestUser(t, testDB.DB, "user2@example.com", "User 2")
	helpers.CreateTestProfile(t, testDB.DB, user2.ID, "user2")

	profiles, err := repo.List(ctx, 10, 0, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(profiles), 2)
}

func TestUserProfileRepository_UpdateLastSeen(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	time.Sleep(100 * time.Millisecond)

	err := repo.UpdateLastSeen(ctx, profile.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastSeenAt)
}

func TestUserProfileRepository_IncrementProfileViews(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	initialViews := profile.ProfileViewsCount

	err := repo.IncrementProfileViews(ctx, profile.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.Equal(t, initialViews+1, retrieved.ProfileViewsCount)
}

func TestUserProfileRepository_Ban(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")
	admin := helpers.CreateTestUser(t, testDB.DB, "admin@example.com", "Admin User")
	adminID := admin.ID

	reason := "Violation of terms"
	err := repo.Ban(ctx, profile.ID, reason, adminID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.IsBanned)
	assert.Equal(t, reason, *retrieved.BanReason)
}

func TestUserProfileRepository_Unban(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")
	admin := helpers.CreateTestUser(t, testDB.DB, "admin@example.com", "Admin User")
	adminID := admin.ID

	reason := "Test ban"
	err := repo.Ban(ctx, profile.ID, reason, adminID)
	require.NoError(t, err)

	err = repo.Unban(ctx, profile.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsBanned)
	assert.Nil(t, retrieved.BanReason)
}

func TestUserProfileRepository_SetVerified(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	profile := helpers.CreateTestProfile(t, testDB.DB, user.ID, "testuser")

	err := repo.SetVerified(ctx, profile.ID, true)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.IsVerified)

	err = repo.SetVerified(ctx, profile.ID, false)
	require.NoError(t, err)

	retrieved, err = repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsVerified)
}

func TestUserProfileRepository_Search(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserProfileRepository(testDB.DB)
	ctx := context.Background()

	user1 := helpers.CreateTestUser(t, testDB.DB, "john@example.com", "John")
	profile1 := helpers.CreateTestProfile(t, testDB.DB, user1.ID, "johndoe")
	displayName1 := "John Doe"
	bio1 := "Software developer"
	profile1.DisplayName = &displayName1
	profile1.Bio = &bio1
	profile1.IsPublic = true
	_ = repo.Update(ctx, profile1)

	user2 := helpers.CreateTestUser(t, testDB.DB, "jane@example.com", "Jane")
	profile2 := helpers.CreateTestProfile(t, testDB.DB, user2.ID, "janedoe")
	displayName2 := "Jane Doe"
	profile2.DisplayName = &displayName2
	profile2.IsPublic = true
	_ = repo.Update(ctx, profile2)

	results, err := repo.Search(ctx, "doe", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
}
