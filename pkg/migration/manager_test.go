package migration

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

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

	mgr := NewManager(db, "test_migrations")

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

	var mgr interface{} = NewManager(db, "test")

	_, ok := mgr.(Manager)
	assert.True(t, ok, "manager should implement Manager interface")
}
