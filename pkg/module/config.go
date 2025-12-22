package module

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the module configuration file
type Config struct {
	Modules ModulesConfig `yaml:"modules"`
}

// ModulesConfig contains module settings
type ModulesConfig struct {
	Enabled []string               `yaml:"enabled"` // List of enabled module names
	Config  map[string]ModuleParam `yaml:"config"`  // Per-module configuration
}

// ModuleParam contains module-specific parameters
type ModuleParam struct {
	Version    string                 `yaml:"version"`     // Required module version
	LicenseKey string                 `yaml:"license_key"` // License key (for commercial modules)
	Settings   map[string]interface{} `yaml:"settings"`    // Module-specific settings
}

// LoadConfig loads module configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves module configuration to YAML file
func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetModuleConfig returns configuration for a specific module
func (c *Config) GetModuleConfig(moduleName string) (ModuleParam, bool) {
	param, exists := c.Modules.Config[moduleName]
	return param, exists
}

// IsEnabled checks if a module is enabled
func (c *Config) IsEnabled(moduleName string) bool {
	for _, name := range c.Modules.Enabled {
		if name == moduleName {
			return true
		}
	}
	return false
}
