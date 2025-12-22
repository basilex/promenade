package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents a module's configuration
type Config struct {
	Module   ModuleSection          `yaml:"module"`
	Settings map[string]interface{} `yaml:"settings"`
	Purge    PurgeSection           `yaml:"purge"`
}

// ModuleSection contains module metadata
type ModuleSection struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// PurgeSection contains purge configuration
type PurgeSection struct {
	Enabled  bool                          `yaml:"enabled"`
	Schedule string                        `yaml:"schedule"`
	Settings map[string]map[string]interface{} `yaml:"settings"`
}

// Load loads module configuration from the specified directory
// It automatically detects the environment and loads the appropriate config file
func Load(configDir string, appName string) (*Config, error) {
	// Detect environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Map environment to file suffix
	suffix := getEnvSuffix(env)

	// Try environment-specific config file (e.g., config.dev.yaml)
	configPath := filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", suffix))
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Fall back to generic config.yaml
		configPath = filepath.Join(configDir, "config.yaml")
		data, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load module config from %s: %w", configDir, err)
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse module config: %w", err)
	}

	return &config, nil
}

// getEnvSuffix returns the file suffix for the given environment
func getEnvSuffix(env string) string {
	switch env {
	case "development":
		return "dev"
	case "test", "testing":
		return "test"
	case "production", "prod":
		return "prod"
	default:
		return "dev"
	}
}

// GetRetentionDays returns the retention days for a specific entity
// Falls back to defaultDays if not found
func (c *Config) GetRetentionDays(entityName string, defaultDays int) int {
	if c.Purge.Settings == nil {
		return defaultDays
	}

	entitySettings, ok := c.Purge.Settings[entityName]
	if !ok {
		return defaultDays
	}

	if days, ok := entitySettings["retention_days"].(int); ok {
		return days
	}

	return defaultDays
}

// GetBatchSize returns the batch size for purge operations
// Falls back to defaultSize if not found
func (c *Config) GetBatchSize(entityName string, defaultSize int) int {
	if c.Purge.Settings == nil {
		return defaultSize
	}

	entitySettings, ok := c.Purge.Settings[entityName]
	if !ok {
		return defaultSize
	}

	if size, ok := entitySettings["batch_size"].(int); ok {
		return size
	}

	return defaultSize
}

// GetFeature returns a feature flag value
// Falls back to defaultValue if not found
func (c *Config) GetFeature(featureName string, defaultValue bool) bool {
	if c.Settings == nil {
		return defaultValue
	}

	features, ok := c.Settings["features"].(map[string]interface{})
	if !ok {
		return defaultValue
	}

	if value, ok := features[featureName].(bool); ok {
		return value
	}

	return defaultValue
}

// GetNestedSetting retrieves a nested setting from the config
// Returns nil if not found
func (c *Config) GetNestedSetting(keys ...string) interface{} {
	if len(keys) == 0 {
		return nil
	}

	var current interface{} = c.Settings

	for _, key := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}

		value, exists := m[key]
		if !exists {
			return nil
		}

		current = value
	}

	return current
}
