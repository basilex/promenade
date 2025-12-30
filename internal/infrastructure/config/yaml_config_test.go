package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAppConfig_ValidYAML(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "app.yaml")

	yamlContent := `
app:
  name: "TestApp"
  environment: "test"
  debug: true
  version: "1.0.0"
  url: "https://test.example.com"

server:
  host: "0.0.0.0"
  port: 9000
  read_timeout: 20s
  write_timeout: 30s
  shutdown_timeout: 15s

database:
  postgres:
    host: "testdb.local"
    port: 5433
    user: "testuser"
    password: "testpass"
    database: "testdb"
    sslmode: "require"
    max_open_conns: 30
    max_idle_conns: 10
    conn_max_lifetime: 10m
    conn_max_idle_time: 5m
  
  redis:
    addr: "redis.test.local:6380"
    password: "redis-test-pass"
    pool_size: 15
    max_retries: 4
    databases:
      revocation: 0
      bus: 2
      cache: 3
      sessions: 4

jwt:
  secret: "test-jwt-secret"
  access_token_duration: 20m
  refresh_token_duration: 240h
  issuer: "test-issuer"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "logs/test.log"

cors:
  allowed_origins:
    - "https://test.example.com"
    - "http://localhost:3001"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
  allowed_headers:
    - "Content-Type"
    - "Authorization"
  allow_credentials: true
  max_age: 6h

bus:
  adapter: "redis"
  worker_pool_size: 6
  buffer_size: 150
  retry_attempts: 5
  retry_delay: 3s
  retry_max_delay: 30s
  retry_multiplier: 2.0

rate_limit:
  enabled: true
  requests_per_second: 50
  burst: 100

email:
  enabled: true
  smtp_host: "smtp.test.local"
  smtp_port: 587
  username: "test@example.com"
  password: "smtp-pass"
  from_address: "noreply@test.com"
  from_name: "Test Mailer"

purge:
  enabled: true
  schedule: "0 3 * * *"
  dry_run: false
  batch_size: 500
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := LoadAppConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify App section
	assert.Equal(t, "TestApp", cfg.App.Name)
	assert.Equal(t, "test", cfg.App.Environment)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "https://test.example.com", cfg.App.URL)

	// Verify Server section
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, 20*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 15*time.Second, cfg.Server.ShutdownTimeout)

	// Verify Database section
	assert.Equal(t, "testdb.local", cfg.Database.Postgres.Host)
	assert.Equal(t, 5433, cfg.Database.Postgres.Port)
	assert.Equal(t, "testuser", cfg.Database.Postgres.User)
	assert.Equal(t, "testpass", cfg.Database.Postgres.Password)
	assert.Equal(t, "testdb", cfg.Database.Postgres.Database)
	assert.Equal(t, "require", cfg.Database.Postgres.SSLMode)
	assert.Equal(t, 30, cfg.Database.Postgres.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.Postgres.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, cfg.Database.Postgres.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, cfg.Database.Postgres.ConnMaxIdleTime)

	// Verify JWT section
	assert.Equal(t, "test-jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, 20*time.Minute, cfg.JWT.AccessTokenDuration)
	assert.Equal(t, 240*time.Hour, cfg.JWT.RefreshTokenDuration)
	assert.Equal(t, "test-issuer", cfg.JWT.Issuer)

	// Verify Logging section
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Equal(t, "logs/test.log", cfg.Logging.FilePath)

	// Verify CORS section
	assert.Len(t, cfg.CORS.AllowedOrigins, 2)
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://test.example.com")
	assert.Len(t, cfg.CORS.AllowedMethods, 3)
	assert.Contains(t, cfg.CORS.AllowedMethods, "POST")
	assert.Len(t, cfg.CORS.AllowedHeaders, 2)
	assert.True(t, cfg.CORS.AllowCredentials)
	assert.Equal(t, 6*time.Hour, cfg.CORS.MaxAge)

	// Verify Bus section
	assert.Equal(t, "redis", cfg.Bus.Adapter)
	assert.Equal(t, 6, cfg.Bus.WorkerPoolSize)
	assert.Equal(t, 150, cfg.Bus.BufferSize)
	assert.Equal(t, 5, cfg.Bus.RetryAttempts)
	assert.Equal(t, 3*time.Second, cfg.Bus.RetryDelay)
	assert.Equal(t, 30*time.Second, cfg.Bus.RetryMaxDelay)
	assert.Equal(t, 2.0, cfg.Bus.RetryMultiplier)

	// Verify Redis section (moved to database.redis)
	assert.Equal(t, "redis.test.local:6380", cfg.Database.Redis.Addr)
	assert.Equal(t, "redis-test-pass", cfg.Database.Redis.Password)
	assert.Equal(t, 2, cfg.Database.Redis.Databases.Bus)
	assert.Equal(t, 4, cfg.Database.Redis.MaxRetries)
	assert.Equal(t, 15, cfg.Database.Redis.PoolSize)

	// Verify RateLimit section
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 50, cfg.RateLimit.RequestsPerSecond)
	assert.Equal(t, 100, cfg.RateLimit.Burst)

	// Verify Email section
	assert.True(t, cfg.Email.Enabled)
	assert.Equal(t, "smtp.test.local", cfg.Email.SMTPHost)
	assert.Equal(t, 587, cfg.Email.SMTPPort)
	assert.Equal(t, "test@example.com", cfg.Email.Username)
	assert.Equal(t, "smtp-pass", cfg.Email.Password)
	assert.Equal(t, "noreply@test.com", cfg.Email.FromAddress)
	assert.Equal(t, "Test Mailer", cfg.Email.FromName)

	// Verify Purge section
	assert.True(t, cfg.Purge.Enabled)
	assert.Equal(t, "0 3 * * *", cfg.Purge.Schedule)
	assert.False(t, cfg.Purge.DryRun)
	assert.Equal(t, 500, cfg.Purge.BatchSize)
}

func TestLoadAppConfig_InvalidPath(t *testing.T) {
	cfg, err := LoadAppConfig("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadAppConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	invalidYAML := `
app:
  name: "Test
  invalid yaml structure
    no proper indentation
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	cfg, err := LoadAppConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := &AppConfig{
		Database: DatabasesSection{
			Postgres: PostgresSection{
				Host:     "original-db",
				Port:     5432,
				User:     "original-user",
				Password: "original-pass",
				Database: "original-db-name",
			},
		},
		Server: ServerSection{
			Port: 8080,
		},
		JWT: JWTSection{
			Secret: "original-secret",
		},
		App: AppSection{
			Environment: "development",
		},
		Bus: BusSection{
			Adapter: "memory",
		},
	}

	// Set environment variables
	_ = os.Setenv("DB_HOST", "env-db-host")
	_ = os.Setenv("DB_PORT", "5433")
	_ = os.Setenv("DB_USER", "env-user")
	_ = os.Setenv("DB_PASSWORD", "env-pass")
	_ = os.Setenv("DB_NAME", "env-db-name")
	_ = os.Setenv("SERVER_PORT", "9000")
	_ = os.Setenv("JWT_SECRET", "env-jwt-secret")
	_ = os.Setenv("ENVIRONMENT", "production")
	_ = os.Setenv("BUS_ADAPTER", "redis")

	defer func() {
		_ = os.Unsetenv("DB_HOST")
		_ = os.Unsetenv("DB_PORT")
		_ = os.Unsetenv("DB_USER")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_NAME")
		_ = os.Unsetenv("SERVER_PORT")
		_ = os.Unsetenv("JWT_SECRET")
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("BUS_ADAPTER")
	}()

	applyEnvOverrides(cfg)

	assert.Equal(t, "env-db-host", cfg.Database.Postgres.Host)
	assert.Equal(t, 5433, cfg.Database.Postgres.Port)
	assert.Equal(t, "env-user", cfg.Database.Postgres.User)
	assert.Equal(t, "env-pass", cfg.Database.Postgres.Password)
	assert.Equal(t, "env-db-name", cfg.Database.Postgres.Database)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "env-jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "redis", cfg.Bus.Adapter)
}

func TestGetEnvSuffix(t *testing.T) {
	tests := []struct {
		env      string
		expected string
	}{
		{"development", "dev"},
		{"test", "test"},
		{"testing", "test"},
		{"production", "prod"},
		{"prod", "prod"},
		{"staging", "dev"}, // default
		{"", "dev"},        // default
		{"unknown", "dev"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			result := getEnvSuffix(tt.env)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModuleConfig_Structure(t *testing.T) {
	moduleConfig := ModuleConfig{
		Module: ModuleSection{
			Name:    "posts",
			Enabled: true,
			Version: "2.0.0",
		},
		Settings: map[string]any{
			"max_post_length": 5000,
			"allow_comments":  true,
		},
		Purge: &ModulePurgeSection{
			Enabled: true,
			Settings: map[string]any{
				"retention_days": 90,
			},
		},
		Permissions: []PermissionSection{
			{
				Resource: "posts",
				Actions:  []string{"create", "read", "update", "delete"},
			},
			{
				Resource: "comments",
				Actions:  []string{"create", "read", "delete"},
			},
		},
		Features: map[string]bool{
			"enable_likes":    true,
			"enable_comments": true,
			"enable_shares":   false,
		},
	}

	assert.Equal(t, "posts", moduleConfig.Module.Name)
	assert.True(t, moduleConfig.Module.Enabled)
	assert.Equal(t, "2.0.0", moduleConfig.Module.Version)
	assert.Equal(t, 5000, moduleConfig.Settings["max_post_length"])
	assert.True(t, moduleConfig.Settings["allow_comments"].(bool))
	assert.NotNil(t, moduleConfig.Purge)
	assert.True(t, moduleConfig.Purge.Enabled)
	assert.Equal(t, 90, moduleConfig.Purge.Settings["retention_days"])
	assert.Len(t, moduleConfig.Permissions, 2)
	assert.Equal(t, "posts", moduleConfig.Permissions[0].Resource)
	assert.Len(t, moduleConfig.Permissions[0].Actions, 4)
	assert.True(t, moduleConfig.Features["enable_likes"])
	assert.False(t, moduleConfig.Features["enable_shares"])
}

func TestCorePurgeSection(t *testing.T) {
	purge := PurgeSection{
		Enabled:   true,
		Schedule:  "0 2 * * *",
		DryRun:    false,
		BatchSize: 1000,
	}

	assert.True(t, purge.Enabled)
	assert.Equal(t, "0 2 * * *", purge.Schedule)
	assert.False(t, purge.DryRun)
	assert.Equal(t, 1000, purge.BatchSize)
}

func TestPermissionSection(t *testing.T) {
	permission := PermissionSection{
		Resource: "users",
		Actions:  []string{"create", "read", "update", "delete", "suspend", "ban"},
	}

	assert.Equal(t, "users", permission.Resource)
	assert.Len(t, permission.Actions, 6)
	assert.Contains(t, permission.Actions, "suspend")
	assert.Contains(t, permission.Actions, "ban")
}

func TestLoadAppConfig_MinimalYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "minimal.yaml")

	minimalYAML := `
app:
  name: "MinimalApp"
  environment: "dev"

server:
  port: 8080

database:
  postgres:
    host: "localhost"
    port: 5432
    user: "user"
    password: "pass"
    database: "db"

jwt:
  secret: "secret123"
`

	err := os.WriteFile(configPath, []byte(minimalYAML), 0644)
	require.NoError(t, err)

	cfg, err := LoadAppConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "MinimalApp", cfg.App.Name)
	assert.Equal(t, "dev", cfg.App.Environment)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Postgres.Host)
	assert.Equal(t, "secret123", cfg.JWT.Secret)
}
