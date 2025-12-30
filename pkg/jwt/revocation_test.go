package jwt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, mr
}

func TestNewTokenRevoker(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	revoker := NewTokenRevoker(client)

	assert.NotNil(t, revoker)
	assert.NotNil(t, revoker.client)
	assert.Equal(t, "revoked_token:", revoker.prefix)
}

func TestTokenRevoker_Revoke(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer client.Close()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	token := "test-token-123"
	expiresAt := time.Now().Add(15 * time.Minute)

	// Revoke token
	err := revoker.Revoke(ctx, token, expiresAt)
	require.NoError(t, err)

	// Verify token is revoked
	revoked, err := revoker.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestTokenRevoker_RevokeExpiredToken(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	token := "expired-token"
	expiresAt := time.Now().Add(-1 * time.Hour) // Already expired

	// Revoke expired token (should succeed but not store)
	err := revoker.Revoke(ctx, token, expiresAt)
	require.NoError(t, err)

	// Verify token is not in blacklist (already expired)
	revoked, err := revoker.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.False(t, revoked)
}

func TestTokenRevoker_IsRevoked_NotRevoked(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer client.Close()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	token := "not-revoked-token"

	// Check non-revoked token
	revoked, err := revoker.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.False(t, revoked)
}

func TestTokenRevoker_TTL(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	token := "ttl-test-token"
	expiresAt := time.Now().Add(5 * time.Second)

	// Revoke token
	err := revoker.Revoke(ctx, token, expiresAt)
	require.NoError(t, err)

	// Verify token is revoked
	revoked, err := revoker.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.True(t, revoked)

	// Fast-forward time in miniredis
	mr.FastForward(6 * time.Second)

	// Verify token is no longer in blacklist (expired)
	revoked, err = revoker.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.False(t, revoked)
}

func TestTokenRevoker_MultipleTokens(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer client.Close()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	tokens := []string{"token1", "token2", "token3"}
	expiresAt := time.Now().Add(15 * time.Minute)

	// Revoke multiple tokens
	for _, token := range tokens {
		err := revoker.Revoke(ctx, token, expiresAt)
		require.NoError(t, err)
	}

	// Verify all tokens are revoked
	for _, token := range tokens {
		revoked, err := revoker.IsRevoked(ctx, token)
		require.NoError(t, err)
		assert.True(t, revoked, "Token %s should be revoked", token)
	}
}

func TestTokenRevoker_Stats(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	// Initially zero
	count, err := revoker.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Revoke 3 tokens
	expiresAt := time.Now().Add(15 * time.Minute)
	for i := 0; i < 3; i++ {
		token := fmt.Sprintf("token-%d", i)
		err := revoker.Revoke(ctx, token, expiresAt)
		require.NoError(t, err)
	}

	// Verify count
	count, err = revoker.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestTokenRevoker_RevokeAllForUser_NotImplemented(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer client.Close()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	err := revoker.RevokeAllForUser(ctx, "user-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestTokenRevoker_ConcurrentRevocations(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() { _ = client.Close() }()

	revoker := NewTokenRevoker(client)
	ctx := context.Background()

	expiresAt := time.Now().Add(15 * time.Minute)
	done := make(chan bool)

	// Revoke 10 tokens concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := fmt.Sprintf("concurrent-token-%d", id)
			err := revoker.Revoke(ctx, token, expiresAt)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all tokens are revoked
	count, err := revoker.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(10), count)
}
