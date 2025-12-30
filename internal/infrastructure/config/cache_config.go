package config

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/cache"
)

// ToCacheConfig converts YAML CacheSection to pkg/cache.Config
func (c *CacheSection) ToCacheConfig() (*cache.Config, error) {
	// Parse default TTL
	defaultTTL, err := time.ParseDuration(c.DefaultTTL)
	if err != nil {
		return nil, fmt.Errorf("invalid default_ttl: %w", err)
	}

	// Parse TTL for each resource type
	countriesTTL, err := time.ParseDuration(c.TTL.Countries)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.countries: %w", err)
	}

	currenciesTTL, err := time.ParseDuration(c.TTL.Currencies)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.currencies: %w", err)
	}

	languagesTTL, err := time.ParseDuration(c.TTL.Languages)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.languages: %w", err)
	}

	timezonesTTL, err := time.ParseDuration(c.TTL.Timezones)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.timezones: %w", err)
	}

	userProfileTTL, err := time.ParseDuration(c.TTL.UserProfile)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.user_profile: %w", err)
	}

	customerTTL, err := time.ParseDuration(c.TTL.Customer)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.customer: %w", err)
	}

	sessionTTL, err := time.ParseDuration(c.TTL.Session)
	if err != nil {
		return nil, fmt.Errorf("invalid ttl.session: %w", err)
	}

	return &cache.Config{
		Enabled:    c.Enabled,
		Adapter:    c.Adapter,
		Prefix:     c.Prefix,
		DefaultTTL: defaultTTL,
		TTL: cache.TTLConfig{
			Countries:   countriesTTL,
			Currencies:  currenciesTTL,
			Languages:   languagesTTL,
			Timezones:   timezonesTTL,
			UserProfile: userProfileTTL,
			Customer:    customerTTL,
			Session:     sessionTTL,
		},
	}, nil
}
