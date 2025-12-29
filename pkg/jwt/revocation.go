package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenRevoker handles token revocation using Redis blacklist.
// Tokens are stored in Redis with TTL matching their expiration time.
//
// Thread-safe: Redis operations are atomic.
//
// Example:
//
//	revoker := NewTokenRevoker(redisClient)
//	err := revoker.Revoke(ctx, token, expiresAt)
type TokenRevoker struct {
	client *redis.Client
	prefix string
}

// NewTokenRevoker creates a new token revoker with Redis client.
//
// Parameters:
//   - client: Redis client instance
//
// Returns TokenRevoker instance.
func NewTokenRevoker(client *redis.Client) *TokenRevoker {
	return &TokenRevoker{
		client: client,
		prefix: "revoked_token:",
	}
}

// Revoke adds a token to the blacklist with TTL matching its expiration.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - tokenString: JWT token string to revoke
//   - expiresAt: Token expiration time (for TTL calculation)
//
// Returns error if Redis operation fails.
//
// Example:
//
//	err := revoker.Revoke(ctx, token, time.Now().Add(15*time.Minute))
func (r *TokenRevoker) Revoke(ctx context.Context, tokenString string, expiresAt time.Time) error {
	// Calculate TTL (time until token expires)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to revoke
		return nil
	}

	// Store token in Redis with TTL
	key := r.prefix + tokenString
	err := r.client.Set(ctx, key, "revoked", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// IsRevoked checks if a token is in the blacklist.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - tokenString: JWT token string to check
//
// Returns true if token is revoked, false otherwise.
//
// Example:
//
//	revoked, err := revoker.IsRevoked(ctx, token)
//	if revoked {
//	    return errors.New("token has been revoked")
//	}
func (r *TokenRevoker) IsRevoked(ctx context.Context, tokenString string) (bool, error) {
	key := r.prefix + tokenString
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}

	return exists > 0, nil
}

// RevokeAllForUser revokes all tokens for a specific user.
// This is useful when user logs out from all devices or account is compromised.
//
// Note: This requires storing user_id → token mapping in Redis.
// Current implementation doesn't support this (would need additional Redis structure).
//
// TODO: Implement user token tracking if needed.
func (r *TokenRevoker) RevokeAllForUser(ctx context.Context, userID string) error {
	// Not implemented yet - would require tracking all user tokens
	return fmt.Errorf("not implemented: RevokeAllForUser requires token tracking")
}

// Stats returns blacklist statistics (for monitoring).
//
// Returns:
//   - count: Number of revoked tokens in Redis
//   - error: If Redis operation fails
func (r *TokenRevoker) Stats(ctx context.Context) (int64, error) {
	// Count keys with revoked_token: prefix
	iter := r.client.Scan(ctx, 0, r.prefix+"*", 0).Iterator()
	count := int64(0)

	for iter.Next(ctx) {
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to get stats: %w", err)
	}

	return count, nil
}
