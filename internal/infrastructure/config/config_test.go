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

func clearEnv() {
	envVars := []string{
		"SERVER_PORT", "ENVIRONMENT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
		"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
	}
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
}
