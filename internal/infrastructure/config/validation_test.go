package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &AppConfig{
		App: AppSection{
			Environment: "development",
		},
		Server: ServerSection{
			Port: 8080,
		},
		Database: DatabaseSection{
			Host:     "localhost",
			Database: "testdb",
		},
		JWT: JWTSection{
			Secret: "this-is-a-long-enough-secret-key-for-testing-32-chars",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_JWTSecretValidation(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		secret      string
		wantErr     bool
		errContains string
	}{
		{
			name:        "production with valid long secret",
			environment: "production",
			secret:      "this-is-a-very-secure-production-secret-key-with-more-than-32-chars",
			wantErr:     false,
		},
		{
			name:        "production with short secret",
			environment: "production",
			secret:      "short",
			wantErr:     true,
			errContains: "must be at least 32 characters",
		},
		{
			name:        "production with default dev secret",
			environment: "production",
			secret:      "dev-secret-key-change-in-production",
			wantErr:     true,
			errContains: "default JWT secret cannot be used in production",
		},
		{
			name:        "production with empty secret",
			environment: "production",
			secret:      "",
			wantErr:     true,
			errContains: "JWT secret is required",
		},
		{
			name:        "development with short secret (allowed)",
			environment: "development",
			secret:      "dev-secret",
			wantErr:     false,
		},
		{
			name:        "development with default secret (allowed)",
			environment: "development",
			secret:      "dev-secret-key-change-in-production",
			wantErr:     false,
		},
		{
			name:        "prod environment alias",
			environment: "prod",
			secret:      "short",
			wantErr:     true,
			errContains: "must be at least 32 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AppConfig{
				App: AppSection{
					Environment: tt.environment,
				},
				Server: ServerSection{
					Port: 8080,
				},
				Database: DatabaseSection{
					Host:     "localhost",
					Database: "testdb",
				},
				JWT: JWTSection{
					Secret: tt.secret,
				},
			}

			err := cfg.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_DatabaseValidation(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		database    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid database config",
			host:     "localhost",
			database: "testdb",
			wantErr:  false,
		},
		{
			name:        "missing host",
			host:        "",
			database:    "testdb",
			wantErr:     true,
			errContains: "database host is required",
		},
		{
			name:        "missing database name",
			host:        "localhost",
			database:    "",
			wantErr:     true,
			errContains: "database name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AppConfig{
				App: AppSection{
					Environment: "development",
				},
				Server: ServerSection{
					Port: 8080,
				},
				Database: DatabaseSection{
					Host:     tt.host,
					Database: tt.database,
				},
				JWT: JWTSection{
					Secret: "valid-secret-key-for-development",
				},
			}

			err := cfg.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ServerValidation(t *testing.T) {
	cfg := &AppConfig{
		App: AppSection{
			Environment: "development",
		},
		Server: ServerSection{
			Port: 0, // Invalid
		},
		Database: DatabaseSection{
			Host:     "localhost",
			Database: "testdb",
		},
		JWT: JWTSection{
			Secret: "valid-secret-key",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server port is required")
}
