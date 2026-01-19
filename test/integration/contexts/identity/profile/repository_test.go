package profile_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/profile/adapter/repository/postgres"
	profileAggregate "github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestProfileRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewProfileRepository(testDB.DB)
		userID := uuidv7.New()

		// Create user
		_, err := tx.Exec(`INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
			userID, fmt.Sprintf("user_%s@test.com", userID), "hash", "active")
		require.NoError(t, err)

		// Create
		p, err := profileAggregate.NewProfile(userID, "Test User")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		// Read by ID
		found, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "Test User", found.DisplayName)
		assert.Equal(t, userID, found.UserID)

		// Read by UserID
		foundByUser, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, foundByUser.ID)

		// Update
		_ = found.UpdateDisplayName("Updated User")
		_ = found.UpdateBio("New bio")
		require.NoError(t, repo.Update(ctx, found))
		updated, _ := repo.GetByID(ctx, p.ID)
		assert.Equal(t, "Updated User", updated.DisplayName)
		assert.Equal(t, "New bio", updated.Bio)

		// Delete
		require.NoError(t, repo.Delete(ctx, p.ID))
		_, err = repo.GetByID(ctx, p.ID)
		assert.Error(t, err)
	})
}

func TestProfileRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewProfileRepository(testDB.DB)

		// Create 3 profiles (2 public, 1 private)
		for i := range 3 {
			userID := uuidv7.New()
			_, err := tx.Exec(`INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
				userID, fmt.Sprintf("user%d@test.com", i), "hash", "active")
			require.NoError(t, err)

			p, _ := profileAggregate.NewProfile(userID, fmt.Sprintf("User %d", i))
			if i < 2 {
				p.SetPublic() // Make first 2 public
			}
			require.NoError(t, repo.Create(ctx, p))
		}

		// List public profiles
		publicProfiles, err := repo.ListPublicProfiles(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(publicProfiles), 2)
		for _, p := range publicProfiles {
			assert.True(t, p.IsPublic)
		}
	})
}
