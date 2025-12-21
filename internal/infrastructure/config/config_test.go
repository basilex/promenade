package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	clearEnv()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "development", cfg.Server.Environment)
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)

	assert.Equal(t, "your-secret-key-change-in-production", cfg.JWT.Secret)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenTTL)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTokenTTL)

	// Bus config defaults
	assert.Equal(t, 10, cfg.Bus.WorkerPoolSize)
	assert.Equal(t, 1000, cfg.Bus.BufferSize)
	assert.Equal(t, 3, cfg.Bus.RetryAttempts)
	assert.Equal(t, 1*time.Second, cfg.Bus.RetryDelay)
	assert.Equal(t, 5*time.Second, cfg.Bus.RetryMaxDelay)
	assert.Equal(t, 2.0, cfg.Bus.RetryMultiplier)
}

func TestValidate_ProductionWithDefaultSecret(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Environment: "production",
		},
		JWT: JWTConfig{
			Secret: "your-secret-key-change-in-production",
		},
		Database: DatabaseConfig{
			SSLMode: "require",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET must be changed")
}

func TestValidate_ProductionWithoutSSL(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Environment: "production",
		},
		JWT: JWTConfig{
			Secret: "my-production-secret",
		},
		Database: DatabaseConfig{
			SSLMode: "disable",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSL should be enabled")
}

func TestValidate_ValidProductionConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Environment: "production",
		},
		JWT: JWTConfig{
			Secret: "my-secure-production-secret-key",
		},
		Database: DatabaseConfig{
			SSLMode: "require",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestLoad_BusConfigFromEnv(t *testing.T) {
	clearEnv()
	os.Setenv("BUS_WORKER_POOL_SIZE", "20")
	os.Setenv("BUS_BUFFER_SIZE", "2000")
	os.Setenv("BUS_RETRY_ATTEMPTS", "5")
	os.Setenv("BUS_RETRY_DELAY", "2s")
	os.Setenv("BUS_RETRY_MAX_DELAY", "10s")
	os.Setenv("BUS_RETRY_MULTIPLIER", "3.5")
	defer clearEnv()

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 20, cfg.Bus.WorkerPoolSize)
	assert.Equal(t, 2000, cfg.Bus.BufferSize)
	assert.Equal(t, 5, cfg.Bus.RetryAttempts)
	assert.Equal(t, 2*time.Second, cfg.Bus.RetryDelay)
	assert.Equal(t, 10*time.Second, cfg.Bus.RetryMaxDelay)
	assert.Equal(t, 3.5, cfg.Bus.RetryMultiplier)
}

func TestGetEnvAsFloat(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue float64
		expected     float64
	}{
		{
			name:         "valid float",
			envValue:     "2.5",
			defaultValue: 1.0,
			expected:     2.5,
		},
		{
			name:         "integer as float",
			envValue:     "3",
			defaultValue: 1.0,
			expected:     3.0,
		},
		{
			name:         "invalid float returns default",
			envValue:     "invalid",
			defaultValue: 2.0,
			expected:     2.0,
		},
		{
			name:         "empty string returns default",
			envValue:     "",
			defaultValue: 1.5,
			expected:     1.5,
		},
		{
			name:         "negative float",
			envValue:     "-1.5",
			defaultValue: 1.0,
			expected:     -1.5,
		},
		{
			name:         "zero",
			envValue:     "0",
			defaultValue: 1.0,
			expected:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_FLOAT_VAR"
			defer os.Unsetenv(key)

			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			}

			result := getEnvAsFloat(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBusConfig_EdgeCases(t *testing.T) {
	t.Run("very large multiplier", func(t *testing.T) {
		clearEnv()
		os.Setenv("BUS_RETRY_MULTIPLIER", "100.0")
		defer clearEnv()

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 100.0, cfg.Bus.RetryMultiplier)
	})

	t.Run("fractional multiplier", func(t *testing.T) {
		clearEnv()
		os.Setenv("BUS_RETRY_MULTIPLIER", "1.5")
		defer clearEnv()

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 1.5, cfg.Bus.RetryMultiplier)
	})

	t.Run("very small max delay", func(t *testing.T) {
		clearEnv()
		os.Setenv("BUS_RETRY_MAX_DELAY", "100ms")
		defer clearEnv()

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 100*time.Millisecond, cfg.Bus.RetryMaxDelay)
	})

	t.Run("very large max delay", func(t *testing.T) {
		clearEnv()
		os.Setenv("BUS_RETRY_MAX_DELAY", "1h")
		defer clearEnv()

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 1*time.Hour, cfg.Bus.RetryMaxDelay)
	})
}

func clearEnv() {
	envVars := []string{
		"SERVER_PORT", "ENVIRONMENT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
		"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"BUS_WORKER_POOL_SIZE", "BUS_BUFFER_SIZE", "BUS_RETRY_ATTEMPTS",
		"BUS_RETRY_DELAY", "BUS_RETRY_MAX_DELAY", "BUS_RETRY_MULTIPLIER",
	}
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
}
