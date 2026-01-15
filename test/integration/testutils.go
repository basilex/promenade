// Package integration provides test utilities for integration testing with PostgreSQL.
// This file contains helpers used by repository tests across the codebase.
//
// Note: VS Code may show errors when editing this file standalone, but it compiles
// correctly when used by actual test files. This is a known IDE limitation.
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"
	"github.com/basilex/promenade/pkg/uuidv7"
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
	// Detect CI environment and use appropriate defaults
	// CI environments (GitHub Actions) set CI or GITHUB_ACTIONS env var
	// CI: PostgreSQL service on port 5432
	// Local: Docker test DB on port 5433
	defaultPort := "5433"
	
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		defaultPort = "5432"
	}
	
	return Config{
		Host:     getEnv("DB_HOST", getEnv("TEST_DB_HOST", "localhost")),
		Port:     getEnv("DB_PORT", getEnv("TEST_DB_PORT", defaultPort)),
		User:     getEnv("DB_USER", getEnv("TEST_DB_USER", "system")),
		Password: getEnv("DB_PASSWORD", getEnv("TEST_DB_PASSWORD", "passw0rd")),
		DBName:   getEnv("DB_NAME", getEnv("TEST_DB_NAME", "promenade_test")),
		SSLMode:  getEnv("DB_SSLMODE", getEnv("TEST_DB_SSLMODE", "disable")),
	}
}

// SetupTestDB initializes a test database connection with migrations
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Skip integration tests in short mode (unit tests)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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
		_ = db.Close()
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
	
	// Register cleanup BEFORE the DB close cleanup (using AddCleanup instead of t.Cleanup)
	// This ensures tables are cleaned while DB is still open
	testDB.AddCleanup(func() {
		testDB.CleanAllTables()
	})

	return testDB
}

// CleanAllTables truncates all tables for a fresh test state
func (tdb *TestDB) CleanAllTables() {
	// List all tables that need cleaning (order doesn't matter with CASCADE)
	tables := []string{
		// Customer Management Context tables
		"customer_interactions",
		"customer_deals",
		"customer_companies",
		"customer_customers",

		// Identity Context tables
		"identity_contacts",
		"identity_user_sessions",
		"identity_user_roles",
		"identity_role_permissions",
		"identity_permissions",
		"identity_roles",
		"identity_login_attempts",
		"identity_password_reset_tokens",
		"identity_email_verification_tokens",
		"identity_users",

		// Order Management Context tables
		"order_orders",
		"order_order_lines",

		// Billing Context tables
		"billing_invoices",
		"billing_invoice_lines",
		"billing_payments",
		"billing_subscriptions",

		// Warehouse Context tables
		"warehouse_stock_movements",
		"warehouse_inventory",
		"warehouse_products",
		"warehouse_locations",

		// Fiscal Context tables
		"fiscal_receipts",
		"fiscal_cash_registers",

		// Shared Kernel tables (reference data - do NOT truncate, needed for tests)
		// "shared_countries", "shared_currencies", "shared_languages", "shared_timezones",
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
			// Use TRUNCATE CASCADE to automatically handle foreign key constraints
			// This is simpler and more reliable than disabling FK checks
			_, _ = tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
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

	// Add transaction to context so repositories can use it via getExecutor(ctx)
	ctx = database.SetTxToContext(ctx, tx)

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
		_ = tdb.DB.Close()
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
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func runMigrations(db *sqlx.DB, log *slog.Logger) error {
	// Find project root (directory containing go.mod)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	migrationsPath := filepath.Join(projectRoot, "migrations")
	mgr := migration.NewManager(db, migrationsPath)
	ctx := context.Background()

	// Run core migrations
	if err := mgr.MigrateNamespace(ctx, "core"); err != nil {
		return fmt.Errorf("core migrations failed: %w", err)
	}

	// Run shared kernel migrations (reference data)
	if err := mgr.MigrateNamespace(ctx, "shared"); err != nil {
		return fmt.Errorf("shared migrations failed: %w", err)
	}

	// Run identity context migrations
	if err := mgr.MigrateNamespace(ctx, "identity"); err != nil {
		return fmt.Errorf("identity migrations failed: %w", err)
	}

	// Run customer management context migrations
	if err := mgr.MigrateNamespace(ctx, "customer-mgmt"); err != nil {
		return fmt.Errorf("customer-mgmt migrations failed: %w", err)
	}

	// Run order management context migrations
	if err := mgr.MigrateNamespace(ctx, "order-mgmt"); err != nil {
		return fmt.Errorf("order-mgmt migrations failed: %w", err)
	}

	// Run billing context migrations
	if err := mgr.MigrateNamespace(ctx, "billing"); err != nil {
		return fmt.Errorf("billing migrations failed: %w", err)
	}

	// Run warehouse context migrations
	if err := mgr.MigrateNamespace(ctx, "warehouse"); err != nil {
		return fmt.Errorf("warehouse migrations failed: %w", err)
	}

	// Run scripting context migrations
	if err := mgr.MigrateNamespace(ctx, "scripting"); err != nil {
		return fmt.Errorf("scripting migrations failed: %w", err)
	}

	// Run fiscal context migrations
	if err := mgr.MigrateNamespace(ctx, "fiscal"); err != nil {
		return fmt.Errorf("fiscal migrations failed: %w", err)
	}

	return nil
}

func findProjectRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up until we find go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ============================================================================
// Test Data Helpers
// ============================================================================

// FakeUUID generates a test UUID v7 for integration tests
func FakeUUID() uuidv7.UUID {
	return uuidv7.New()
}

// FakeSKU generates a test SKU with unique suffix
func FakeSKU(suffix int) string {
	return fmt.Sprintf("TEST-SKU-%03d-%s", suffix, uuidv7.New().String()[:8])
}

// FakeName generates a test name with prefix and suffix
func FakeName(prefix string, suffix int) string {
	return fmt.Sprintf("%s %d", prefix, suffix)
}
