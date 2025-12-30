// Package noop provides a no-operation cache implementation
// Used when caching is disabled or Redis is unavailable
package noop

import (
	"context"
	"time"
)

// NoOpCache is a cache implementation that does nothing
// All operations succeed but have no effect
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache instance
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// cacheMissError is a local cache miss error type
type cacheMissError struct {
	Key string
}

func (e *cacheMissError) Error() string {
	return "cache miss: " + e.Key
}

// Get always returns cache miss
func (c *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return &cacheMissError{Key: key}
}

// Set does nothing and returns success
func (c *NoOpCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

// Delete does nothing and returns success
func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

// DeletePattern does nothing and returns success
func (c *NoOpCache) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

// Clear does nothing and returns success
func (c *NoOpCache) Clear(ctx context.Context) error {
	return nil
}

// Health always returns success
func (c *NoOpCache) Health(ctx context.Context) error {
	return nil
}

// Close does nothing and returns success
func (c *NoOpCache) Close(ctx context.Context) error {
	return nil
}
