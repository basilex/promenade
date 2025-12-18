package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/ref"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("creates user successfully", func(t *testing.T) {
		testDB := helpers.SetupTestDB(t)
		defer testDB.Close()
		defer testDB.CleanupTables(t)

		repo := postgres.NewUserRepository(testDB.DB)
		user := helpers.UserFixture()

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Verify user was created
		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, retrieved.Email)
		assert.Equal(t, user.Name, retrieved.Name)
		assert.Equal(t, user.Status, retrieved.Status)
	})

	t.Run("fails on duplicate email", func(t *testing.T) {
		testDB := helpers.SetupTestDB(t)
		defer testDB.Close()
		defer testDB.CleanupTables(t)

		repo := postgres.NewUserRepository(testDB.DB)
		user1 := helpers.UserFixture()
		user2 := helpers.UserFixture(func(u *entity.User) {
			u.ID = uuidv7.New()   // Different ID
			u.Email = user1.Email // Same email
		})

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		err = repo.Create(ctx, user2)
		assert.Error(t, err, "Should fail on duplicate email")
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("finds user by email", func(t *testing.T) {
		user := helpers.UserFixture()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("returns not found for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("updates user status", func(t *testing.T) {
		user := helpers.UserFixture()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, user.ID, entity.UserStatusSuspended)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.UserStatusSuspended, retrieved.Status)
	})
}

func TestUserRepository_Suspend(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("suspends user with reason and expiry", func(t *testing.T) {
		user := helpers.UserFixture()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		reason := "Test violation"
		until := time.Now().Add(24 * time.Hour)

		err = repo.Suspend(ctx, user.ID, reason, ref.Time(until))
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.UserStatusSuspended, retrieved.Status)
		assert.Equal(t, reason, ref.StringValue(retrieved.SuspendedReason))
		assert.NotNil(t, retrieved.SuspendedUntil)
	})
}

func TestUserRepository_Ban(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("bans user permanently", func(t *testing.T) {
		user := helpers.UserFixture()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		reason := "Permanent violation"
		err = repo.Ban(ctx, user.ID, reason)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.UserStatusBanned, retrieved.Status)
		assert.Equal(t, reason, ref.StringValue(retrieved.SuspendedReason))
	})
}

func TestUserRepository_Reactivate(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("reactivates suspended user", func(t *testing.T) {
		user := helpers.SuspendedUserFixture()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.Reactivate(ctx, user.ID)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.UserStatusActive, retrieved.Status)
		assert.Nil(t, retrieved.SuspendedReason)
		assert.Nil(t, retrieved.SuspendedUntil)
	})
}

func TestUserRepository_VerifyEmail(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("verifies email and activates user", func(t *testing.T) {
		user := helpers.UnverifiedUserFixture()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.VerifyEmail(ctx, user.ID)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.UserStatusActive, retrieved.Status)
		assert.NotNil(t, retrieved.EmailVerifiedAt)
	})
}
