package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseRepository_Get(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()

	// Create test user
	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Test Get - retrieve user
	var retrieved struct {
		ID    uuidv7.UUID `db:"id"`
		Email string      `db:"email"`
		Name  string      `db:"name"`
	}

	query := `SELECT id, email, name FROM users WHERE id = $1`
	err := repo.Get(ctx, &retrieved, query, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.Name, retrieved.Name)

	// Test Get - not found
	err = repo.Get(ctx, &retrieved, query, uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestBaseRepository_Select(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()

	// Create multiple test users
	user1 := helpers.CreateTestUser(t, testDB.DB, "user1@example.com", "User One")
	user2 := helpers.CreateTestUser(t, testDB.DB, "user2@example.com", "User Two")

	// Test Select - retrieve multiple users
	var users []struct {
		ID    uuidv7.UUID `db:"id"`
		Email string      `db:"email"`
	}

	query := `SELECT id, email FROM users ORDER BY email`
	err := repo.Select(ctx, &users, query)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, user1.Email, users[0].Email)
	assert.Equal(t, user2.Email, users[1].Email)

	// Test Select - empty result
	var emptyResult []struct {
		ID uuidv7.UUID `db:"id"`
	}
	query = `SELECT id FROM users WHERE email = 'nonexistent@example.com'`
	err = repo.Select(ctx, &emptyResult, query)
	require.NoError(t, err)
	assert.Len(t, emptyResult, 0)
}

func TestBaseRepository_Exec(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Test Exec - update user
	query := `UPDATE users SET name = $1 WHERE id = $2`
	err := repo.Exec(ctx, query, "Updated Name", user.ID)
	require.NoError(t, err)

	// Verify update
	var updatedName string
	err = testDB.DB.Get(&updatedName, `SELECT name FROM users WHERE id = $1`, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedName)
}

func TestBaseRepository_NamedExec(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Test NamedExec - update using named parameters
	type updateParams struct {
		Name string      `db:"name"`
		ID   uuidv7.UUID `db:"id"`
	}

	query := `UPDATE users SET name = :name WHERE id = :id`
	params := updateParams{
		Name: "Named Update",
		ID:   user.ID,
	}

	err := repo.NamedExec(ctx, query, params)
	require.NoError(t, err)

	// Verify update
	var updatedName string
	err = testDB.DB.Get(&updatedName, `SELECT name FROM users WHERE id = $1`, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Named Update", updatedName)
}

func TestBaseRepository_WithTransaction(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()

	txManager := database.NewTransactionManager(testDB.DB)

	t.Run("successful transaction", func(t *testing.T) {
		var userID uuidv7.UUID

		err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			// Insert user within transaction
			userID = uuidv7.New()
			query := `INSERT INTO users (id, email, name, password, status, created_at, updated_at)
					  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
			return repo.Exec(txCtx, query, userID, "tx@example.com", "TX User", "password", "active")
		})

		require.NoError(t, err)

		// Verify user was created
		var count int
		err = testDB.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE id = $1`, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("rolled back transaction", func(t *testing.T) {
		initialCount := 0
		err := testDB.DB.Get(&initialCount, `SELECT COUNT(*) FROM users`)
		require.NoError(t, err)

		userID := uuidv7.New()
		err = txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			// Insert user
			query := `INSERT INTO users (id, email, name, password, status, created_at, updated_at)
					  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
			if err := repo.Exec(txCtx, query, userID, "rollback@example.com", "Rollback User", "password", "active"); err != nil {
				return err
			}

			// Force rollback by returning error
			return assert.AnError
		})

		require.Error(t, err)

		// Verify user was NOT created (transaction rolled back)
		var count int
		err = testDB.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE id = $1`, userID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify total count unchanged
		var finalCount int
		err = testDB.DB.Get(&finalCount, `SELECT COUNT(*) FROM users`)
		require.NoError(t, err)
		assert.Equal(t, initialCount, finalCount)
	})
}

func TestBaseRepository_GetExecutor(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()

	t.Run("returns db connection without transaction", func(t *testing.T) {
		executor := repo.getExecutor(ctx)
		assert.NotNil(t, executor)
		assert.Equal(t, testDB.DB, executor)
	})

	t.Run("returns transaction when present in context", func(t *testing.T) {
		txManager := database.NewTransactionManager(testDB.DB)

		err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			executor := repo.getExecutor(txCtx)
			assert.NotNil(t, executor)

			// Verify we got transaction, not DB
			tx, ok := database.GetTx(txCtx)
			assert.True(t, ok)
			assert.Equal(t, tx, executor)

			return nil
		})

		require.NoError(t, err)
	})
}

func TestBaseRepository_MultipleOperationsInTransaction(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewBaseRepository(testDB.DB)
	ctx := context.Background()
	txManager := database.NewTransactionManager(testDB.DB)

	// Test multiple operations within same transaction
	var user1ID, user2ID uuidv7.UUID

	err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Insert first user
		user1ID = uuidv7.New()
		query := `INSERT INTO users (id, email, name, password, status, created_at, updated_at)
				  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
		if err := repo.Exec(txCtx, query, user1ID, "user1@tx.com", "User 1", "pass", "active"); err != nil {
			return err
		}

		// Insert second user
		user2ID = uuidv7.New()
		if err := repo.Exec(txCtx, query, user2ID, "user2@tx.com", "User 2", "pass", "active"); err != nil {
			return err
		}

		// Query both users
		var users []struct {
			ID uuidv7.UUID `db:"id"`
		}
		selectQuery := `SELECT id FROM users WHERE id IN ($1, $2)`
		if err := repo.Select(txCtx, &users, selectQuery, user1ID, user2ID); err != nil {
			return err
		}

		assert.Len(t, users, 2)
		return nil
	})

	require.NoError(t, err)

	// Verify both users exist after transaction
	var count int
	err = testDB.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE id IN ($1, $2)`, user1ID, user2ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}
