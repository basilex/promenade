package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewManager(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		config := Config{SecretKey: "test-secret"}
		manager := NewManager(config)

		assert.Equal(t, "test-secret", manager.config.SecretKey)
		assert.Equal(t, 15*time.Minute, manager.config.AccessTokenDuration)
		assert.Equal(t, 7*24*time.Hour, manager.config.RefreshTokenDuration)
		assert.Equal(t, "promenade-platform", manager.config.Issuer)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := Config{
			SecretKey:            "custom-secret",
			AccessTokenDuration:  30 * time.Minute,
			RefreshTokenDuration: 14 * 24 * time.Hour,
			Issuer:               "custom-issuer",
		}
		manager := NewManager(config)

		assert.Equal(t, "custom-secret", manager.config.SecretKey)
		assert.Equal(t, 30*time.Minute, manager.config.AccessTokenDuration)
		assert.Equal(t, 14*24*time.Hour, manager.config.RefreshTokenDuration)
		assert.Equal(t, "custom-issuer", manager.config.Issuer)
	})
}

func TestManager_GenerateTokenPair(t *testing.T) {
	manager := NewManager(Config{SecretKey: "test-secret"})
	userID := uuidv7.New()
	email := "test@example.com"
	roles := []string{"user", "admin"}

	t.Run("generates valid token pair", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair(userID, email, roles)

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.Equal(t, "Bearer", tokenPair.TokenType)
		assert.True(t, tokenPair.ExpiresAt.After(time.Now()))
	})
}

func TestManager_ValidateAccessToken(t *testing.T) {
	manager := NewManager(Config{SecretKey: "test-secret"})
	userID := uuidv7.New()
	email := "test@example.com"
	roles := []string{"user"}

	t.Run("validates correct token", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair(userID, email, roles)
		assert.NoError(t, err)

		claims, err := manager.ValidateAccessToken(tokenPair.AccessToken)

		assert.NoError(t, err)
		assert.Equal(t, userID.String(), claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roles, claims.Roles)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		managerShort := NewManager(Config{
			SecretKey:           "test-secret",
			AccessTokenDuration: 1 * time.Nanosecond,
		})
		tokenPair, err := managerShort.GenerateTokenPair(userID, email, roles)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		_, err = manager.ValidateAccessToken(tokenPair.AccessToken)
		assert.Error(t, err)
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair(userID, email, roles)
		assert.NoError(t, err)

		wrongManager := NewManager(Config{SecretKey: "wrong-secret"})
		_, err = wrongManager.ValidateAccessToken(tokenPair.AccessToken)

		assert.Error(t, err)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := manager.ValidateAccessToken("invalid-token")
		assert.Error(t, err)
	})
}

func TestManager_RefreshAccessToken(t *testing.T) {
	manager := NewManager(Config{SecretKey: "test-secret"})
	userID := uuidv7.New()
	email := "test@example.com"
	roles := []string{"user"}

	t.Run("refreshes with valid token", func(t *testing.T) {
		tokenPair, err := manager.GenerateTokenPair(userID, email, roles)
		assert.NoError(t, err)

		newTokenPair, err := manager.RefreshAccessToken(tokenPair.RefreshToken)

		assert.NoError(t, err)
		assert.NotEmpty(t, newTokenPair.AccessToken)
		assert.NotEmpty(t, newTokenPair.RefreshToken)
		assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	})

	t.Run("rejects invalid refresh token", func(t *testing.T) {
		_, err := manager.RefreshAccessToken("invalid-token")
		assert.Error(t, err)
	})
}

func TestClaims_ExtractUserID(t *testing.T) {
	userID := uuidv7.New()

	t.Run("extracts valid UUID", func(t *testing.T) {
		claims := &Claims{UserID: userID.String()}

		extractedID, err := claims.ExtractUserID()

		assert.NoError(t, err)
		assert.Equal(t, userID, extractedID)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		claims := &Claims{UserID: "invalid-uuid"}

		_, err := claims.ExtractUserID()

		assert.Error(t, err)
	})
}

func TestClaims_HasRole(t *testing.T) {
	claims := &Claims{Roles: []string{"user", "admin"}}

	t.Run("finds existing role", func(t *testing.T) {
		assert.True(t, claims.HasRole("user"))
		assert.True(t, claims.HasRole("admin"))
	})

	t.Run("does not find non-existing role", func(t *testing.T) {
		assert.False(t, claims.HasRole("superadmin"))
	})

	t.Run("handles empty roles", func(t *testing.T) {
		emptyClaims := &Claims{Roles: []string{}}
		assert.False(t, emptyClaims.HasRole("user"))
	})
}

func TestClaims_HasAnyRole(t *testing.T) {
	claims := &Claims{Roles: []string{"user", "admin"}}

	t.Run("finds when has one role", func(t *testing.T) {
		assert.True(t, claims.HasAnyRole("user", "superadmin"))
		assert.True(t, claims.HasAnyRole("moderator", "admin"))
	})

	t.Run("does not find when has none", func(t *testing.T) {
		assert.False(t, claims.HasAnyRole("superadmin", "moderator"))
	})

	t.Run("handles empty input", func(t *testing.T) {
		assert.False(t, claims.HasAnyRole())
	})
}

func TestClaims_HasAllRoles(t *testing.T) {
	claims := &Claims{Roles: []string{"user", "admin"}}

	t.Run("finds when has all roles", func(t *testing.T) {
		assert.True(t, claims.HasAllRoles("user", "admin"))
		assert.True(t, claims.HasAllRoles("user"))
	})

	t.Run("does not find when missing some", func(t *testing.T) {
		assert.False(t, claims.HasAllRoles("user", "admin", "superadmin"))
	})

	t.Run("handles empty input", func(t *testing.T) {
		assert.True(t, claims.HasAllRoles())
	})
}
