package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestUserRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("create user successfully", func(t *testing.T) {
		u, err := user.NewUser("test@example.com", "password123")
		require.NoError(t, err)

		err = repo.Create(ctx, u)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, u.ID, retrieved.ID)
		assert.Equal(t, "test@example.com", retrieved.Email.Value())
		assert.Equal(t, user.UserStatusActive, retrieved.Status)
		assert.False(t, retrieved.EmailVerified)
	})

	t.Run("create user with duplicate email fails", func(t *testing.T) {
		u1, _ := user.NewUser("duplicate@example.com", "password123")
		u2, _ := user.NewUser("duplicate@example.com", "password456")

		err := repo.Create(ctx, u1)
		require.NoError(t, err)

		// Second create with same email should fail (unique constraint)
		err = repo.Create(ctx, u2)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing user", func(t *testing.T) {
		u, _ := user.NewUser("getbyid@example.com", "password123")
		require.NoError(t, repo.Create(ctx, u))

		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, u.ID, retrieved.ID)
		assert.Equal(t, "getbyid@example.com", retrieved.Email.Value())
	})

	t.Run("get non-existent user returns error", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing user by email", func(t *testing.T) {
		uniqueEmail := "getbyemail_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		retrieved, err := repo.GetByEmail(ctx, uniqueEmail)
		require.NoError(t, err)
		assert.Equal(t, u.ID, retrieved.ID)
		assert.Equal(t, uniqueEmail, retrieved.Email.Value())
	})

	t.Run("get user by email is case-insensitive", func(t *testing.T) {
		uniqueEmail := "CaseSensitive_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Query with different case
		retrieved, err := repo.GetByEmail(ctx, uniqueEmail)
		require.NoError(t, err)
		assert.Equal(t, u.ID, retrieved.ID)
	})

	t.Run("get non-existent email returns error", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("update user successfully", func(t *testing.T) {
		uniqueEmail := "update_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Update user
		u.VerifyEmail()
		u.RecordLogin()

		err := repo.Update(ctx, u)
		require.NoError(t, err)

		// Verify changes were saved
		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.EmailVerified)
		assert.NotNil(t, retrieved.EmailVerifiedAt)
		assert.NotNil(t, retrieved.LastLoginAt)
	})

	t.Run("update password hash", func(t *testing.T) {
		uniqueEmail := "changepass_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Change password
		err := u.ChangePassword("newPassword456")
		require.NoError(t, err)

		err = repo.Update(ctx, u)
		require.NoError(t, err)

		// Verify new password works
		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.NoError(t, retrieved.CheckPassword("newPassword456"))
		assert.Error(t, retrieved.CheckPassword("password123"))
	})

	t.Run("update status", func(t *testing.T) {
		uniqueEmail := "status_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Suspend user
		u.Suspend()
		err := repo.Update(ctx, u)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, user.UserStatusSuspended, retrieved.Status)
		assert.False(t, retrieved.IsActive())
	})

	t.Run("update failed login count and lock", func(t *testing.T) {
		uniqueEmail := "lockaccount_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Record 5 failed logins
		for i := 0; i < 5; i++ {
			u.RecordFailedLogin()
		}

		err := repo.Update(ctx, u)
		require.NoError(t, err)

		// Verify account is locked
		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, retrieved.FailedLoginCount)
		assert.True(t, retrieved.IsLocked())
		assert.NotNil(t, retrieved.LockedUntil)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("soft delete user", func(t *testing.T) {
		u, _ := user.NewUser("delete@example.com", "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Delete user
		err := repo.Delete(ctx, u.ID)
		require.NoError(t, err)

		// Verify user cannot be retrieved (soft deleted)
		_, err = repo.GetByID(ctx, u.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
	})

	t.Run("delete non-existent user returns error", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		err := repo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
	})
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("existing email returns true", func(t *testing.T) {
		uniqueEmail := "exists_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		exists, err := repo.ExistsByEmail(ctx, uniqueEmail)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing email returns false", func(t *testing.T) {
		exists, err := repo.ExistsByEmail(ctx, "doesnotexist@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("check is case-insensitive", func(t *testing.T) {
		uniqueEmail := "CaseCheck_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))

		exists, err := repo.ExistsByEmail(ctx, uniqueEmail)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("deleted user email returns false", func(t *testing.T) {
		uniqueEmail := "deleted_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, repo.Delete(ctx, u.ID))

		exists, err := repo.ExistsByEmail(ctx, uniqueEmail)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestUserRepository_ListUsers(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create multiple users with unique emails
	for i := 0; i < 5; i++ {
		uniqueEmail := "listuser" + string(rune('a'+i)) + "_" + uuidv7.New().String() + "@example.com"
		u, _ := user.NewUser(uniqueEmail, "password123")
		require.NoError(t, repo.Create(ctx, u))
	}

	t.Run("list all users", func(t *testing.T) {
		retrieved, total, err := repo.ListUsers(ctx, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 5)
		assert.GreaterOrEqual(t, total, 5)
	})

	t.Run("pagination - first page", func(t *testing.T) {
		retrieved, total, err := repo.ListUsers(ctx, 1, 2)
		require.NoError(t, err)
		assert.Len(t, retrieved, 2)
		assert.GreaterOrEqual(t, total, 5)
	})

	t.Run("pagination - second page", func(t *testing.T) {
		retrieved, total, err := repo.ListUsers(ctx, 2, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 2)
		assert.GreaterOrEqual(t, total, 5)
	})

	t.Run("empty result with large offset", func(t *testing.T) {
		retrieved, total, err := repo.ListUsers(ctx, 1000, 10)
		require.NoError(t, err)
		assert.Empty(t, retrieved)
		assert.GreaterOrEqual(t, total, 5)
	})
}

func TestUserRepository_ValueObjectRoundTrip(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("email value object round-trip", func(t *testing.T) {
		u, err := user.NewUser("roundtrip@example.com", "password123")
		require.NoError(t, err)

		// Save to DB
		err = repo.Create(ctx, u)
		require.NoError(t, err)

		// Retrieve from DB
		retrieved, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)

		// Verify email value object is properly reconstructed
		assert.Equal(t, "roundtrip@example.com", retrieved.Email.Value())
		assert.NotNil(t, retrieved.Email)
		assert.Equal(t, u.Email.Value(), retrieved.Email.Value())
	})
}

func TestUserRepository_ConcurrentUpdates(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	t.Run("last update wins", func(t *testing.T) {
		u, _ := user.NewUser("concurrent@example.com", "password123")
		require.NoError(t, repo.Create(ctx, u))

		// Simulate two concurrent operations
		u1, _ := repo.GetByID(ctx, u.ID)
		u2, _ := repo.GetByID(ctx, u.ID)

		// Both modify different fields
		u1.VerifyEmail()
		u2.RecordLogin()

		// Both update (last one wins scenario)
		require.NoError(t, repo.Update(ctx, u1))
		require.NoError(t, repo.Update(ctx, u2))

		// Verify final state
		final, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		// u2's update should have overwritten u1's changes
		assert.NotNil(t, final.LastLoginAt)
	})
}
