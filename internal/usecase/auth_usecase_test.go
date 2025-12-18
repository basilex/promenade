package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
)

func TestAuthUseCase_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		email := "test@example.com"
		name := "Test User"
		password := "password123"

		mockUserRepo.On("GetByEmail", ctx, email).Return(nil, entity.ErrNotFound)
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

		user, err := uc.Register(ctx, email, name, password)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, entity.UserStatusUnverified, user.Status)
		assert.NotEmpty(t, user.Password)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		existingUser := &entity.User{
			ID:    uuidv7.New(),
			Email: "test@example.com",
		}

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(existingUser, nil)

		_, err := uc.Register(ctx, "test@example.com", "Test User", "password")
		assert.ErrorIs(t, err, ErrEmailAlreadyExists)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		// Create user with hashed password
		user := &entity.User{
			ID:     uuidv7.New(),
			Email:  "test@example.com",
			Status: entity.UserStatusActive,
		}
		_ = user.HashPassword("password123")

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		mockSessionRepo.On("CountUserSessions", ctx, user.ID).Return(0, nil)
		mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)
		mockUserRepo.On("UpdateLastLogin", ctx, user.ID, mock.AnythingOfType("time.Time")).Return(nil)

		accessToken, refreshToken, returnedUser, err := uc.Login(ctx, "test@example.com", "password123", "Test UA", "127.0.0.1")
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, user, returnedUser)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		user := &entity.User{
			ID:     uuidv7.New(),
			Email:  "test@example.com",
			Status: entity.UserStatusActive,
		}
		_ = user.HashPassword("correct-password")

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		_, _, _, err := uc.Login(ctx, "test@example.com", "wrong-password", "UA", "IP")
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials - user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("GetByEmail", ctx, "notfound@example.com").Return(nil, entity.ErrNotFound)

		_, _, _, err := uc.Login(ctx, "notfound@example.com", "password", "UA", "IP")
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user suspended", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		user := &entity.User{
			ID:     uuidv7.New(),
			Email:  "test@example.com",
			Status: entity.UserStatusSuspended,
		}
		_ = user.HashPassword("password")

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		_, _, _, err := uc.Login(ctx, "test@example.com", "password", "UA", "IP")
		assert.ErrorIs(t, err, ErrUserSuspended)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user banned", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		user := &entity.User{
			ID:     uuidv7.New(),
			Email:  "test@example.com",
			Status: entity.UserStatusBanned,
		}
		_ = user.HashPassword("password")

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		_, _, _, err := uc.Login(ctx, "test@example.com", "password", "UA", "IP")
		assert.ErrorIs(t, err, ErrUserBanned)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user inactive", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		user := &entity.User{
			ID:     uuidv7.New(),
			Email:  "test@example.com",
			Status: entity.UserStatusInactive,
		}
		_ = user.HashPassword("password")

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		_, _, _, err := uc.Login(ctx, "test@example.com", "password", "UA", "IP")
		assert.ErrorIs(t, err, ErrUserNotActive)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("successful logout", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		refreshToken := "valid-refresh-token"
		session := &entity.Session{
			ID:     uuidv7.New(),
			UserID: uuidv7.New(),
		}

		mockSessionRepo.On("GetByRefreshToken", ctx, mock.AnythingOfType("string")).Return(session, nil)
		mockSessionRepo.On("Delete", ctx, session.ID).Return(nil)

		err := uc.Logout(ctx, refreshToken)
		require.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockSessionRepo.On("GetByRefreshToken", ctx, mock.AnythingOfType("string")).Return(nil, entity.ErrNotFound)

		err := uc.Logout(ctx, "invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("successful token refresh", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		userID := uuidv7.New()
		user := &entity.User{
			ID:     userID,
			Email:  "test@example.com",
			Status: entity.UserStatusActive,
		}

		session := &entity.Session{
			ID:        uuidv7.New(),
			UserID:    userID,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockSessionRepo.On("GetByRefreshToken", ctx, mock.AnythingOfType("string")).Return(session, nil)
		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)

		newAccessToken, newRefreshToken, err := uc.RefreshToken(ctx, "old-refresh-token")
		require.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		session := &entity.Session{
			ID:        uuidv7.New(),
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}

		mockSessionRepo.On("GetByRefreshToken", ctx, mock.AnythingOfType("string")).Return(session, nil)
		mockSessionRepo.On("Delete", ctx, session.ID).Return(nil)

		_, _, err := uc.RefreshToken(ctx, "expired-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockSessionRepo.On("GetByRefreshToken", ctx, mock.AnythingOfType("string")).Return(nil, entity.ErrNotFound)

		_, _, err := uc.RefreshToken(ctx, "invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("user cannot login anymore", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		userID := uuidv7.New()
		user := &entity.User{
			ID:     userID,
			Status: entity.UserStatusBanned, // User banned after session created
		}

		session := &entity.Session{
			ID:        uuidv7.New(),
			UserID:    userID,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockSessionRepo.On("GetByRefreshToken", ctx, mock.AnythingOfType("string")).Return(session, nil)
		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

		_, _, err := uc.RefreshToken(ctx, "token")
		assert.ErrorIs(t, err, ErrUserNotActive)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_GetMe(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		expectedUser := &entity.User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		user, err := uc.GetMe(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("GetByID", ctx, userID).Return(nil, entity.ErrNotFound)

		_, err := uc.GetMe(ctx, userID)
		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_GetUserSessions(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		sessions := []*entity.Session{
			{ID: uuidv7.New(), UserID: userID},
			{ID: uuidv7.New(), UserID: userID},
		}

		mockSessionRepo.On("GetUserSessions", ctx, userID).Return(sessions, nil)

		result, err := uc.GetUserSessions(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockSessionRepo.On("GetUserSessions", ctx, userID).Return(nil, errors.New("db error"))

		_, err := uc.GetUserSessions(ctx, userID)
		assert.Error(t, err)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_SuspendUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	reason := "Violation of terms"

	t.Run("successful suspension", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		until := time.Now().Add(7 * 24 * time.Hour)

		mockUserRepo.On("Suspend", ctx, userID, reason, &until).Return(nil)
		mockSessionRepo.On("DeleteByUserID", ctx, userID).Return(nil)

		err := uc.SuspendUser(ctx, userID, reason, &until)
		require.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("permanent suspension", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("Suspend", ctx, userID, reason, (*time.Time)(nil)).Return(nil)
		mockSessionRepo.On("DeleteByUserID", ctx, userID).Return(nil)

		err := uc.SuspendUser(ctx, userID, reason, nil)
		require.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_BanUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	reason := "Severe violation"

	t.Run("successful ban", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("Ban", ctx, userID, reason).Return(nil)
		mockSessionRepo.On("DeleteByUserID", ctx, userID).Return(nil)

		err := uc.BanUser(ctx, userID, reason)
		require.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("Ban", ctx, userID, reason).Return(errors.New("db error"))

		err := uc.BanUser(ctx, userID, reason)
		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_ReactivateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful reactivation", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("Reactivate", ctx, userID).Return(nil)

		err := uc.ReactivateUser(ctx, userID)
		require.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockSessionRepo := new(mocks.MockSessionRepository)
		jwtManager := jwtpkg.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
		uc := NewAuthUseCase(mockUserRepo, mockSessionRepo, jwtManager)

		mockUserRepo.On("Reactivate", ctx, userID).Return(errors.New("db error"))

		err := uc.ReactivateUser(ctx, userID)
		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}
