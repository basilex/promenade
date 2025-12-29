package user_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func TestUserRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewUserRepository(testDB.DB)

		// Create with unique email
		email := valueobject.MustNewEmail(fmt.Sprintf("test_%s@example.com", uuidv7.New().String()))
		u, err := user.NewUser(email.Value(), "password123")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, u))
		assert.NotEqual(t, "", u.ID.String())

		// GetByID
		found, err := repo.GetByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, email.Value(), found.Email.Value())

		// GetByEmail
		foundByEmail, err := repo.GetByEmail(ctx, email.Value())
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
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewUserRepository(testDB.DB)

		// Create 2 users with unique emails
		email1 := valueobject.MustNewEmail(fmt.Sprintf("user1_%s@example.com", uuidv7.New().String()))
		email2 := valueobject.MustNewEmail(fmt.Sprintf("user2_%s@example.com", uuidv7.New().String()))
		u1, _ := user.NewUser(email1.Value(), "pass1")
		u2, _ := user.NewUser(email2.Value(), "pass2")
		require.NoError(t, repo.Create(ctx, u1))
		require.NoError(t, repo.Create(ctx, u2))

		// ExistsByEmail
		exists, err := repo.ExistsByEmail(ctx, email1.Value())
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByEmail(ctx, "nonexistent@example.com")
		require.NoError(t, err)
		assert.False(t, exists)

		// ListUsers
		users, total, err := repo.ListUsers(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(users), 2)

		// ValueObject roundtrip (Email preservation)
		email1 := "VALUE@OBJECT.COM"
		email2, _ := valueobject.NewEmail(email1)
		u3, _ := user.NewUser(email2.Value(), "pass3")
		require.NoError(t, repo.Create(ctx, u3))
		retrieved, _ := repo.GetByID(ctx, u3.ID)
		assert.Equal(t, email2.Value(), retrieved.Email.Value())
	})
}

func TestUserRepository_ConcurrentUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	var userID uuidv7.UUID
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewUserRepository(testDB.DB)
		u, _ := user.NewUser("concurrent@example.com", "password123")
		require.NoError(t, repo.Create(ctx, u))
		userID = u.ID
	})

	// Run 10 concurrent updates
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for range 10 {
		wg.Go(func() {
			testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
				repo := postgres.NewUserRepository(testDB.DB)
				u, err := repo.GetByID(ctx, userID)
				if err != nil {
					errors <- err
					return
				}
				u.Activate()
				if err := repo.Update(ctx, u); err != nil {
					errors <- err
				}
			})
		})
	}

	wg.Wait()
	close(errors)

	// Check no errors occurred
	for err := range errors {
		assert.NoError(t, err)
	}

	// Verify final state
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewUserRepository(testDB.DB)
		u, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, user.UserStatusActive, u.Status)
	})
}
