package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"
)

// TestDB holds test database connection and cleanup functions
type TestDB struct {
	DB             *sqlx.DB
	TransactionMgr database.TransactionManager
	Logger         *slog.Logger
	cleanup        []func()
}

// Config for test database
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns default test database configuration
func DefaultConfig() Config {
	return Config{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5433"), // Use dedicated test container port
		User:     getEnv("TEST_DB_USER", "promenade"),
		Password: getEnv("TEST_DB_PASSWORD", "promenade"),
		DBName:   getEnv("TEST_DB_NAME", "promenade_test"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}
}

// SetupTestDB initializes a test database connection with migrations
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	cfg := DefaultConfig()
	db, err := connectDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create logger
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
	})

	// Run migrations
	if err := runMigrations(db, log.Logger); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	testDB := &TestDB{
		DB:             db,
		TransactionMgr: database.NewTransactionManager(db),
		Logger:         log.Logger,
		cleanup:        []func(){},
	}

	// Register cleanup
	t.Cleanup(func() {
		testDB.Cleanup()
	})

	return testDB
}

// SetupTestDBWithCleanTables sets up test database and cleans all tables before tests
func SetupTestDBWithCleanTables(t *testing.T) *TestDB {
	t.Helper()

	testDB := SetupTestDB(t)
	testDB.CleanAllTables()

	return testDB
}

// CleanAllTables truncates all tables for a fresh test state
func (tdb *TestDB) CleanAllTables() {
	// Order matters - respect foreign key constraints
	tables := []string{
		// Module tables (analytics)
		"analytics_dashboard_widgets",
		"analytics_dashboards",
		"analytics_report_schedules",
		"analytics_reports",
		"analytics_metric_aggregates",
		"analytics_metrics",
		// Module tables (profiles)
		"profiles_contacts",
		"profiles_profiles",
		// Module tables (posts)
		"posts_comment_likes",
		"posts_comments",
		"posts_posts",

		// Core tables
		"core_user_sessions",
		"core_user_roles",
		"core_role_permissions",
		"core_permissions",
		"core_roles",
		"core_login_attempts",
		"core_password_reset_tokens",
		"core_email_verification_tokens",
		"core_users",
		"core_cities",
		"core_regions",
		"core_country_currencies",
		"core_currencies",
		"core_countries",
		"core_timezones",
		"core_languages",
		"core_payment_methods",
	}

	for _, table := range tables {
		// Check if table exists first
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`
		if err := tdb.DB.Get(&exists, query, table); err != nil {
			continue
		}

		if exists {
			_, _ = tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		}
	}
}

// WithTransaction runs a test function within a transaction that is rolled back
func (tdb *TestDB) WithTransaction(t *testing.T, fn func(ctx context.Context, tx *sqlx.Tx)) {
	t.Helper()

	ctx := context.Background()
	tx, err := tdb.DB.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("Failed to rollback transaction: %v", err)
		}
	}()

	fn(ctx, tx)
}

// MustExec executes a query and fails the test on error
func (tdb *TestDB) MustExec(t *testing.T, query string, args ...interface{}) {
	t.Helper()

	_, err := tdb.DB.Exec(query, args...)
	if err != nil {
		t.Fatalf("Failed to execute query: %v\nQuery: %s", err, query)
	}
}

// GetContext returns a context with logger
func (tdb *TestDB) GetContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.RequestIDKey, "test-request-id")
	return ctx
}

// Cleanup performs cleanup operations
func (tdb *TestDB) Cleanup() {
	for _, fn := range tdb.cleanup {
		fn()
	}

	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// AddCleanup registers a cleanup function
func (tdb *TestDB) AddCleanup(fn func()) {
	tdb.cleanup = append(tdb.cleanup, fn)
}

// Helper functions

func connectDB(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func runMigrations(db *sqlx.DB, log *slog.Logger) error {
	mgr := migration.NewManager(db, "migrations")
	ctx := context.Background()

	// Run core migrations
	if err := mgr.MigrateNamespace(ctx, "core"); err != nil {
		return fmt.Errorf("core migrations failed: %w", err)
	}

	// Run module migrations (only enabled ones)
	modules := []string{"posts", "profiles", "analytics"}
	for _, module := range modules {
		if err := mgr.MigrateNamespace(ctx, module); err != nil {
			return fmt.Errorf("%s module migrations failed: %w", module, err)
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
