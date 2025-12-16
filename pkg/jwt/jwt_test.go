package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	manager := NewJWTManager(secretKey, accessTTL, refreshTTL)

	assert.NotNil(t, manager)
	assert.Equal(t, accessTTL, manager.GetAccessTokenTTL())
	assert.Equal(t, refreshTTL, manager.GetRefreshTokenTTL())
}

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	tokenPair, err := manager.GenerateTokenPair(userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.NotZero(t, tokenPair.ExpiresAt)
	assert.True(t, tokenPair.ExpiresAt.After(time.Now()))
}

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	token, expiresAt, err := manager.GenerateAccessToken(userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotZero(t, expiresAt)
	assert.True(t, expiresAt.After(time.Now()))
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), expiresAt, 1*time.Second)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	t.Run("validates valid token", func(t *testing.T) {
		token, _, err := manager.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		shortManager := NewJWTManager("test-secret", 1*time.Millisecond, 7*24*time.Hour)
		token, _, err := shortManager.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		_, err = shortManager.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid.token.string")
		assert.Error(t, err)
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		token, _, err := manager.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		wrongManager := NewJWTManager("wrong-secret", 15*time.Minute, 7*24*time.Hour)
		_, err = wrongManager.ValidateToken(token)
		assert.Error(t, err)
	})

	t.Run("rejects empty token", func(t *testing.T) {
		_, err := manager.ValidateToken("")
		assert.Error(t, err)
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		// Token with invalid structure
		malformedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"
		_, err := manager.ValidateToken(malformedToken)
		assert.Error(t, err)
	})
}

func TestJWTManager_RefreshAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	t.Run("refreshes access token successfully", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair(userID, email)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond) // Ensure different IssuedAt

		newAccessToken, expiresAt, err := manager.RefreshAccessToken(tokenPair.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.True(t, expiresAt.After(time.Now()))
	})

	t.Run("rejects expired refresh token", func(t *testing.T) {
		shortManager := NewJWTManager("test-secret", 15*time.Minute, 1*time.Millisecond)
		tokenPair, err := shortManager.GenerateTokenPair(userID, email)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		_, _, err = shortManager.RefreshAccessToken(tokenPair.RefreshToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("rejects invalid refresh token", func(t *testing.T) {
		_, _, err := manager.RefreshAccessToken("invalid.token.string")
		assert.Error(t, err)
	})
}

func TestJWTManager_TokenClaims(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	token, expiresAt, err := manager.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	t.Run("contains correct user ID", func(t *testing.T) {
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("contains correct email", func(t *testing.T) {
		assert.Equal(t, email, claims.Email)
	})

	t.Run("has valid expiration", func(t *testing.T) {
		assert.NotNil(t, claims.ExpiresAt)
		assert.WithinDuration(t, expiresAt, claims.ExpiresAt.Time, 1*time.Second)
	})

	t.Run("has issued at timestamp", func(t *testing.T) {
		assert.NotNil(t, claims.IssuedAt)
		assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, 1*time.Second)
	})

	t.Run("has not before timestamp", func(t *testing.T) {
		assert.NotNil(t, claims.NotBefore)
		assert.WithinDuration(t, time.Now(), claims.NotBefore.Time, 1*time.Second)
	})
}

func TestJWTManager_TTLGetters(t *testing.T) {
	accessTTL := 30 * time.Minute
	refreshTTL := 14 * 24 * time.Hour

	manager := NewJWTManager("test-secret", accessTTL, refreshTTL)

	t.Run("returns correct access token TTL", func(t *testing.T) {
		assert.Equal(t, accessTTL, manager.GetAccessTokenTTL())
	})

	t.Run("returns correct refresh token TTL", func(t *testing.T) {
		assert.Equal(t, refreshTTL, manager.GetRefreshTokenTTL())
	})
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	t.Run("access token expires after TTL", func(t *testing.T) {
		manager := NewJWTManager("test-secret", 100*time.Millisecond, 7*24*time.Hour)
		userID := uuid.New()
		email := "test@example.com"

		token, _, err := manager.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		// Wait for token to expire (add extra buffer for clock skew)
		time.Sleep(150 * time.Millisecond)

		// Token should be expired now
		_, err = manager.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("refresh token expires after TTL", func(t *testing.T) {
		manager := NewJWTManager("test-secret", 15*time.Minute, 100*time.Millisecond)
		userID := uuid.New()
		email := "test@example.com"

		tokenPair, err := manager.GenerateTokenPair(userID, email)
		require.NoError(t, err)

		// Wait for refresh token to expire (add extra buffer for clock skew)
		time.Sleep(150 * time.Millisecond)

		// Refresh token should be expired
		_, err = manager.ValidateToken(tokenPair.RefreshToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

func TestJWTManager_MultipleTokens(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	user1ID := uuid.New()
	user1Email := "user1@example.com"
	user2ID := uuid.New()
	user2Email := "user2@example.com"

	token1, _, err := manager.GenerateAccessToken(user1ID, user1Email)
	require.NoError(t, err)

	token2, _, err := manager.GenerateAccessToken(user2ID, user2Email)
	require.NoError(t, err)

	t.Run("validates first user token correctly", func(t *testing.T) {
		claims, err := manager.ValidateToken(token1)
		require.NoError(t, err)
		assert.Equal(t, user1ID, claims.UserID)
		assert.Equal(t, user1Email, claims.Email)
	})

	t.Run("validates second user token correctly", func(t *testing.T) {
		claims, err := manager.ValidateToken(token2)
		require.NoError(t, err)
		assert.Equal(t, user2ID, claims.UserID)
		assert.Equal(t, user2Email, claims.Email)
	})

	t.Run("tokens are different", func(t *testing.T) {
		assert.NotEqual(t, token1, token2)
	})
}

func TestJWTManager_ErrorConstants(t *testing.T) {
	t.Run("ErrInvalidToken is defined", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidToken)
		assert.Equal(t, "invalid token", ErrInvalidToken.Error())
	})

	t.Run("ErrExpiredToken is defined", func(t *testing.T) {
		assert.NotNil(t, ErrExpiredToken)
		assert.Equal(t, "token has expired", ErrExpiredToken.Error())
	})
}
