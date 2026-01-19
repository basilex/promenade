package user_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rolePostgres "github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/user"
	userPostgres "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	userAggregate "github.com/basilex/promenade/internal/contexts/identity/user/aggregate"
	userUseCase "github.com/basilex/promenade/internal/contexts/identity/user/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestUserUseCase_RegisterAndAuthenticate tests user registration and authentication
func TestUserUseCase_RegisterAndAuthenticate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register new user
		email := fmt.Sprintf("test_%s@example.com", uuidv7.New().String())
		name := "Test User"
		password := "password123"

		registeredUser, err := uc.Register(ctx, email, name, password)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, registeredUser.ID)
		assert.Equal(t, email, registeredUser.Email.Value())
		assert.Equal(t, userAggregate.UserStatusActive, registeredUser.Status)
		assert.False(t, registeredUser.EmailVerified)

		// Authenticate with correct password
		authenticatedUser, err := uc.Authenticate(ctx, email, password)
		require.NoError(t, err)
		assert.Equal(t, registeredUser.ID, authenticatedUser.ID)
		assert.Equal(t, email, authenticatedUser.Email.Value())

		// Authenticate with wrong password
		_, err = uc.Authenticate(ctx, email, "wrongpassword")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrInvalidCredentials))
	})
}

// TestUserUseCase_GetUserAndGetUserByEmail tests user retrieval methods
func TestUserUseCase_GetUserAndGetUserByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register user
		email := fmt.Sprintf("getuser_%s@example.com", uuidv7.New().String())
		name := "Get User Test"
		password := "password123"

		registeredUser, err := uc.Register(ctx, email, name, password)
		require.NoError(t, err)

		// GetUser by ID
		retrievedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.Equal(t, registeredUser.ID, retrievedUser.ID)
		assert.Equal(t, email, retrievedUser.Email.Value())

		// GetUserByEmail
		userByEmail, err := uc.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, registeredUser.ID, userByEmail.ID)
		assert.Equal(t, email, userByEmail.Email.Value())

		// GetUser with non-existent ID
		_, err = uc.GetUser(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, user.ErrUserNotFound, err)

		// GetUserByEmail with non-existent email
		_, err = uc.GetUserByEmail(ctx, fmt.Sprintf("nonexistent_%s@example.com", uuidv7.New().String()))
		assert.Error(t, err)
		assert.Equal(t, user.ErrUserNotFound, err)
	})
}

// TestUserUseCase_VerifyEmail tests email verification
func TestUserUseCase_VerifyEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register user
		registeredUser, err := uc.Register(ctx, fmt.Sprintf("verify_%s@example.com", uuidv7.New().String()), "Verify User", "password123")
		require.NoError(t, err)
		assert.False(t, registeredUser.EmailVerified)

		// Verify email
		err = uc.VerifyEmail(ctx, registeredUser.ID)
		require.NoError(t, err)

		// Check verification status
		verifiedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.True(t, verifiedUser.EmailVerified)

		// Verify non-existent user
		err = uc.VerifyEmail(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrUserNotFound))
	})
}

// TestUserUseCase_ChangePassword tests password change functionality
func TestUserUseCase_ChangePassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register user
		email := fmt.Sprintf("changepass_%s@example.com", uuidv7.New().String())
		oldPassword := "oldpassword123"
		newPassword := "newpassword456"

		registeredUser, err := uc.Register(ctx, email, "Change Pass User", oldPassword)
		require.NoError(t, err)

		// Change password with correct old password
		err = uc.ChangePassword(ctx, registeredUser.ID, oldPassword, newPassword)
		require.NoError(t, err)

		// Authenticate with new password
		_, err = uc.Authenticate(ctx, email, newPassword)
		require.NoError(t, err)

		// Authenticate with old password should fail
		_, err = uc.Authenticate(ctx, email, oldPassword)
		assert.Error(t, err)

		// Change password with wrong old password
		err = uc.ChangePassword(ctx, registeredUser.ID, "wrongoldpassword", "anotherpassword")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrInvalidCredentials))

		// Change password for non-existent user
		err = uc.ChangePassword(ctx, uuidv7.New(), oldPassword, newPassword)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrUserNotFound))
	})
}

// TestUserUseCase_SuspendUser tests user suspension
func TestUserUseCase_SuspendUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register user
		email := fmt.Sprintf("suspend_%s@example.com", uuidv7.New().String())
		password := "password123"
		registeredUser, err := uc.Register(ctx, email, "Suspend User", password)
		require.NoError(t, err)
		assert.Equal(t, userAggregate.UserStatusActive, registeredUser.Status)

		// Suspend user
		err = uc.SuspendUser(ctx, registeredUser.ID)
		require.NoError(t, err)

		// Check suspended status
		suspendedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.Equal(t, userAggregate.UserStatusSuspended, suspendedUser.Status)

		// Authenticate suspended user should fail
		_, err = uc.Authenticate(ctx, email, password)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrAccountNotActive))

		// Suspend non-existent user
		err = uc.SuspendUser(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrUserNotFound))
	})
}

// TestUserUseCase_BanUser tests user ban functionality
func TestUserUseCase_BanUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register user
		email := fmt.Sprintf("ban_%s@example.com", uuidv7.New().String())
		password := "password123"
		registeredUser, err := uc.Register(ctx, email, "Ban User", password)
		require.NoError(t, err)

		// Ban user
		err = uc.BanUser(ctx, registeredUser.ID)
		require.NoError(t, err)

		// Check banned status
		bannedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.Equal(t, userAggregate.UserStatusBanned, bannedUser.Status)

		// Authenticate banned user should fail
		_, err = uc.Authenticate(ctx, email, password)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrAccountNotActive))

		// Ban non-existent user
		err = uc.BanUser(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrUserNotFound))
	})
}

// TestUserUseCase_ActivateUser tests user activation
func TestUserUseCase_ActivateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register and suspend user
		email := fmt.Sprintf("activate_%s@example.com", uuidv7.New().String())
		password := "password123"
		registeredUser, err := uc.Register(ctx, email, "Activate User", password)
		require.NoError(t, err)

		err = uc.SuspendUser(ctx, registeredUser.ID)
		require.NoError(t, err)

		// Activate user
		err = uc.ActivateUser(ctx, registeredUser.ID)
		require.NoError(t, err)

		// Check active status
		activatedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.Equal(t, userAggregate.UserStatusActive, activatedUser.Status)

		// Authenticate activated user should work
		_, err = uc.Authenticate(ctx, email, password)
		require.NoError(t, err)

		// Activate non-existent user
		err = uc.ActivateUser(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrUserNotFound))
	})
}

// TestUserUseCase_UnlockUser tests manual unlock functionality
func TestUserUseCase_UnlockUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		// Register user
		email := fmt.Sprintf("unlock_%s@example.com", uuidv7.New().String())
		password := "password123"
		registeredUser, err := uc.Register(ctx, email, "Unlock User", password)
		require.NoError(t, err)

		// Trigger auto-lock by multiple failed login attempts
		for i := 0; i < 5; i++ {
			_, _ = uc.Authenticate(ctx, email, "wrongpassword")
		}

		// Verify user is locked
		lockedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.True(t, lockedUser.IsLocked())

		// Try to authenticate locked user
		_, err = uc.Authenticate(ctx, email, password)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrAccountLocked))

		// Unlock user
		err = uc.UnlockUser(ctx, registeredUser.ID)
		require.NoError(t, err)

		// Check unlocked status
		unlockedUser, err := uc.GetUser(ctx, registeredUser.ID)
		require.NoError(t, err)
		assert.False(t, unlockedUser.IsLocked())
		assert.Equal(t, 0, unlockedUser.FailedLoginCount)

		// Authenticate unlocked user should work
		_, err = uc.Authenticate(ctx, email, password)
		require.NoError(t, err)

		// Unlock non-existent user
		err = uc.UnlockUser(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrUserNotFound))
	})
}

// TestUserUseCase_ListUsers tests user listing with pagination
func TestUserUseCase_ListUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	ctx := context.Background()

	userRepo := userPostgres.NewUserRepository(testDB.DB)
	roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
	uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

	// Create multiple users with unique emails
	testID := uuidv7.New().String()
	for i := 1; i <= 5; i++ {
		email := fmt.Sprintf("listuser%s_%d@example.com", testID, i)
		name := fmt.Sprintf("List User %d", i)
		_, err := uc.Register(ctx, email, name, "password123")
		require.NoError(t, err)
	}

	// List first page
	users, total, err := uc.ListUsers(ctx, 2, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 2)
	assert.GreaterOrEqual(t, total, 5)

	// List second page
	users, total, err = uc.ListUsers(ctx, 2, 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 1)
	assert.GreaterOrEqual(t, total, 5)
}

// TestUserUseCase_DuplicateEmailRegistration tests duplicate email handling
func TestUserUseCase_DuplicateEmailRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		email := fmt.Sprintf("duplicate_%s@example.com", uuidv7.New().String())

		// Register first user
		_, err := uc.Register(ctx, email, "User 1", "password123")
		require.NoError(t, err)

		// Try to register with same email
		_, err = uc.Register(ctx, email, "User 2", "password456")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrEmailAlreadyExists))
	})
}

// TestUserUseCase_CompleteWorkflow tests complete user lifecycle
func TestUserUseCase_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := userUseCase.NewUserUseCase(userRepo, roleRepo)

		email := fmt.Sprintf("workflow_%s@example.com", uuidv7.New().String())
		password := "password123"
		newPassword := "newpassword456"

		// 1. Register
		u, err := uc.Register(ctx, email, "Workflow User", password)
		require.NoError(t, err)
		assert.False(t, u.EmailVerified)
		assert.Equal(t, userAggregate.UserStatusActive, u.Status)

		// 2. Authenticate
		_, err = uc.Authenticate(ctx, email, password)
		require.NoError(t, err)

		// 3. Verify email
		err = uc.VerifyEmail(ctx, u.ID)
		require.NoError(t, err)

		u, _ = uc.GetUser(ctx, u.ID)
		assert.True(t, u.EmailVerified)

		// 4. Change password
		err = uc.ChangePassword(ctx, u.ID, password, newPassword)
		require.NoError(t, err)

		// 5. Authenticate with new password
		_, err = uc.Authenticate(ctx, email, newPassword)
		require.NoError(t, err)

		// 6. Suspend
		err = uc.SuspendUser(ctx, u.ID)
		require.NoError(t, err)

		u, _ = uc.GetUser(ctx, u.ID)
		assert.Equal(t, userAggregate.UserStatusSuspended, u.Status)

		// 7. Reactivate
		err = uc.ActivateUser(ctx, u.ID)
		require.NoError(t, err)

		u, _ = uc.GetUser(ctx, u.ID)
		assert.Equal(t, userAggregate.UserStatusActive, u.Status)

		// 8. Ban
		err = uc.BanUser(ctx, u.ID)
		require.NoError(t, err)

		u, _ = uc.GetUser(ctx, u.ID)
		assert.Equal(t, userAggregate.UserStatusBanned, u.Status)
	})
}
