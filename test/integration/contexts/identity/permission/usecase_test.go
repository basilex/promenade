package permission_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	permissionPostgres "github.com/basilex/promenade/internal/contexts/identity/permission/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestPermissionUseCase_CreatePermission tests creating a new permission
func TestPermissionUseCase_CreatePermission(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Test - Create permission
		perm, err := uc.CreatePermission(ctx, "users", "create", "Create users")
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, perm.GetID())
		assert.Equal(t, "users:create", perm.Name)
		assert.Equal(t, "users", perm.Resource)
		assert.Equal(t, "create", perm.Action)
		assert.Equal(t, "Create users", perm.Description)

		// Test - Duplicate name should fail
		_, err = uc.CreatePermission(ctx, "users", "create", "Duplicate")
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNameExists)
	})
}

// TestPermissionUseCase_GetPermission tests retrieving a permission by ID
func TestPermissionUseCase_GetPermission(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Create permission
		created, err := uc.CreatePermission(ctx, "users", "read", "Read users")
		require.NoError(t, err)

		// Test - Get by ID
		retrieved, err := uc.GetPermission(ctx, created.GetID())
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, "users:read", retrieved.Name)

		// Test - Non-existent ID
		_, err = uc.GetPermission(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

// TestPermissionUseCase_GetPermissionByName tests retrieving a permission by name
func TestPermissionUseCase_GetPermissionByName(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Create permission
		created, err := uc.CreatePermission(ctx, "users", "update", "Update users")
		require.NoError(t, err)

		// Test - Get by name
		retrieved, err := uc.GetPermissionByName(ctx, "users:update")
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, "users:update", retrieved.Name)

		// Test - Non-existent name
		_, err = uc.GetPermissionByName(ctx, "nonexistent:action")
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

// TestPermissionUseCase_UpdatePermission tests updating an existing permission
func TestPermissionUseCase_UpdatePermission(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Create permission
		created, err := uc.CreatePermission(ctx, "users", "delete", "Delete users")
		require.NoError(t, err)

		// Test - Update description
		updated, err := uc.UpdatePermission(ctx, created.GetID(), "Soft delete users")
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), updated.GetID())
		assert.Equal(t, "Soft delete users", updated.Description)

		// Verify persistence
		retrieved, err := uc.GetPermission(ctx, created.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Soft delete users", retrieved.Description)

		// Test - Non-existent permission
		_, err = uc.UpdatePermission(ctx, uuidv7.New(), "Should fail")
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

// TestPermissionUseCase_DeletePermission tests soft deleting a permission
func TestPermissionUseCase_DeletePermission(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Create permission
		created, err := uc.CreatePermission(ctx, "posts", "create", "Create posts")
		require.NoError(t, err)

		// Test - Delete permission
		err = uc.DeletePermission(ctx, created.GetID())
		require.NoError(t, err)

		// Verify permission is soft deleted (not found)
		_, err = uc.GetPermission(ctx, created.GetID())
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)

		// Test - Delete non-existent permission
		err = uc.DeletePermission(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)
	})
}

// TestPermissionUseCase_ListPermissions tests listing permissions with pagination
func TestPermissionUseCase_ListPermissions(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Create multiple permissions
		_, err := uc.CreatePermission(ctx, "users", "create", "Create users")
		require.NoError(t, err)
		_, err = uc.CreatePermission(ctx, "users", "read", "Read users")
		require.NoError(t, err)
		_, err = uc.CreatePermission(ctx, "users", "update", "Update users")
		require.NoError(t, err)

		// Test - List all permissions
		perms, total, err := uc.ListPermissions(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(perms), 3)
		assert.GreaterOrEqual(t, total, 3)

		// Test - Pagination
		perms, total, err = uc.ListPermissions(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(perms))
		assert.GreaterOrEqual(t, total, 3)

		// Test - Empty result with high offset
		perms, total, err = uc.ListPermissions(ctx, 10, 1000)
		require.NoError(t, err)
		assert.Equal(t, 0, len(perms))
		assert.GreaterOrEqual(t, total, 3)
	})
}

// TestPermissionUseCase_GetRolePermissions tests retrieving permissions assigned to a role
func TestPermissionUseCase_GetRolePermissions(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Create permissions
		perm1, err := uc.CreatePermission(ctx, "users", "create", "Create users")
		require.NoError(t, err)
		perm2, err := uc.CreatePermission(ctx, "users", "read", "Read users")
		require.NoError(t, err)

		// Create role (use tx for transaction support)
		roleID := uuidv7.New()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_roles (id, name, display_name, is_system)
			 VALUES ($1, $2, $3, $4)`,
			roleID, "test_role", "Test Role", false,
		)
		require.NoError(t, err)

		// Assign permissions to role
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, perm1.GetID(),
		)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, perm2.GetID(),
		)
		require.NoError(t, err)

		// Test - Get role permissions
		rolePerms, err := uc.GetRolePermissions(ctx, roleID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(rolePerms))

		// Test - Role with no permissions
		emptyRoleID := uuidv7.New()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_roles (id, name, display_name, is_system)
			 VALUES ($1, $2, $3, $4)`,
			emptyRoleID, "empty_role", "Empty Role", false,
		)
		require.NoError(t, err)

		rolePerms, err = uc.GetRolePermissions(ctx, emptyRoleID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(rolePerms))
	})
}

// TestPermissionUseCase_CompleteWorkflow tests complete permission management workflow
func TestPermissionUseCase_CompleteWorkflow(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		permRepo := permissionPostgres.NewPermissionRepository(testDB.DB)
		uc := permission.NewUseCase(permRepo)

		// Step 1: Create permissions for CRUD operations
		createPerm, err := uc.CreatePermission(ctx, "posts", "create", "Create posts")
		require.NoError(t, err)

		readPerm, err := uc.CreatePermission(ctx, "posts", "read", "Read posts")
		require.NoError(t, err)

		updatePerm, err := uc.CreatePermission(ctx, "posts", "update", "Update posts")
		require.NoError(t, err)

		deletePerm, err := uc.CreatePermission(ctx, "posts", "delete", "Delete posts")
		require.NoError(t, err)

		// Step 2: List all permissions
		perms, total, err := uc.ListPermissions(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(perms), 4)
		assert.GreaterOrEqual(t, total, 4)

		// Step 3: Get permission by name
		retrieved, err := uc.GetPermissionByName(ctx, "posts:read")
		require.NoError(t, err)
		assert.Equal(t, readPerm.GetID(), retrieved.GetID())

		// Step 4: Update permission description
		updated, err := uc.UpdatePermission(ctx, updatePerm.GetID(), "Modify existing posts")
		require.NoError(t, err)
		assert.Equal(t, "Modify existing posts", updated.Description)

		// Step 5: Create role and assign permissions
		roleID := uuidv7.New()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_roles (id, name, display_name, is_system)
			 VALUES ($1, $2, $3, $4)`,
			roleID, "workflow_admin", "Workflow Admin", false,
		)
		require.NoError(t, err)

		// Assign all CRUD permissions to role
		for _, perm := range []*permission.Permission{createPerm, readPerm, updatePerm, deletePerm} {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO identity_role_permissions (role_id, permission_id) VALUES ($1, $2)`,
				roleID, perm.GetID(),
			)
			require.NoError(t, err)
		}

		// Step 6: Get role permissions
		rolePerms, err := uc.GetRolePermissions(ctx, roleID)
		require.NoError(t, err)
		assert.Equal(t, 4, len(rolePerms))

		// Step 7: Soft delete one permission
		err = uc.DeletePermission(ctx, deletePerm.GetID())
		require.NoError(t, err)

		// Step 8: Verify deleted permission not found
		_, err = uc.GetPermission(ctx, deletePerm.GetID())
		assert.Error(t, err)
		assert.ErrorIs(t, err, permission.ErrPermissionNotFound)

		// Step 9: List permissions (deleted one should not appear)
		perms, total, err = uc.ListPermissions(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(perms), 3) // At least 3 active permissions
	})
}
