package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
)

// TestDB manages test database lifecycle
type TestDB struct {
	DB     *sqlx.DB
	Config *config.DatabaseConfig
}

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *TestDB {
	cfg := &config.DatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     5433,
		User:     getEnv("TEST_DB_USER", "system"),
		Password: getEnv("TEST_DB_PASSWORD", "passw0rd"),
		DBName:   getEnv("TEST_DB_NAME", "promenade_test"),
		SSLMode:  "disable",
	}

	db, err := database.NewPostgresConnection(cfg)
	require.NoError(t, err, "Failed to connect to test database")

	return &TestDB{
		DB:     db,
		Config: cfg,
	}
}

// Close closes the test database connection
func (tdb *TestDB) Close() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// CleanupTables truncates all tables (except migrations)
func (tdb *TestDB) CleanupTables(t *testing.T) {
	tables := []string{
		"sessions",
		"login_attempts",
		"email_verification_tokens",
		"password_reset_tokens",
		"users",
	}

	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(t, err, "Failed to truncate table: "+table)
	}
}

// RunInTransaction runs a function in a transaction and rolls back
func (tdb *TestDB) RunInTransaction(t *testing.T, fn func(tx *sqlx.Tx)) {
	tx, err := tdb.DB.Beginx()
	require.NoError(t, err, "Failed to begin transaction")

	defer tx.Rollback()

	fn(tx)
}

// WaitForDB waits for database to be ready
func (tdb *TestDB) WaitForDB(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for database")
		case <-ticker.C:
			if err := tdb.DB.Ping(); err == nil {
				return nil
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := testConfig[key]; value != "" {
		return value
	}
	return defaultValue
}

var testConfig = map[string]string{
	"TEST_DB_HOST":     "localhost",
	"TEST_DB_PORT":     "5433",
	"TEST_DB_USER":     "system",
	"TEST_DB_PASSWORD": "passw0rd",
	"TEST_DB_NAME":     "promenade_test",
}
