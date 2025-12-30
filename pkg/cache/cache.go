// Package cache provides a unified caching interface with multiple adapter implementations.
// It supports Redis-based caching with configurable TTL per resource type and graceful
// degradation when cache is unavailable.
package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from cache and unmarshals it into dest
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in cache with the given TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a single key from cache
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all keys matching the pattern (e.g., "user:*")
	DeletePattern(ctx context.Context, pattern string) error

	// Clear removes all keys from cache (use with caution)
	Clear(ctx context.Context) error

	// Health checks if the cache is operational
	Health(ctx context.Context) error

	// Close gracefully shuts down the cache
	Close(ctx context.Context) error
}

// Config holds cache configuration
type Config struct {
	// Enabled determines if caching is active
	Enabled bool

	// Adapter specifies which cache implementation to use ("redis" or "noop")
	Adapter string

	// Prefix is prepended to all cache keys for namespace isolation
	Prefix string

	// DefaultTTL is used when no specific TTL is configured
	DefaultTTL time.Duration

	// TTL holds TTL configuration per resource type
	TTL TTLConfig
}

// TTLConfig defines TTL durations for different resource types
type TTLConfig struct {
	// Reference data (rarely changes)
	Countries  time.Duration
	Currencies time.Duration
	Languages  time.Duration
	Timezones  time.Duration

	// User data (changes more frequently)
	UserProfile time.Duration
	Customer    time.Duration

	// Session data (short-lived)
	Session time.Duration
}

// ErrCacheMiss is returned when a key is not found in cache
type ErrCacheMiss struct {
	Key string
}

func (e *ErrCacheMiss) Error() string {
	return "cache miss: " + e.Key
}

// IsCacheMiss checks if an error is a cache miss
func IsCacheMiss(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*ErrCacheMiss)
	return ok
}

// GetTTL returns the TTL for a specific resource type
func (c *Config) GetTTL(resourceType string) time.Duration {
	switch resourceType {
	case "countries":
		return c.TTL.Countries
	case "currencies":
		return c.TTL.Currencies
	case "languages":
		return c.TTL.Languages
	case "timezones":
		return c.TTL.Timezones
	case "user_profile":
		return c.TTL.UserProfile
	case "customer":
		return c.TTL.Customer
	case "session":
		return c.TTL.Session
	default:
		return c.DefaultTTL
	}
}

// NewConfig creates a new cache config with sensible defaults
func NewConfig() *Config {
	return &Config{
		Enabled:    true,
		Adapter:    "redis",
		Prefix:     "promenade:",
		DefaultTTL: 5 * time.Minute,
		TTL: TTLConfig{
			Countries:   1 * time.Hour,
			Currencies:  1 * time.Hour,
			Languages:   1 * time.Hour,
			Timezones:   1 * time.Hour,
			UserProfile: 15 * time.Minute,
			Customer:    10 * time.Minute,
			Session:     30 * time.Minute,
		},
	}
}
