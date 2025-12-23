//go:build integration

package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5433")
	user := getEnv("TEST_DB_USER", "promenade")
	password := getEnv("TEST_DB_PASSWORD", "promenade")
	dbname := getEnv("TEST_DB_NAME", "promenade_test")

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
	
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestTransactionManager_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	tm := database.NewTransactionManager(db)

	t.Run("Successful transaction commits", func(t *testing.T) {
		userID := uuidv7.New().String()
		name := "tx_commit_user"

		err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			query := `INSERT INTO core_users (id, name, email, password) VALUES ($1, $2, $3, $4)`
			
			executor, ok := database.GetTx(txCtx)
			if !ok || executor == nil {
				t.Fatal("Transaction not found in context")
			}

			_, err := executor.ExecContext(txCtx, query, userID, name, "tx@test.com", "hash")
			return err
		})

		require.NoError(t, err)

		// Verify user was committed
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM core_users WHERE id = $1", userID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Failed transaction rolls back", func(t *testing.T) {
		userID := uuidv7.New().String()
		name := "tx_rollback_user"

		err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			query := `INSERT INTO core_users (id, name, email, password) VALUES ($1, $2, $3, $4)`
			
			executor, ok := database.GetTx(txCtx)
			if !ok || executor == nil {
				t.Fatal("Transaction not found in context")
			}

			_, err := executor.ExecContext(txCtx, query, userID, name, "rollback@test.com", "hash")
			if err != nil {
				return err
			}

			// Force error to trigger rollback
			return assert.AnError
		})

		require.Error(t, err)

		// Verify user was NOT committed
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM core_users WHERE id = $1", userID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Nested transaction operations", func(t *testing.T) {
		userID := uuidv7.New().String()
		name := "tx_nested_user"

		err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			executor, ok := database.GetTx(txCtx)
			if !ok || executor == nil {
				t.Fatal("Transaction not found in context")
			}

			// First insert
			query1 := `INSERT INTO core_users (id, name, email, password) VALUES ($1, $2, $3, $4)`
			_, err := executor.ExecContext(txCtx, query1, userID, name, "nested@test.com", "hash")
			if err != nil {
				return err
			}

			// Second operation in same transaction
			query2 := `UPDATE core_users SET email = $1 WHERE id = $2`
			_, err = executor.ExecContext(txCtx, query2, "updated_nested@test.com", userID)
			return err
		})

		require.NoError(t, err)

		// Verify both operations committed
		var email string
		err = db.Get(&email, "SELECT email FROM core_users WHERE id = $1", userID)
		require.NoError(t, err)
		assert.Equal(t, "updated_nested@test.com", email)
	})

	t.Run("Multiple transactions are isolated", func(t *testing.T) {
		user1ID := uuidv7.New().String()
		user2ID := uuidv7.New().String()

		// Transaction 1 - succeeds
		err1 := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			executor, _ := database.GetTx(txCtx)
			query := `INSERT INTO core_users (id, name, email, password) VALUES ($1, $2, $3, $4)`
			_, err := executor.ExecContext(txCtx, query, user1ID, "user1", "user1@test.com", "hash")
			return err
		})

		// Transaction 2 - fails
		err2 := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			executor, _ := database.GetTx(txCtx)
			query := `INSERT INTO core_users (id, username, email, password_hash) VALUES ($1, $2, $3, $4)`
			_, err := executor.ExecContext(txCtx, query, user2ID, "user2", "user2@test.com", "hash")
			if err != nil {
				return err
			}
			return assert.AnError // Force rollback
		})

		require.NoError(t, err1)
		require.Error(t, err2)

		// Verify only first transaction committed
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM core_users WHERE id IN ($1, $2)", user1ID, user2ID)
		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only user1 committed
	})

	t.Run("Transaction context propagation", func(t *testing.T) {
		userID := uuidv7.New().String()

		err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			// Verify transaction is in context
			tx, ok := database.GetTx(txCtx)
			assert.True(t, ok, "Transaction should be in context")
			assert.NotNil(t, tx, "Transaction should not be nil")

			query := `INSERT INTO core_users (id, name, email, password) VALUES ($1, $2, $3, $4)`
			_, err := tx.ExecContext(txCtx, query, userID, "ctx_prop_user", "ctx@test.com", "hash")
			return err
		})

		require.NoError(t, err)

		// Verify outside transaction, GetTx returns nil
		tx, ok := database.GetTx(ctx)
		assert.False(t, ok, "Transaction should not exist outside WithTransaction")
		assert.Nil(t, tx, "Transaction should be nil outside WithTransaction")
	})
}

func cleanupTestData(t *testing.T, db *sqlx.DB) {
	t.Helper()
	
	// Clean up test users created during tests
	_, err := db.Exec("DELETE FROM core_users WHERE email LIKE '%@test.com'")
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}
