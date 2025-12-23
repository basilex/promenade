//go:build integration
// +build integration

package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/test/integration"
)

func TestUserRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	repo := postgres.NewUserRepository(testDB.DB)
	ctx := testDB.GetContext()

	t.Run("GetByEmail", func(t *testing.T) {
		user := fixtures.CreateUser(t, "email-test@example.com", "password123")

		retrieved, err := repo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})
}
