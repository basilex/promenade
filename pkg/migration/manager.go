package migration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/database"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/jmoiron/sqlx"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IMigrationManager manages database migrations with namespace support
type IMigrationManager interface {
	// MigrateNamespace applies all pending migrations for a namespace
	MigrateNamespace(ctx context.Context, namespace string) error

	// MigrateAll applies all pending migrations (core + enabled modules)
	MigrateAll(ctx context.Context, enabledModules []string) error

	// Rollback rolls back N migrations for a namespace
	Rollback(ctx context.Context, namespace string, steps int) error

	// Version returns current version for a namespace
	Version(ctx context.Context, namespace string) (int, error)

	// Status returns migration status for all namespaces
	Status(ctx context.Context) (map[string]MigrationStatus, error)
}

// MigrationStatus represents the migration status for a namespace
type MigrationStatus struct {
	Namespace      string
	CurrentVersion int
	PendingCount   int
	Dirty          bool
}

// MigrationFile represents a migration file
type MigrationFile struct {
	Version     int
	Description string
	Namespace   string
	UpSQL       string
	DownSQL     string
}

type manager struct {
	db            *sqlx.DB
	migrationsDir string
	driver        string // Database driver: postgres
	dialect       database.Dialect
}

// NewManager creates a new migration manager
// driver: "postgres"
// migrationsDir: base directory (e.g., "migrations"), manager will look in {migrationsDir}/{driver}/
// dialect: database dialect for SQL syntax differences
func NewManager(db *sqlx.DB, driver string, migrationsDir string, dialect database.Dialect) IMigrationManager {
	return &manager{
		db:            db,
		migrationsDir: migrationsDir,
		driver:        driver,
		dialect:       dialect,
	}
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *manager) ensureMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     BIGINT       NOT NULL,
			namespace   VARCHAR(50)  NOT NULL,
			dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
			applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (namespace, version)
		);

		CREATE INDEX IF NOT EXISTS idx_schema_migrations_namespace 
		ON schema_migrations(namespace);
	`

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getCurrentVersion returns the current version for a namespace
func (m *manager) getCurrentVersion(ctx context.Context, namespace string) (int, bool, error) {
	var version int
	var dirty bool

	query := fmt.Sprintf(`
		SELECT version, dirty 
		FROM schema_migrations 
		WHERE namespace = %s 
		ORDER BY version DESC 
		LIMIT 1
	`, m.dialect.Placeholder(1))

	err := m.db.QueryRowContext(ctx, query, namespace).Scan(&version, &dirty)
	if err == sql.ErrNoRows {
		return 0, false, nil // No migrations applied yet
	}
	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

// setVersion sets the current version for a namespace
func (m *manager) setVersion(ctx context.Context, namespace string, version int, dirty bool) error {
	query := fmt.Sprintf(`
		INSERT INTO schema_migrations (namespace, version, dirty, applied_at)
		VALUES (%s, %s, %s, %s)
		ON CONFLICT (namespace, version) 
		DO UPDATE SET dirty = %s, applied_at = %s
	`, m.dialect.Placeholder(1), m.dialect.Placeholder(2), m.dialect.Placeholder(3),
		m.dialect.Placeholder(4), m.dialect.Placeholder(3), m.dialect.Placeholder(4))

	_, err := m.db.ExecContext(ctx, query, namespace, version, dirty, time.Now())
	return err
}

// deleteVersion removes a version from schema_migrations
func (m *manager) deleteVersion(ctx context.Context, namespace string, version int) error {
	query := fmt.Sprintf("DELETE FROM schema_migrations WHERE namespace = %s AND version = %s",
		m.dialect.Placeholder(1), m.dialect.Placeholder(2))
	_, err := m.db.ExecContext(ctx, query, namespace, version)
	return err
}

// loadMigrationFiles reads migration files from disk for a namespace
func (m *manager) loadMigrationFiles(namespace string) ([]MigrationFile, error) {
	// migrationsDir already includes driver path: migrations/{driver}
	namespacePath := filepath.Join(m.migrationsDir, namespace)

	// Check if namespace directory exists
	if _, err := os.Stat(namespacePath); os.IsNotExist(err) {
		logger.Warn("Migration directory not found", "namespace", namespace, "driver", m.driver, "path", namespacePath)
		return []MigrationFile{}, nil // No migrations for this namespace
	}

	files, err := os.ReadDir(namespacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	migrations := make(map[int]*MigrationFile)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()

		// Parse filename: 000001_description.up.sql or 000001_description.down.sql
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		// Read file content
		content, err := os.ReadFile(filepath.Join(namespacePath, filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Determine if it's up or down migration
		if strings.HasSuffix(filename, ".up.sql") {
			if migrations[version] == nil {
				migrations[version] = &MigrationFile{
					Version:   version,
					Namespace: namespace,
				}
			}
			migrations[version].UpSQL = string(content)

			// Extract description from filename
			descParts := strings.Split(filename, "_")
			if len(descParts) > 1 {
				desc := strings.Join(descParts[1:], "_")
				desc = strings.TrimSuffix(desc, ".up.sql")
				migrations[version].Description = desc
			}
		} else if strings.HasSuffix(filename, ".down.sql") {
			if migrations[version] == nil {
				migrations[version] = &MigrationFile{
					Version:   version,
					Namespace: namespace,
				}
			}
			migrations[version].DownSQL = string(content)
		}
	}

	// Convert map to sorted slice
	result := make([]MigrationFile, 0, len(migrations))
	for _, mig := range migrations {
		result = append(result, *mig)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

// MigrateNamespace applies all pending migrations for a namespace
func (m *manager) MigrateNamespace(ctx context.Context, namespace string) error {
	log := logger.FromContext(ctx)

	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, dirty, err := m.getCurrentVersion(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("namespace %s is in dirty state at version %d, manual intervention required", namespace, currentVersion)
	}

	// Load migration files
	migrations, err := m.loadMigrationFiles(namespace)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	if len(migrations) == 0 {
		log.Info("No migrations found", "namespace", namespace)
		return nil
	}

	// Find pending migrations
	pending := []MigrationFile{}
	for _, mig := range migrations {
		if mig.Version > currentVersion {
			pending = append(pending, mig)
		}
	}

	if len(pending) == 0 {
		log.Info("No pending migrations", "namespace", namespace, "current_version", currentVersion)
		return nil
	}

	log.Info("Applying migrations",
		"namespace", namespace,
		"current_version", currentVersion,
		"pending_count", len(pending))

	// Apply each pending migration
	for _, mig := range pending {
		if err := m.applyMigration(ctx, mig); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", mig.Version, err)
		}

		log.Info("Applied migration",
			"namespace", namespace,
			"version", mig.Version,
			"description", mig.Description)
	}

	log.Info("All migrations applied successfully", "namespace", namespace)
	return nil
}

// applyMigration applies a single migration in a transaction
func (m *manager) applyMigration(ctx context.Context, mig MigrationFile) error {
	log := logger.FromContext(ctx)

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Mark as dirty
	if err = m.setVersion(ctx, mig.Namespace, mig.Version, true); err != nil {
		return fmt.Errorf("failed to mark as dirty: %w", err)
	}

	// Log SQL for debugging
	log.Debug("Executing migration SQL",
		"namespace", mig.Namespace,
		"version", mig.Version,
		"sql_length", len(mig.UpSQL),
		"sql_preview", mig.UpSQL[:min(200, len(mig.UpSQL))])

	// Execute migration
	if _, err = tx.ExecContext(ctx, mig.UpSQL); err != nil {
		log.Error("Migration execution failed",
			"namespace", mig.Namespace,
			"version", mig.Version,
			"error", err)
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Mark as clean
	if err = m.setVersion(ctx, mig.Namespace, mig.Version, false); err != nil {
		return fmt.Errorf("failed to mark as clean: %w", err)
	}

	return tx.Commit()
}

// MigrateAll applies all pending migrations (core + enabled modules)
func (m *manager) MigrateAll(ctx context.Context, enabledModules []string) error {
	log := logger.FromContext(ctx)

	log.Info("Starting migrations", "enabled_modules", enabledModules)

	// Always migrate core first
	if err := m.MigrateNamespace(ctx, "core"); err != nil {
		return fmt.Errorf("failed to migrate core: %w", err)
	}

	// Migrate each enabled module
	for _, module := range enabledModules {
		if err := m.MigrateNamespace(ctx, module); err != nil {
			return fmt.Errorf("failed to migrate module %s: %w", module, err)
		}
	}

	log.Info("All migrations completed successfully")
	return nil
}

// Rollback rolls back N migrations for a namespace
func (m *manager) Rollback(ctx context.Context, namespace string, steps int) error {
	log := logger.FromContext(ctx)

	// Get current version
	currentVersion, dirty, err := m.getCurrentVersion(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("namespace %s is in dirty state at version %d", namespace, currentVersion)
	}

	if currentVersion == 0 {
		log.Info("No migrations to rollback", "namespace", namespace)
		return nil
	}

	// Load migration files
	migrations, err := m.loadMigrationFiles(namespace)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	// Find migrations to rollback (in reverse order)
	toRollback := []MigrationFile{}
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].Version <= currentVersion && len(toRollback) < steps {
			toRollback = append(toRollback, migrations[i])
		}
	}

	if len(toRollback) == 0 {
		log.Info("No migrations to rollback", "namespace", namespace)
		return nil
	}

	log.Info("Rolling back migrations",
		"namespace", namespace,
		"current_version", currentVersion,
		"steps", len(toRollback))

	// Rollback each migration
	for _, mig := range toRollback {
		if err := m.rollbackMigration(ctx, mig); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", mig.Version, err)
		}

		log.Info("Rolled back migration",
			"namespace", namespace,
			"version", mig.Version,
			"description", mig.Description)
	}

	return nil
}

// rollbackMigration rolls back a single migration
func (m *manager) rollbackMigration(ctx context.Context, mig MigrationFile) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Execute down migration
	if mig.DownSQL == "" {
		return fmt.Errorf("no down migration found for version %d", mig.Version)
	}

	if _, err = tx.ExecContext(ctx, mig.DownSQL); err != nil {
		return fmt.Errorf("failed to execute down migration: %w", err)
	}

	// Delete version record
	if err = m.deleteVersion(ctx, mig.Namespace, mig.Version); err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	return tx.Commit()
}

// Version returns current version for a namespace
func (m *manager) Version(ctx context.Context, namespace string) (int, error) {
	version, _, err := m.getCurrentVersion(ctx, namespace)
	return version, err
}

// Status returns migration status for all namespaces
func (m *manager) Status(ctx context.Context) (map[string]MigrationStatus, error) {
	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all namespaces from schema_migrations
	query := `SELECT DISTINCT namespace FROM schema_migrations ORDER BY namespace`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() // Ignore close error
	}()

	namespaces := []string{}
	for rows.Next() {
		var namespace string
		if err := rows.Scan(&namespace); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, namespace)
	}

	// Also check filesystem for namespaces
	files, err := os.ReadDir(m.migrationsDir)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				found := false
				for _, ns := range namespaces {
					if ns == file.Name() {
						found = true
						break
					}
				}
				if !found {
					namespaces = append(namespaces, file.Name())
				}
			}
		}
	}

	// Get status for each namespace
	result := make(map[string]MigrationStatus)
	for _, namespace := range namespaces {
		currentVersion, dirty, err := m.getCurrentVersion(ctx, namespace)
		if err != nil {
			return nil, err
		}

		migrations, err := m.loadMigrationFiles(namespace)
		if err != nil {
			return nil, err
		}

		pendingCount := 0
		for _, mig := range migrations {
			if mig.Version > currentVersion {
				pendingCount++
			}
		}

		result[namespace] = MigrationStatus{
			Namespace:      namespace,
			CurrentVersion: currentVersion,
			PendingCount:   pendingCount,
			Dirty:          dirty,
		}
	}

	return result, nil
}
