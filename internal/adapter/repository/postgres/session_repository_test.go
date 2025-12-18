package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
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

func TestSessionRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("updates session successfully", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		session := helpers.SessionFixture(user.ID)
		err = repo.Create(ctx, session)
		require.NoError(t, err)

		// Update refresh token and expiration
		session.RefreshToken = "new-hashed-token"
		session.ExpiresAt = session.ExpiresAt.Add(24 * time.Hour)

		err = repo.Update(ctx, session)
		require.NoError(t, err)

		// Verify update
		updated, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, "new-hashed-token", updated.RefreshToken)
		assert.Equal(t, session.ExpiresAt.Unix(), updated.ExpiresAt.Unix())
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		session := helpers.SessionFixture(uuidv7.New())
		err := repo.Update(ctx, session)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestSessionRepository_CountUserSessions(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("counts active sessions correctly", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create 3 active sessions
		for i := 0; i < 3; i++ {
			session := helpers.SessionFixture(user.ID)
			err = repo.Create(ctx, session)
			require.NoError(t, err)
		}

		// Create 1 expired session
		expiredSession := helpers.ExpiredSessionFixture(user.ID)
		err = repo.Create(ctx, expiredSession)
		require.NoError(t, err)

		// Count should only return active sessions
		count, err := repo.CountUserSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("returns zero for user with no sessions", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		count, err := repo.CountUserSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
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

func TestSessionRepository_GetOldestSession(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewSessionRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("returns oldest active session", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create sessions at different times
		var sessions []*entity.Session
		for i := 0; i < 3; i++ {
			session := helpers.SessionFixture(user.ID)
			session.CreatedAt = time.Now().Add(time.Duration(i) * time.Hour)
			err = repo.Create(ctx, session)
			require.NoError(t, err)
			sessions = append(sessions, session)
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		// Get oldest session
		oldest, err := repo.GetOldestSession(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, sessions[0].ID, oldest.ID)
	})

	t.Run("returns ErrNotFound when no sessions exist", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		_, err = repo.GetOldestSession(ctx, user.ID)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})

	t.Run("ignores expired sessions", func(t *testing.T) {
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create expired session (oldest)
		expiredSession := helpers.ExpiredSessionFixture(user.ID)
		expiredSession.CreatedAt = time.Now().Add(-10 * time.Hour)
		err = repo.Create(ctx, expiredSession)
		require.NoError(t, err)

		// Create active session (newer)
		activeSession := helpers.SessionFixture(user.ID)
		activeSession.CreatedAt = time.Now()
		err = repo.Create(ctx, activeSession)
		require.NoError(t, err)

		// Should return active session, not expired one
		oldest, err := repo.GetOldestSession(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, activeSession.ID, oldest.ID)
	})
}
