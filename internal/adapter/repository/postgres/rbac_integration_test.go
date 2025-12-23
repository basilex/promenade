//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestRoleRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDB(t)
	defer testDB.CleanAllTables()

	ctx := context.Background()
	repo := NewRoleRepository(testDB.DB)

	t.Run("Create and GetByID", func(t *testing.T) {
		desc := "Test role"
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "test_role",
			DisplayName: "Test Role",
			Description: &desc,
		}

		err := repo.Create(ctx, role)
		require.NoError(t, err)

		fetchedRole, err := repo.GetByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Equal(t, role.ID, fetchedRole.ID)
		assert.Equal(t, role.Name, fetchedRole.Name)
	})

	t.Run("GetByName", func(t *testing.T) {
		desc := "Named role"
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "named_role",
			DisplayName: "Named Role",
			Description: &desc,
		}

		require.NoError(t, repo.Create(ctx, role))

		fetchedRole, err := repo.GetByName(ctx, role.Name)
		require.NoError(t, err)
		assert.Equal(t, role.ID, fetchedRole.ID)
		assert.Equal(t, role.Name, fetchedRole.Name)
	})
}

func TestPermissionRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDB(t)
	defer testDB.CleanAllTables()

	ctx := context.Background()
	repo := NewPermissionRepository(testDB.DB)

	t.Run("Create and GetByID", func(t *testing.T) {
		desc := "Test read permission"
		perm := &entity.Permission{
			ID:          uuidv7.New(),
			Resource:    "test",
			Action:      "read",
			Description: &desc,
		}

		err := repo.Create(ctx, perm)
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.Equal(t, perm.ID, fetched.ID)
		assert.Equal(t, perm.Resource, fetched.Resource)
		assert.Equal(t, perm.Action, fetched.Action)
	})
}
