package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/test/helpers"
)

func TestSessionRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("creates session successfully", func(t *testing.T) {
		// Create user first
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create session
		session := helpers.SessionFixture(user.ID)
		err = repo.Create(ctx, session)
		require.NoError(t, err)

		// Verify session was created
		retrieved, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.UserID, retrieved.UserID)
		assert.Equal(t, session.RefreshToken, retrieved.RefreshToken)
	})
}

func TestSessionRepository_GetByRefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("finds session by refresh token", func(t *testing.T) {
		testDB := helpers.SetupTestDB(t)
		defer testDB.Close()
		defer testDB.CleanupTables(t)

		repo := postgres.NewSessionRepository(testDB.DB)
		userRepo := postgres.NewUserRepository(testDB.DB)

		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		session := helpers.SessionFixture(user.ID)
		err = repo.Create(ctx, session)
		require.NoError(t, err)

		retrieved, err := repo.GetByRefreshToken(ctx, session.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
	})

	t.Run("does not find expired session", func(t *testing.T) {
		testDB := helpers.SetupTestDB(t)
		defer testDB.Close()
		defer testDB.CleanupTables(t)

		repo := postgres.NewSessionRepository(testDB.DB)
		userRepo := postgres.NewUserRepository(testDB.DB)

		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		session := helpers.ExpiredSessionFixture(user.ID)
		err = repo.Create(ctx, session)
		require.NoError(t, err)

		_, err = repo.GetByRefreshToken(ctx, session.RefreshToken)
		assert.ErrorIs(t, err, entity.ErrNotFound, "Should not find expired session")
	})
}

func TestSessionRepository_GetUserSessions(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("returns all active sessions for user", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create 3 sessions
		session1 := helpers.SessionFixture(user.ID)
		session2 := helpers.SessionFixture(user.ID)
		session3 := helpers.ExpiredSessionFixture(user.ID)

		err = repo.Create(ctx, session1)
		require.NoError(t, err)
		err = repo.Create(ctx, session2)
		require.NoError(t, err)
		err = repo.Create(ctx, session3)
		require.NoError(t, err)

		sessions, err := repo.GetUserSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, sessions, 2, "Should only return active sessions")
	})
}

func TestSessionRepository_DeleteByUserID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("deletes all user sessions", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create 2 sessions
		session1 := helpers.SessionFixture(user.ID)
		session2 := helpers.SessionFixture(user.ID)
		err = repo.Create(ctx, session1)
		require.NoError(t, err)
		err = repo.Create(ctx, session2)
		require.NoError(t, err)

		// Delete all sessions
		err = repo.DeleteByUserID(ctx, user.ID)
		require.NoError(t, err)

		// Verify no sessions remain
		sessions, err := repo.GetUserSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, sessions, 0)
	})
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("deletes only expired sessions", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create active and expired sessions
		activeSession := helpers.SessionFixture(user.ID)
		expiredSession := helpers.ExpiredSessionFixture(user.ID)

		err = repo.Create(ctx, activeSession)
		require.NoError(t, err)
		err = repo.Create(ctx, expiredSession)
		require.NoError(t, err)

		// Delete expired
		err = repo.DeleteExpired(ctx)
		require.NoError(t, err)

		// Verify active session still exists
		_, err = repo.GetByID(ctx, activeSession.ID)
		require.NoError(t, err)

		// Verify expired session is gone
		_, err = repo.GetByID(ctx, expiredSession.ID)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}
