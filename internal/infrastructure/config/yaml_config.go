package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// AppConfig represents core application configuration
type AppConfig struct {
	App       AppSection       `yaml:"app"`
	Server    ServerSection    `yaml:"server"`
	Database  DatabaseSection  `yaml:"database"`
	JWT       JWTSection       `yaml:"jwt"`
	Logging   LoggingSection   `yaml:"logging"`
	CORS      CORSSection      `yaml:"cors"`
	Bus       BusSection       `yaml:"bus"`
	RateLimit RateLimitSection `yaml:"rate_limit"`
	Email     EmailSection     `yaml:"email"`
	Purge     PurgeSection     `yaml:"purge"`
	Modules   ModulesSection   `yaml:"modules"`
}

type AppSection struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
	Version     string `yaml:"version"`
	URL         string `yaml:"url"`
}

type ServerSection struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseSection struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

type JWTSection struct {
	Secret               string        `yaml:"secret"`
	AccessTokenDuration  time.Duration `yaml:"access_token_duration"`
	RefreshTokenDuration time.Duration `yaml:"refresh_token_duration"`
	Issuer               string        `yaml:"issuer"`
}

type LoggingSection struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

type CORSSection struct {
	AllowedOrigins   []string      `yaml:"allowed_origins"`
	AllowedMethods   []string      `yaml:"allowed_methods"`
	AllowedHeaders   []string      `yaml:"allowed_headers"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
}

type BusSection struct {
	Adapter          string        `yaml:"adapter"`
	WorkerPoolSize   int           `yaml:"worker_pool_size"`
	BufferSize       int           `yaml:"buffer_size"`
	RetryAttempts    int           `yaml:"retry_attempts"`
	RetryDelay       time.Duration `yaml:"retry_delay"`
	RetryMaxDelay    time.Duration `yaml:"retry_max_delay"`
	RetryMultiplier  float64       `yaml:"retry_multiplier"`
	Redis            RedisSection  `yaml:"redis"`
}

type RedisSection struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Password   string `yaml:"password"`
	DB         int    `yaml:"db"`
	MaxRetries int    `yaml:"max_retries"`
	PoolSize   int    `yaml:"pool_size"`
}

type RateLimitSection struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

type EmailSection struct {
	Enabled     bool   `yaml:"enabled"`
	SMTPHost    string `yaml:"smtp_host"`
	SMTPPort    int    `yaml:"smtp_port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"from_address"`
	FromName    string `yaml:"from_name"`
	AppURL      string `yaml:"app_url"`
	AppName     string `yaml:"app_name"`
}

// PurgeSection represents purge infrastructure settings
type PurgeSection struct {
	Enabled   bool   `yaml:"enabled"`
	Schedule  string `yaml:"schedule"`
	DryRun    bool   `yaml:"dry_run"`
	BatchSize int    `yaml:"batch_size"`
}

// ModulesSection represents modules configuration
type ModulesSection struct {
	Enabled []string `yaml:"enabled"`
}

// ModuleConfig represents a module's configuration
type ModuleConfig struct {
	Module      ModuleSection          `yaml:"module"`
	Settings    map[string]any `yaml:",inline"` // Module-specific settings
	Purge       *ModulePurgeSection    `yaml:"purge,omitempty"`
	Permissions []PermissionSection    `yaml:"permissions,omitempty"`
	Features    map[string]bool        `yaml:"features,omitempty"`
}

type ModuleSection struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
	Version string `yaml:"version"`
}

type ModulePurgeSection struct {
	Enabled  bool                   `yaml:"enabled"`
	Settings map[string]any `yaml:",inline"` // Entity-specific purge settings
}

type PermissionSection struct {
	Resource string   `yaml:"resource"`
	Actions  []string `yaml:"actions"`
}

// LoadAppConfig loads the core application config from YAML
// Supports environment-specific configs: app.dev.yaml, app.test.yaml, app.prod.yaml
func LoadAppConfig(configPath string) (*AppConfig, error) {
	// If no path specified, auto-detect based on environment
	if configPath == "" {
		env := os.Getenv("ENVIRONMENT")
		if env == "" {
			env = "development"
		}

		// Try environment-specific config first
		envConfigPath := fmt.Sprintf("config/app.%s.yaml", getEnvSuffix(env))
		if _, err := os.Stat(envConfigPath); err == nil {
			configPath = envConfigPath
		} else {
			// Fall back to generic app.yaml
			configPath = "config/app.yaml"
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	return &config, nil
}

// getEnvSuffix returns the file suffix for environment
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

// NOTE: LoadModuleConfig and LoadAllModuleConfigs have been removed.
// Modules are now fully autonomous and load their own configs from their own directories
// using pkg/module/config.Load(). Core config (yaml_config.go) only handles app.*.yaml files.

// Load is the main entry point for loading application configuration from YAML.
func Load() (*AppConfig, error) {
	return LoadAppConfig("")
}

// applyEnvOverrides allows environment variables to override config values
func applyEnvOverrides(cfg *AppConfig) {
	// Database overrides
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Database.Port)
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Database = v
	}

	// Server overrides
	if v := os.Getenv("SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}

	// JWT overrides
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}

	// Environment override
	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.App.Environment = v
	}

	// Bus adapter override
	if v := os.Getenv("BUS_ADAPTER"); v != "" {
		cfg.Bus.Adapter = v
	}
}
