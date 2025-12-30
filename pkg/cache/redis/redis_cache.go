// Package redis provides a Redis-based cache implementation
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/basilex/promenade/pkg/logger"
)

// RedisCache implements cache.Cache using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// cacheMissError represents a cache miss error
type cacheMissError struct {
	Key string
}

func (e *cacheMissError) Error() string {
	return "cache miss: " + e.Key
}

// Get retrieves a value from Redis cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.prefix + key

	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err == redis.Nil {
		return &cacheMissError{Key: key}
	}
	if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// Set stores a value in Redis cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	err = c.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	logger.FromContext(ctx).Debug("Cache set",
		"key", key,
		"ttl", ttl.String(),
	)

	return nil
}

// Delete removes a single key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.prefix + key

	err := c.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}

	logger.FromContext(ctx).Debug("Cache delete", "key", key)

	return nil
}

// DeletePattern removes all keys matching the pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.prefix + pattern

	// Use SCAN for production-safe deletion (doesn't block Redis)
	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan failed: %w", err)
	}

	if len(keys) > 0 {
		err := c.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("redis delete pattern failed: %w", err)
		}

		logger.FromContext(ctx).Debug("Cache delete pattern",
			"pattern", pattern,
			"keys_deleted", len(keys),
		)
	}

	return nil
}

// Clear removes all keys with the prefix (use with caution)
func (c *RedisCache) Clear(ctx context.Context) error {
	return c.DeletePattern(ctx, "*")
}

// Health checks Redis connection
func (c *RedisCache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close gracefully closes Redis connection
func (c *RedisCache) Close(ctx context.Context) error {
	return c.client.Close()
}
