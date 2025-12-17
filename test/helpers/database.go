package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
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
		_ = tdb.DB.Close()
	}
}

// CleanupTables truncates all tables (except migrations)
func (tdb *TestDB) CleanupTables(t *testing.T) {
	tables := []string{
		"user_profiles",
		"user_contacts",
		"user_sessions",
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

	defer func() { _ = tx.Rollback() }()

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

// CreateTestUser creates and inserts a test user into the database
func CreateTestUser(t *testing.T, db *sqlx.DB, email, name string) *entity.User {
	user := UserFixture(func(u *entity.User) {
		u.Email = email
		u.Name = name
	})

	query := `
		INSERT INTO users (id, email, name, password, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query, user.ID, user.Email, user.Name, user.Password, user.Status, user.CreatedAt, user.UpdatedAt)
	require.NoError(t, err, "Failed to create test user")

	return user
}

// CreateTestProfile creates and inserts a test user profile into the database
func CreateTestProfile(t *testing.T, db *sqlx.DB, userID uuidv7.UUID, nickname string) *entity.UserProfile {
	profile := UserProfileFixture(userID, nickname)

	// Marshal JSONB fields
	err := profile.MarshalSocialLinks()
	require.NoError(t, err, "Failed to marshal social links")
	err = profile.MarshalPreferences()
	require.NoError(t, err, "Failed to marshal preferences")

	query := `
		INSERT INTO user_profiles (
			id, user_id, display_name, nickname, bio, timezone, locale,
			is_public, is_verified, show_email, show_location, show_birthday,
			social_links, preferences, profile_views_count, followers_count, following_count,
			is_banned, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20
		)
	`
	_, err = db.Exec(query,
		profile.ID, profile.UserID, profile.DisplayName, profile.Nickname, profile.Bio,
		profile.Timezone, profile.Locale,
		profile.IsPublic, profile.IsVerified, profile.ShowEmail, profile.ShowLocation, profile.ShowBirthday,
		profile.SocialLinksJSON, profile.PreferencesJSON,
		profile.ProfileViewsCount, profile.FollowersCount, profile.FollowingCount,
		profile.IsBanned, profile.CreatedAt, profile.UpdatedAt,
	)
	require.NoError(t, err, "Failed to create test profile")

	return profile
}
