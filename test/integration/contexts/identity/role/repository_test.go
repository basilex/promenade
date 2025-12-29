package role_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestRoleRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("create role successfully", func(t *testing.T) {
		r, err := role.NewRole("manager", "Manager", "Manager role")
		require.NoError(t, err)

		err = repo.Create(ctx, r)
		require.NoError(t, err)

		// Verify it was created
		retrieved, err := repo.GetByID(ctx, r.ID)
		require.NoError(t, err)
		assert.Equal(t, r.ID, retrieved.ID)
		assert.Equal(t, "manager", retrieved.Name)
		assert.Equal(t, "Manager", retrieved.DisplayName)
		assert.False(t, retrieved.IsSystem)
	})

	t.Run("create system role", func(t *testing.T) {
		r, err := role.NewSystemRole("admin", "Administrator", "Admin role")
		require.NoError(t, err)

		err = repo.Create(ctx, r)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, r.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsSystem)
	})

	t.Run("create role with duplicate name fails", func(t *testing.T) {
		r1, _ := role.NewRole("duplicate", "Duplicate 1", "Test")
		r2, _ := role.NewRole("duplicate", "Duplicate 2", "Test")

		err := repo.Create(ctx, r1)
		require.NoError(t, err)

		// Second create with same name should fail
		err = repo.Create(ctx, r2)
		assert.Error(t, err)
	})
}

func TestRoleRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing role", func(t *testing.T) {
		r, _ := role.NewRole("test_role", "Test Role", "Test")
		require.NoError(t, repo.Create(ctx, r))

		retrieved, err := repo.GetByID(ctx, r.ID)
		require.NoError(t, err)
		assert.Equal(t, r.ID, retrieved.ID)
		assert.Equal(t, "test_role", retrieved.Name)
	})

	t.Run("get non-existent role returns error", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, role.ErrRoleNotFound)
	})
}

func TestRoleRepository_GetByName(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing role by name", func(t *testing.T) {
		r, _ := role.NewRole("unique_role", "Unique Role", "Test")
		require.NoError(t, repo.Create(ctx, r))

		retrieved, err := repo.GetByName(ctx, "unique_role")
		require.NoError(t, err)
		assert.Equal(t, r.ID, retrieved.ID)
		assert.Equal(t, "unique_role", retrieved.Name)
	})

	t.Run("get non-existent role by name returns error", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, role.ErrRoleNotFound)
	})
}

func TestRoleRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("update role successfully", func(t *testing.T) {
		r, _ := role.NewRole("updatable", "Updatable", "Original description")
		require.NoError(t, repo.Create(ctx, r))

		r.UpdateDescription("Updated description")
		err := r.UpdateDisplayName("Updated Display Name")
		require.NoError(t, err)

		err = repo.Update(ctx, r)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, r.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", retrieved.Description)
		assert.Equal(t, "Updated Display Name", retrieved.DisplayName)
	})
}

func TestRoleRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("delete regular role successfully", func(t *testing.T) {
		r, _ := role.NewRole("deletable", "Deletable", "Test")
		require.NoError(t, repo.Create(ctx, r))

		err := repo.Delete(ctx, r.ID)
		require.NoError(t, err)

		// Verify it's deleted (soft delete)
		_, err = repo.GetByID(ctx, r.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, role.ErrRoleNotFound)
	})

	t.Run("delete system role fails", func(t *testing.T) {
		r, _ := role.NewSystemRole("system_admin", "System Admin", "System")
		require.NoError(t, repo.Create(ctx, r))

		err := repo.Delete(ctx, r.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, role.ErrCannotDeleteSystem)
	})

	t.Run("delete non-existent role returns error", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		err := repo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, role.ErrRoleNotFound)
	})
}

func TestRoleRepository_ListRoles(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("list all roles", func(t *testing.T) {
		// Create several roles
		r1, _ := role.NewRole("role1", "Role 1", "Test 1")
		r2, _ := role.NewRole("role2", "Role 2", "Test 2")
		r3, _ := role.NewSystemRole("sysrole", "System Role", "System")

		require.NoError(t, repo.Create(ctx, r1))
		require.NoError(t, repo.Create(ctx, r2))
		require.NoError(t, repo.Create(ctx, r3))

		roles, total, err := repo.ListRoles(ctx, 20, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(roles), 3)
		assert.GreaterOrEqual(t, total, 3)
	})
}

func TestRoleRepository_ExistsByName(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("existing role returns true", func(t *testing.T) {
		r, _ := role.NewRole("exists", "Exists", "Test")
		require.NoError(t, repo.Create(ctx, r))

		exists, err := repo.ExistsByName(ctx, "exists")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent role returns false", func(t *testing.T) {
		exists, err := repo.ExistsByName(ctx, "does_not_exist")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestRoleRepository_GetUserRoles(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	roleRepo := postgres.NewRoleRepository(db.DB)
	ctx := context.Background()

	t.Run("get user roles successfully", func(t *testing.T) {
		// Create roles
		r1, _ := role.NewRole("user_role", "User Role", "Test")
		r2, _ := role.NewRole("admin_role", "Admin Role", "Test")
		require.NoError(t, roleRepo.Create(ctx, r1))
		require.NoError(t, roleRepo.Create(ctx, r2))

		// Create user and assign roles (this requires user repo and assignment logic)
		// For now, test with empty result
		userID := uuidv7.New()
		roles, err := roleRepo.GetUserRoles(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, roles)
		// Without role assignments, should be empty
		assert.Empty(t, roles)
	})
}
