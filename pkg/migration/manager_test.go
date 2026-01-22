package migration

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// mockDialect implements database.Dialect for testing
type mockDialect struct{}

func (d *mockDialect) Name() string                       { return "postgres" }
func (d *mockDialect) Placeholder(n int) string           { return fmt.Sprintf("$%d", n) }
func (d *mockDialect) SupportsReturning() bool            { return true }
func (d *mockDialect) SupportsJSON() bool                 { return true }
func (d *mockDialect) SupportsJSONIndex() bool            { return true }
func (d *mockDialect) SupportsUUID() bool                 { return true }
func (d *mockDialect) QuoteIdentifier(name string) string { return `"` + name + `"` }
func (d *mockDialect) UUIDType() string                   { return "UUID" }
func (d *mockDialect) JSONType() string                   { return "JSONB" }
func (d *mockDialect) TimestampType() string              { return "TIMESTAMP" }
func (d *mockDialect) BoolType() string                   { return "BOOLEAN" }

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock
}

func TestNewManager(t *testing.T) {
	db, _ := setupMockDB(t)
	defer func() { _ = db.Close() }()

	dialect := &mockDialect{}
	mgr := NewManager(db, "postgres", "test_migrations", dialect)

	assert.NotNil(t, mgr)
}

func TestMigrationFile_Structure(t *testing.T) {
	mf := MigrationFile{
		Version:     1,
		Description: "initial_schema",
		Namespace:   "core",
		UpSQL:       "CREATE TABLE users (id BIGINT);",
		DownSQL:     "DROP TABLE users;",
	}

	assert.Equal(t, 1, mf.Version)
	assert.Equal(t, "initial_schema", mf.Description)
	assert.Equal(t, "core", mf.Namespace)
	assert.NotEmpty(t, mf.UpSQL)
	assert.NotEmpty(t, mf.DownSQL)
}

func TestMigrationStatus_Structure(t *testing.T) {
	status := MigrationStatus{
		Namespace:      "core",
		CurrentVersion: 5,
		PendingCount:   2,
		Dirty:          false,
	}

	assert.Equal(t, "core", status.Namespace)
	assert.Equal(t, 5, status.CurrentVersion)
	assert.Equal(t, 2, status.PendingCount)
	assert.False(t, status.Dirty)
}

func TestManager_InterfaceCompliance(t *testing.T) {
	db, _ := setupMockDB(t)
	defer func() { _ = db.Close() }()

	dialect := &mockDialect{}
	var mgr interface{} = NewManager(db, "postgres", "test", dialect)

	_, ok := mgr.(IMigrationManager)
	assert.True(t, ok, "manager should implement IMigrationManager interface")
}
