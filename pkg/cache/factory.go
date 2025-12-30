package cache

import (
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/basilex/promenade/pkg/cache/noop"
	rediscache "github.com/basilex/promenade/pkg/cache/redis"
)

// NewCache creates a cache instance based on configuration
func NewCache(cfg *Config, redisClient *redis.Client) (ICache, error) {
	// If caching is disabled, return no-op cache
	if !cfg.Enabled {
		return noop.NewNoOpCache(), nil
	}

	switch cfg.Adapter {
	case "redis":
		if redisClient == nil {
			return nil, fmt.Errorf("redis client is required for redis adapter")
		}
		return rediscache.NewRedisCache(redisClient, cfg.Prefix), nil

	case "noop", "":
		return noop.NewNoOpCache(), nil

	default:
		return nil, fmt.Errorf("unsupported cache adapter: %s", cfg.Adapter)
	}
}
