package permission_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	"github.com/basilex/promenade/internal/contexts/identity/permission/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestPermissionRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("create permission successfully", func(t *testing.T) {
		p, err := permission.NewPermission("users", "read", "Read users")
		require.NoError(t, err)

		err = repo.Create(ctx, p)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, "users", retrieved.Resource)
		assert.Equal(t, "read", retrieved.Action)
	})

	t.Run("create permission with duplicate resource:action fails", func(t *testing.T) {
		p1, _ := permission.NewPermission("posts", "create", "Create posts 1")
		p2, _ := permission.NewPermission("posts", "create", "Create posts 2")

		err := repo.Create(ctx, p1)
		require.NoError(t, err)

		// Second create with same resource:action should fail
		err = repo.Create(ctx, p2)
		assert.Error(t, err)
	})
}

func TestPermissionRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing permission", func(t *testing.T) {
		p, _ := permission.NewPermission("orders", "update", "Update orders")
		require.NoError(t, repo.Create(ctx, p))

		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, "orders", retrieved.Resource)
	})

	t.Run("get non-existent permission returns error", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

func TestPermissionRepository_GetByName(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing permission by name", func(t *testing.T) {
		p, _ := permission.NewPermission("customers", "delete", "Delete customers")
		require.NoError(t, repo.Create(ctx, p))

		retrieved, err := repo.GetByName(ctx, "customers:delete")
		require.NoError(t, err)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, "customers:delete", retrieved.Name)
	})

	t.Run("get non-existent permission by name returns error", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "nonexistent:action")
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

func TestPermissionRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("update permission successfully", func(t *testing.T) {
		p, _ := permission.NewPermission("products", "create", "Original description")
		require.NoError(t, repo.Create(ctx, p))

		p.UpdateDescription("Updated description")

		err := repo.Update(ctx, p)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", retrieved.Description)
	})
}

func TestPermissionRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("delete permission successfully (soft delete)", func(t *testing.T) {
		p, _ := permission.NewPermission("inventory", "list", "List inventory")
		require.NoError(t, repo.Create(ctx, p))

		err := repo.Delete(ctx, p.ID)
		require.NoError(t, err)

		// Verify it's deleted (soft delete)
		_, err = repo.GetByID(ctx, p.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})

	t.Run("delete non-existent permission returns error", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		err := repo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

func TestPermissionRepository_ListPermissions(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("list all permissions", func(t *testing.T) {
		// Create several permissions
		p1, _ := permission.NewPermission("resource1", "read", "Test 1")
		p2, _ := permission.NewPermission("resource2", "write", "Test 2")
		p3, _ := permission.NewPermission("resource3", "delete", "Test 3")

		require.NoError(t, repo.Create(ctx, p1))
		require.NoError(t, repo.Create(ctx, p2))
		require.NoError(t, repo.Create(ctx, p3))

		permissions, total, err := repo.ListPermissions(ctx, 20, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(permissions), 3)
		assert.GreaterOrEqual(t, total, 3)
	})
}

func TestPermissionRepository_ExistsByName(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("existing permission returns true", func(t *testing.T) {
		p, _ := permission.NewPermission("reports", "generate", "Generate reports")
		require.NoError(t, repo.Create(ctx, p))

		exists, err := repo.ExistsByName(ctx, "reports:generate")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent permission returns false", func(t *testing.T) {
		exists, err := repo.ExistsByName(ctx, "does_not:exist")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestPermissionRepository_GetRolePermissions(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	permRepo := postgres.NewPermissionRepository(db.DB)
	ctx := context.Background()

	t.Run("get role permissions successfully", func(t *testing.T) {
		// Create permissions
		p1, _ := permission.NewPermission("documents", "read", "Read documents")
		p2, _ := permission.NewPermission("documents", "write", "Write documents")
		require.NoError(t, permRepo.Create(ctx, p1))
		require.NoError(t, permRepo.Create(ctx, p2))

		// Without role-permission assignments, should be empty
		roleID := uuidv7.New()
		permissions, err := permRepo.GetRolePermissions(ctx, roleID)
		require.NoError(t, err)
		assert.NotNil(t, permissions)
		assert.Empty(t, permissions)
	})
}
