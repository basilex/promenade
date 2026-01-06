package user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestUserRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewUserRepository(testDB.DB)

		// Create with unique email
		email := fmt.Sprintf("test_%s@example.com", uuidv7.New().String())
		u, err := user.NewUser(email, "password123")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, u))
		assert.NotEqual(t, "", u.ID.String())

		// GetByID
		found, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, email, found.Email.Value())

		// GetByEmail
		foundByEmail, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, u.ID, foundByEmail.ID)

		// Update (status)
		u.Activate()
		require.NoError(t, repo.Update(ctx, u))
		updated, _ := repo.GetByID(ctx, u.ID)
		assert.Equal(t, user.UserStatusActive, updated.Status)

		// Update (password)
		require.NoError(t, u.ChangePassword("newpassword123"))
		require.NoError(t, repo.Update(ctx, u))
		withNewPass, _ := repo.GetByID(ctx, u.ID)
		assert.NoError(t, withNewPass.CheckPassword("newpassword123"))

		// Delete
		require.NoError(t, repo.Delete(ctx, u.ID))
		_, err = repo.GetByID(ctx, u.ID)
		assert.Error(t, err)
	})
}

func TestUserRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	// Create 2 users with unique emails
	email1 := fmt.Sprintf("user1_%s@example.com", uuidv7.New().String())
	email2 := fmt.Sprintf("user2_%s@example.com", uuidv7.New().String())
	u1, err := user.NewUser(email1, "password123")
	require.NoError(t, err)
	u2, err := user.NewUser(email2, "password456")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, u1))
	require.NoError(t, repo.Create(ctx, u2))

	// ExistsByEmail
	exists, err := repo.ExistsByEmail(ctx, email1)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByEmail(ctx, fmt.Sprintf("nonexistent_%s@example.com", uuidv7.New().String()))
	require.NoError(t, err)
	assert.False(t, exists)

	// ListUsers
	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize
	users, total, err := repo.ListUsers(ctx, pageSize, offset) // limit, offset
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 2)
	assert.GreaterOrEqual(t, len(users), 2)

		// ValueObject roundtrip (Email preservation)
		email3 := fmt.Sprintf("value_%s@object.com", uuidv7.New().String())
		u3, err := user.NewUser(email3, "password789")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, u3))
		retrieved, _ := repo.GetByID(ctx, u3.ID)
		assert.Equal(t, email3, retrieved.Email.Value())


	// Create user WITHOUT transaction so it persists for concurrent tests
	var userID uuidv7.UUID
	func() {
		repo := postgres.NewUserRepository(testDB.DB)
		email := fmt.Sprintf("concurrent_%s@example.com", uuidv7.New().String())
		u, _ := user.NewUser(email, "password123")
		require.NoError(t, repo.Create(context.Background(), u))
		userID = u.ID
	}()

	// Cleanup after test
	defer func() {
		_, _ = testDB.DB.Exec("DELETE FROM identity_users WHERE id = $1", userID)
	}()

	// Run 10 concurrent updates
	var g errgroup.Group

	for range 10 {
		g.Go(func() error {
			testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
				repo := postgres.NewUserRepository(testDB.DB)
				u, err := repo.GetByID(ctx, userID)
				if err != nil {
					t.Errorf("failed to get user: %v", err)
					return
				}
				u.Activate()
				if err := repo.Update(ctx, u); err != nil {
					t.Errorf("failed to update user: %v", err)
				}
			})
			return nil
		})
	}

	// Wait for all goroutines
	if err := g.Wait(); err != nil {
		t.Fatalf("errgroup wait failed: %v", err)
	}

	// Verify final state (use context.Background(), not a transaction)
	finalRepo := postgres.NewUserRepository(testDB.DB)
	u, err := finalRepo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, user.UserStatusActive, u.Status)
}
