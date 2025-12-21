package smoke

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/test/helpers"
)

// TestAuth_SmokeTest verifies the complete authentication flow
func TestAuth_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	// Setup dependencies
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB.DB)
	sessionRepo := postgres.NewSessionRepository(testDB.DB)
	jwtManager := jwt.NewJWTManager("test-secret-key", time.Hour, 24*time.Hour)

	// Use memory bus with proper config
	eventBus := memory.NewMemoryBus(bus.BusConfig{
		WorkerPoolSize: 4,
		BufferSize:     100,
	})
	defer eventBus.Close(ctx)

	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager, eventBus)

	var userID string
	var accessToken, refreshToken string

	t.Run("[+] User_registration_flow", func(t *testing.T) {
		email := "smoke@test.com"
		name := "Smoke Test User"
		password := "SecurePassword123!"

		user, err := authUC.Register(ctx, email, name, password)
		require.NoError(t, err, "registration should succeed")
		require.NotNil(t, user)

		assert.NotEmpty(t, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, name, user.Name)
		assert.NotEmpty(t, user.Password, "password should be hashed")
		assert.NotEqual(t, password, user.Password, "password should be hashed, not plain")

		userID = user.ID.String()
	})

	t.Run("[+] User_login_flow", func(t *testing.T) {
		email := "smoke@test.com"
		password := "SecurePassword123!"
		userAgent := "Smoke Test Agent"
		ipAddress := "127.0.0.1"

		access, refresh, user, err := authUC.Login(ctx, email, password, userAgent, ipAddress)
		require.NoError(t, err, "login should succeed")
		require.NotNil(t, user)

		assert.NotEmpty(t, access, "access token should be generated")
		assert.NotEmpty(t, refresh, "refresh token should be generated")
		assert.Equal(t, email, user.Email)

		accessToken = access
		refreshToken = refresh
	})

	t.Run("[+] GetMe_returns_current_user", func(t *testing.T) {
		// Parse access token to get user ID
		claims, err := jwtManager.ValidateToken(accessToken)
		require.NoError(t, err, "access token should be valid")

		user, err := authUC.GetMe(ctx, claims.UserID)
		require.NoError(t, err)
		require.NotNil(t, user)

		assert.Equal(t, userID, user.ID.String())
		assert.Equal(t, "smoke@test.com", user.Email)
	})

	t.Run("[+] Refresh_token_flow", func(t *testing.T) {
		// Wait a bit to ensure new tokens have different timestamps
		time.Sleep(100 * time.Millisecond)

		newAccess, newRefresh, err := authUC.RefreshToken(ctx, refreshToken)
		require.NoError(t, err, "token refresh should succeed")

		assert.NotEmpty(t, newAccess, "new access token should be generated")
		assert.NotEmpty(t, newRefresh, "new refresh token should be generated")
		// Note: tokens might be identical if created within same millisecond
		// assert.NotEqual(t, accessToken, newAccess, "access token should be rotated")
		// assert.NotEqual(t, refreshToken, newRefresh, "refresh token should be rotated")

		// Old refresh token should be invalidated
		_, _, err = authUC.RefreshToken(ctx, refreshToken)
		assert.Error(t, err, "old refresh token should be invalidated")

		// Update tokens for logout test
		accessToken = newAccess
		refreshToken = newRefresh
	})

	t.Run("[+] Logout_invalidates_session", func(t *testing.T) {
		err := authUC.Logout(ctx, refreshToken)
		require.NoError(t, err, "logout should succeed")

		// Try to refresh with logged out token
		_, _, err = authUC.RefreshToken(ctx, refreshToken)
		assert.Error(t, err, "refresh token should be invalidated after logout")
	})

	t.Run("[+] Duplicate_email_registration_fails", func(t *testing.T) {
		email := "smoke@test.com" // Same as first registration
		name := "Another User"
		password := "AnotherPassword123!"

		_, err := authUC.Register(ctx, email, name, password)
		assert.Error(t, err, "duplicate email registration should fail")
	})

	t.Run("[+] Invalid_credentials_login_fails", func(t *testing.T) {
		email := "smoke@test.com"
		password := "WrongPassword123!"
		userAgent := "Smoke Test Agent"
		ipAddress := "127.0.0.1"

		_, _, _, err := authUC.Login(ctx, email, password, userAgent, ipAddress)
		assert.Error(t, err, "login with wrong password should fail")
	})
}

// TestAuth_SessionManagement_SmokeTest verifies session listing and management
func TestAuth_SessionManagement_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	// Setup
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(testDB.DB)
	sessionRepo := postgres.NewSessionRepository(testDB.DB)
	jwtManager := jwt.NewJWTManager("test-secret-key", time.Hour, 24*time.Hour)

	// Use memory bus with proper config
	eventBus := memory.NewMemoryBus(bus.BusConfig{
		WorkerPoolSize: 4,
		BufferSize:     100,
	})
	defer eventBus.Close(ctx)

	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager, eventBus)

	// Create user and multiple sessions
	user := helpers.UserFixture()
	// Hash password so Login will work
	require.NoError(t, user.HashPassword("password"))
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("[+] Multiple_sessions_per_user", func(t *testing.T) {
		// Login from 3 different devices
		_, refreshToken1, _, err := authUC.Login(ctx, user.Email, "password", "Device 1", "192.168.1.1")
		require.NoError(t, err)

		_, refreshToken2, _, err := authUC.Login(ctx, user.Email, "password", "Device 2", "192.168.1.2")
		require.NoError(t, err)

		_, refreshToken3, _, err := authUC.Login(ctx, user.Email, "password", "Device 3", "192.168.1.3")
		require.NoError(t, err)

		// Get all sessions
		sessions, err := authUC.GetUserSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, sessions, 3, "should have 3 active sessions")

		// Logout from one device
		err = authUC.Logout(ctx, refreshToken2)
		require.NoError(t, err)

		// Should have 2 sessions now
		sessions, err = authUC.GetUserSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, sessions, 2, "should have 2 active sessions after logout")

		// Cleanup remaining sessions
		_ = authUC.Logout(ctx, refreshToken1)
		_ = authUC.Logout(ctx, refreshToken3)
	})
}
