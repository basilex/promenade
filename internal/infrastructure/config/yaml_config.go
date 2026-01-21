package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/basilex/promenade/pkg/scheduler"
)

// AppConfig represents core application configuration
type AppConfig struct {
	App       AppSection       `yaml:"app"`
	Server    ServerSection    `yaml:"server"`
	Database  DatabasesSection `yaml:"database"` // Centralized databases (Postgres, Redis, etc.)
	JWT       JWTSection       `yaml:"jwt"`
	Logging   LoggingSection   `yaml:"logging"`
	CORS      CORSSection      `yaml:"cors"`
	Bus       BusSection       `yaml:"bus"`
	Cache     CacheSection     `yaml:"cache"`
	Scheduler scheduler.Config `yaml:"scheduler"`
	RateLimit RateLimitSection `yaml:"rate_limit"`
	Email     EmailSection     `yaml:"email"`
	Purge     PurgeSection     `yaml:"purge"`
	Modules   ModulesSection   `yaml:"modules"`
	Fiscal    FiscalSection    `yaml:"fiscal"`
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

// DatabasesSection holds all database configurations
type DatabasesSection struct {
	Driver   string          `yaml:"driver"`   // Database driver: currently only "postgres" supported
	Postgres PostgresSection `yaml:"postgres"` // PostgreSQL config (required)
	SQLite   SQLiteSection   `yaml:"sqlite"`   // SQLite config (deprecated, for legacy support only)
	Redis    RedisSection    `yaml:"redis"`    // Redis for cache/sessions/bus
}

// PostgresSection holds PostgreSQL configuration
type PostgresSection struct {
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

// SQLiteSection holds SQLite configuration
type SQLiteSection struct {
	Path         string `yaml:"path"`           // Database file path (e.g., "./data/promenade.db")
	Mode         string `yaml:"mode"`           // rwc (read-write-create), ro (read-only), memory
	Cache        string `yaml:"cache"`          // shared, private
	MaxOpenConns int    `yaml:"max_open_conns"` // Default: 1 (SQLite limitation)
	MaxIdleConns int    `yaml:"max_idle_conns"` // Default: 1
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
	Adapter         string        `yaml:"adapter"`
	WorkerPoolSize  int           `yaml:"worker_pool_size"`
	BufferSize      int           `yaml:"buffer_size"`
	RetryAttempts   int           `yaml:"retry_attempts"`
	RetryDelay      time.Duration `yaml:"retry_delay"`
	RetryMaxDelay   time.Duration `yaml:"retry_max_delay"`
	RetryMultiplier float64       `yaml:"retry_multiplier"`
	// Redis config moved to database.redis with DB selection via databases.bus
}

// RedisSection holds unified Redis configuration for all services
type RedisSection struct {
	Addr       string         `yaml:"addr"`        // Redis address (host:port)
	Password   string         `yaml:"password"`    // Redis password
	PoolSize   int            `yaml:"pool_size"`   // Connection pool size
	MaxRetries int            `yaml:"max_retries"` // Max retry attempts
	Databases  RedisDatabases `yaml:"databases"`   // Logical DB separation
}

// RedisDatabases defines logical database numbers for different services
type RedisDatabases struct {
	Revocation int `yaml:"revocation"` // JWT token blacklist
	Bus        int `yaml:"bus"`        // Event Bus (when redis adapter)
	Cache      int `yaml:"cache"`      // Cache layer
	Sessions   int `yaml:"sessions"`   // User sessions
}

type RateLimitSection struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerSecond int  `yaml:"requests_per_second"`
	Burst             int  `yaml:"burst"`
}

// CacheSection holds cache configuration
type CacheSection struct {
	Enabled    bool     `yaml:"enabled"`     // Enable/disable caching
	Adapter    string   `yaml:"adapter"`     // redis or noop
	Prefix     string   `yaml:"prefix"`      // Key prefix for namespace isolation
	DefaultTTL string   `yaml:"default_ttl"` // Default TTL (e.g., "5m")
	TTL        CacheTTL `yaml:"ttl"`         // TTL per resource type
}

// CacheTTL holds TTL configuration for different resource types
type CacheTTL struct {
	Countries   string `yaml:"countries"`    // e.g., "1h"
	Currencies  string `yaml:"currencies"`   // e.g., "1h"
	Languages   string `yaml:"languages"`    // e.g., "1h"
	Timezones   string `yaml:"timezones"`    // e.g., "1h"
	UserProfile string `yaml:"user_profile"` // e.g., "15m"
	Customer    string `yaml:"customer"`     // e.g., "10m"
	Session     string `yaml:"session"`      // e.g., "30m"
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

// FiscalSection holds fiscal integrations configuration
type FiscalSection struct {
	Checkbox       CheckboxSection `yaml:"checkbox"`
	PDFOutputDir   string          `yaml:"pdf_output_dir"`
	RetryCron      string          `yaml:"retry_cron"`
	ShiftOpenCron  string          `yaml:"shift_open_cron"`
	ShiftCloseCron string          `yaml:"shift_close_cron"`
}

// CheckboxSection holds Checkbox API configuration
type CheckboxSection struct {
	APIKey  string        `yaml:"api_key"`
	Sandbox bool          `yaml:"sandbox"`
	Timeout time.Duration `yaml:"timeout"`
}

// ModuleConfig represents a module's configuration
type ModuleConfig struct {
	Module      ModuleSection       `yaml:"module"`
	Settings    map[string]any      `yaml:",inline"` // Module-specific settings
	Purge       *ModulePurgeSection `yaml:"purge,omitempty"`
	Permissions []PermissionSection `yaml:"permissions,omitempty"`
	Features    map[string]bool     `yaml:"features,omitempty"`
}

type ModuleSection struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
	Version string `yaml:"version"`
}

type ModulePurgeSection struct {
	Enabled  bool           `yaml:"enabled"`
	Settings map[string]any `yaml:",inline"` // Entity-specific purge settings
}

type PermissionSection struct {
	Resource string   `yaml:"resource"`
	Actions  []string `yaml:"actions"`
}

// LoadAppConfig loads the core application config from YAML
// Supports driver-environment configs: app.{driver}-{env}.yaml
// Examples: app.postgres-dev.yaml, app.sqlite-test.yaml, app.mysql-prod.yaml
func LoadAppConfig(configPath string) (*AppConfig, error) {
	// If no path specified, auto-detect based on driver and environment
	if configPath == "" {
		driver := os.Getenv("DATABASE_DRIVER")
		if driver == "" {
			driver = "postgres" // Default to PostgreSQL
		}

		env := os.Getenv("ENVIRONMENT")
		if env == "" {
			env = "development"
		}

		// Build config path: app.{driver}-{env}.yaml
		envSuffix := getEnvSuffix(env)
		envConfigPath := fmt.Sprintf("config/app.%s-%s.yaml", driver, envSuffix)

		if _, err := os.Stat(envConfigPath); err == nil {
			configPath = envConfigPath
		} else {
			// Fall back to generic app.yaml (backward compatibility)
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

// Validate validates the configuration based on environment
func (cfg *AppConfig) Validate() error {
	// JWT secret validation (critical for security)
	if err := cfg.validateJWTSecret(); err != nil {
		return err
	}

	// Database validation (driver-specific)
	switch cfg.Database.Driver {
	case "postgres":
		if cfg.Database.Postgres.Host == "" {
			return fmt.Errorf("database.postgres.host is required")
		}
		if cfg.Database.Postgres.Database == "" {
			return fmt.Errorf("database.postgres.database is required")
		}
	case "sqlite":
		if cfg.Database.SQLite.Path == "" {
			return fmt.Errorf("database.sqlite.path is required")
		}
	default:
		return fmt.Errorf("database.driver must be 'postgres' or 'sqlite'")
	}

	// Server validation
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}

	return nil
}

// validateJWTSecret validates JWT secret key based on environment
func (cfg *AppConfig) validateJWTSecret() error {
	const minSecretLength = 32
	const defaultDevSecret = "dev-secret-key-change-in-production"

	secret := cfg.JWT.Secret

	// Check if secret is empty
	if secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	// Production environment checks
	if cfg.App.Environment == "production" || cfg.App.Environment == "prod" {
		// Check minimum length
		if len(secret) < minSecretLength {
			return fmt.Errorf("JWT secret must be at least %d characters in production (current: %d)", minSecretLength, len(secret))
		}

		// Check for default dev secret
		if secret == defaultDevSecret {
			return fmt.Errorf("default JWT secret cannot be used in production. Please set JWT_SECRET environment variable")
		}
	}

	// Note: Short secrets are allowed in development for convenience
	// In production, they are blocked above. No action needed here.

	return nil
}

// applyEnvOverrides allows environment variables to override config values
func applyEnvOverrides(cfg *AppConfig) {
	// PostgreSQL overrides
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Postgres.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Postgres.Port = port
		}
		// If parsing fails, keep default port from config
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.Postgres.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Postgres.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Postgres.Database = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Database.Redis.Addr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Database.Redis.Password = v
	}

	// Server overrides
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
		// If parsing fails, keep default port from config
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

	// Fiscal Checkbox overrides
	if v := os.Getenv("CHECKBOX_API_KEY"); v != "" {
		cfg.Fiscal.Checkbox.APIKey = v
	}
	if v := os.Getenv("CHECKBOX_SANDBOX"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			cfg.Fiscal.Checkbox.Sandbox = parsed
		}
	}
	if v := os.Getenv("CHECKBOX_TIMEOUT"); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			cfg.Fiscal.Checkbox.Timeout = parsed
		}
	}
	if v := os.Getenv("FISCAL_PDF_OUTPUT_DIR"); v != "" {
		cfg.Fiscal.PDFOutputDir = v
	}
}
