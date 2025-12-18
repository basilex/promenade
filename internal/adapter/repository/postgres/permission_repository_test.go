package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

func TestPermissionRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t) // Clean migration data first
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("creates permission successfully", func(t *testing.T) {
		perm := helpers.PermissionFixture("posts", "create")
		err := repo.Create(ctx, perm)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.Equal(t, perm.Resource, retrieved.Resource)
		assert.Equal(t, perm.Action, retrieved.Action)
	})

	t.Run("creates wildcard permission", func(t *testing.T) {
		perm := helpers.PermissionFixture("*", "*")
		err := repo.Create(ctx, perm)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.Equal(t, "*", retrieved.Resource)
		assert.Equal(t, "*", retrieved.Action)
	})
}

func TestPermissionRepository_GetByResourceAction(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("finds permission by resource and action", func(t *testing.T) {
		perm := helpers.PermissionFixture("posts", "create")
		err := repo.Create(ctx, perm)
		require.NoError(t, err)

		retrieved, err := repo.GetByResourceAction(ctx, "posts", "create")
		require.NoError(t, err)
		assert.Equal(t, perm.ID, retrieved.ID)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		_, err := repo.GetByResourceAction(ctx, "nonexistent", "action")
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestPermissionRepository_List(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("lists all permissions", func(t *testing.T) {
		perm1 := helpers.PermissionFixture("posts", "create")
		perm2 := helpers.PermissionFixture("posts", "read")
		perm3 := helpers.PermissionFixture("users", "create")

		require.NoError(t, repo.Create(ctx, perm1))
		require.NoError(t, repo.Create(ctx, perm2))
		require.NoError(t, repo.Create(ctx, perm3))

		permissions, err := repo.List(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(permissions), 3)
	})

	t.Run("returns empty list when no permissions", func(t *testing.T) {
		permissions, err := repo.List(ctx)
		require.NoError(t, err)
		assert.NotNil(t, permissions)
	})
}

func TestPermissionRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("updates permission successfully", func(t *testing.T) {
		perm := helpers.PermissionFixture("posts", "create")
		err := repo.Create(ctx, perm)
		require.NoError(t, err)

		newDesc := "Updated description"
		perm.Description = &newDesc
		err = repo.Update(ctx, perm)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.Equal(t, newDesc, *retrieved.Description)
	})

	t.Run("returns error when permission not found", func(t *testing.T) {
		perm := helpers.PermissionFixture("nonexistent", "action")
		perm.ID = uuidv7.New()
		err := repo.Update(ctx, perm)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestPermissionRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("deletes permission successfully", func(t *testing.T) {
		perm := helpers.PermissionFixture("posts", "create")
		err := repo.Create(ctx, perm)
		require.NoError(t, err)

		err = repo.Delete(ctx, perm.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, perm.ID)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})

	t.Run("returns error when permission not found", func(t *testing.T) {
		err := repo.Delete(ctx, uuidv7.New())
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestPermissionRepository_CreateMany(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("creates multiple permissions", func(t *testing.T) {
		perms := []*entity.Permission{
			helpers.PermissionFixture("posts", "create"),
			helpers.PermissionFixture("posts", "read"),
			helpers.PermissionFixture("posts", "update"),
		}

		err := repo.CreateMany(ctx, perms)
		require.NoError(t, err)

		for _, perm := range perms {
			retrieved, err := repo.GetByID(ctx, perm.ID)
			require.NoError(t, err)
			assert.Equal(t, perm.Resource, retrieved.Resource)
		}
	})

	t.Run("creates empty list successfully", func(t *testing.T) {
		err := repo.CreateMany(ctx, []*entity.Permission{})
		require.NoError(t, err)
	})
}

func TestPermissionRepository_GetByIDs(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("retrieves multiple permissions by IDs", func(t *testing.T) {
		perm1 := helpers.PermissionFixture("posts", "create")
		perm2 := helpers.PermissionFixture("posts", "read")
		perm3 := helpers.PermissionFixture("users", "create")

		require.NoError(t, repo.Create(ctx, perm1))
		require.NoError(t, repo.Create(ctx, perm2))
		require.NoError(t, repo.Create(ctx, perm3))

		ids := []uuidv7.UUID{perm1.ID, perm2.ID}
		retrieved, err := repo.GetByIDs(ctx, ids)
		require.NoError(t, err)
		assert.Len(t, retrieved, 2)
	})

	t.Run("returns empty list for non-existent IDs", func(t *testing.T) {
		ids := []uuidv7.UUID{uuidv7.New(), uuidv7.New()}
		retrieved, err := repo.GetByIDs(ctx, ids)
		require.NoError(t, err)
		assert.Empty(t, retrieved)
	})
}

func TestPermissionRepository_FindByResource(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("finds permissions by resource", func(t *testing.T) {
		perm1 := helpers.PermissionFixture("posts", "create")
		perm2 := helpers.PermissionFixture("posts", "read")
		perm3 := helpers.PermissionFixture("users", "create")

		require.NoError(t, repo.Create(ctx, perm1))
		require.NoError(t, repo.Create(ctx, perm2))
		require.NoError(t, repo.Create(ctx, perm3))

		permissions, err := repo.FindByResource(ctx, "posts")
		require.NoError(t, err)
		assert.Len(t, permissions, 2)

		for _, perm := range permissions {
			assert.Equal(t, "posts", perm.Resource)
		}
	})

	t.Run("returns empty list when resource not found", func(t *testing.T) {
		permissions, err := repo.FindByResource(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, permissions)
	})
}
