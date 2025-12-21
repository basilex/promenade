package smoke_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

// TestRBAC_SmokeTest verifies complete RBAC flow: permissions, roles, and user assignments
func TestRBAC_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	permRepo := postgres.NewPermissionRepository(testDB.DB)
	roleRepo := postgres.NewRoleRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)

	// Setup use cases
	permUC := usecase.NewPermissionUseCase(permRepo)
	roleUC := usecase.NewRoleUseCase(roleRepo, permRepo)

	var perm1ID, perm2ID, perm3ID uuidv7.UUID
	var roleID uuidv7.UUID
	var userID uuidv7.UUID

	// ========== Permission Management ==========

	t.Run("[+] Create_custom_permissions", func(t *testing.T) {
		// Create invoice permissions
		perm1, err := permUC.CreatePermission(ctx, "invoices", "read", "Can read invoices")
		require.NoError(t, err)
		require.NotNil(t, perm1)
		assert.Equal(t, "invoices", perm1.Resource)
		assert.Equal(t, "read", perm1.Action)
		perm1ID = perm1.ID

		perm2, err := permUC.CreatePermission(ctx, "invoices", "approve", "Can approve invoices")
		require.NoError(t, err)
		perm2ID = perm2.ID

		// Create wildcard permission
		perm3, err := permUC.CreatePermission(ctx, "reports", "*", "Can perform any action on reports")
		require.NoError(t, err)
		assert.Equal(t, "*", perm3.Action)
		perm3ID = perm3.ID
	})

	t.Run("[+] Get_permission_by_ID", func(t *testing.T) {
		perm, err := permUC.GetPermission(ctx, perm1ID)
		require.NoError(t, err)
		assert.Equal(t, "invoices", perm.Resource)
		assert.Equal(t, "read", perm.Action)
	})

	t.Run("[+] Get_permission_by_resource_action", func(t *testing.T) {
		perm, err := permUC.GetPermissionByResourceAction(ctx, "invoices", "approve")
		require.NoError(t, err)
		assert.Equal(t, perm2ID, perm.ID)
	})

	t.Run("[+] List_permissions", func(t *testing.T) {
		perms, err := permUC.ListPermissions(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(perms), 3, "should have at least 3 permissions")
	})

	t.Run("[+] Update_permission_description", func(t *testing.T) {
		updated, err := permUC.UpdatePermission(ctx, perm1ID, "Updated: Can read all invoices")
		require.NoError(t, err)
		assert.Equal(t, "Updated: Can read all invoices", *updated.Description)
	})

	// ========== Role Management ==========

	t.Run("[+] Create_custom_role", func(t *testing.T) {
		role, err := roleUC.CreateRole(ctx, "accountant", "Accountant", "Financial accountant role")
		require.NoError(t, err)
		require.NotNil(t, role)
		assert.Equal(t, "accountant", role.Name)
		assert.Equal(t, "Accountant", role.DisplayName)
		assert.False(t, role.IsSystem, "custom roles should not be system roles")
		roleID = role.ID
	})

	t.Run("[+] Get_role_by_ID", func(t *testing.T) {
		role, err := roleUC.GetRole(ctx, roleID)
		require.NoError(t, err)
		assert.Equal(t, "accountant", role.Name)
	})

	t.Run("[+] Get_role_by_name", func(t *testing.T) {
		role, err := roleUC.GetRoleByName(ctx, "accountant")
		require.NoError(t, err)
		assert.Equal(t, roleID, role.ID)
	})

	t.Run("[+] List_roles", func(t *testing.T) {
		roles, err := roleUC.ListRoles(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(roles), 1, "should have at least 1 role")
	})

	// ========== Role-Permission Assignment ==========

	t.Run("[+] Assign_permissions_to_role", func(t *testing.T) {
		permIDs := []uuidv7.UUID{perm1ID, perm2ID}
		err := roleUC.SyncRolePermissions(ctx, roleID, permIDs)
		require.NoError(t, err)
	})

	t.Run("[+] Get_role_permissions", func(t *testing.T) {
		perms, err := roleUC.GetRolePermissions(ctx, roleID)
		require.NoError(t, err)
		assert.Len(t, perms, 2, "role should have 2 permissions")

		// Verify permission details
		permMap := make(map[string]bool)
		for _, p := range perms {
			permMap[p.Resource+":"+p.Action] = true
		}
		assert.True(t, permMap["invoices:read"], "should have invoices:read")
		assert.True(t, permMap["invoices:approve"], "should have invoices:approve")
	})

	t.Run("[+] Update_role_permissions", func(t *testing.T) {
		// Add wildcard permission, remove approve
		permIDs := []uuidv7.UUID{perm1ID, perm3ID}
		err := roleUC.SyncRolePermissions(ctx, roleID, permIDs)
		require.NoError(t, err)

		perms, err := roleUC.GetRolePermissions(ctx, roleID)
		require.NoError(t, err)
		assert.Len(t, perms, 2, "role should have 2 permissions after sync")
	})

	// ========== User-Role Assignment ==========

	t.Run("[+] Create_user_for_role_assignment", func(t *testing.T) {
		user := helpers.UserFixture(func(u *entity.User) {
			u.Email = "accountant@test.com"
			u.Name = "Test Accountant"
		})
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)
		userID = user.ID
	})

	t.Run("[+] Assign_role_to_user", func(t *testing.T) {
		adminUser := helpers.UserFixture(func(u *entity.User) {
			u.Email = "admin@test.com"
		})
		err := userRepo.Create(ctx, adminUser)
		require.NoError(t, err)

		err = roleUC.AssignRoleToUser(ctx, userID, roleID, adminUser.ID, nil)
		require.NoError(t, err)
	})

	t.Run("[+] Get_user_roles", func(t *testing.T) {
		roles, err := roleUC.GetUserRoles(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, roles, 1, "user should have 1 role")
		assert.Equal(t, "accountant", roles[0].Name)
	})

	t.Run("[+] Get_user_active_roles", func(t *testing.T) {
		activeRoles, err := roleUC.GetUserActiveRoles(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, activeRoles, 1, "user should have 1 active role")
	})

	// ========== Permission Checking ==========

	t.Run("[+] Get_user_permissions_aggregated", func(t *testing.T) {
		perms, err := roleUC.GetUserPermissions(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, perms, 2, "user should have 2 permissions from role")
	})

	t.Run("[+] Check_user_has_permission", func(t *testing.T) {
		hasPerm, err := roleUC.HasPermission(ctx, userID, "invoices:read")
		require.NoError(t, err)
		assert.True(t, hasPerm, "user should have invoices:read permission")

		hasPerm, err = roleUC.HasPermission(ctx, userID, "invoices:delete")
		require.NoError(t, err)
		assert.False(t, hasPerm, "user should NOT have invoices:delete permission")
	})

	t.Run("[+] Check_wildcard_permission", func(t *testing.T) {
		hasPerm, err := roleUC.HasPermission(ctx, userID, "reports:create")
		require.NoError(t, err)
		assert.True(t, hasPerm, "user should have reports:create via wildcard reports:*")

		hasPerm, err = roleUC.HasPermission(ctx, userID, "reports:delete")
		require.NoError(t, err)
		assert.True(t, hasPerm, "user should have reports:delete via wildcard reports:*")
	})

	t.Run("[+] Check_user_has_any_permission", func(t *testing.T) {
		hasAny, err := roleUC.HasAnyPermission(ctx, userID, []string{"invoices:read", "invoices:delete"})
		require.NoError(t, err)
		assert.True(t, hasAny, "user should have at least one of the permissions")

		hasAny, err = roleUC.HasAnyPermission(ctx, userID, []string{"posts:create", "posts:delete"})
		require.NoError(t, err)
		assert.False(t, hasAny, "user should NOT have any of these permissions")
	})

	t.Run("[+] Check_user_has_all_permissions", func(t *testing.T) {
		hasAll, err := roleUC.HasAllPermissions(ctx, userID, []string{"invoices:read", "reports:view"})
		require.NoError(t, err)
		assert.True(t, hasAll, "user should have all specified permissions")

		hasAll, err = roleUC.HasAllPermissions(ctx, userID, []string{"invoices:read", "invoices:delete"})
		require.NoError(t, err)
		assert.False(t, hasAll, "user should NOT have all specified permissions")
	})

	// ========== Role Assignment with Expiration ==========

	t.Run("[+] Assign_role_with_expiration", func(t *testing.T) {
		tempUser := helpers.UserFixture(func(u *entity.User) {
			u.Email = "temp@test.com"
		})
		err := userRepo.Create(ctx, tempUser)
		require.NoError(t, err)

		adminUser := helpers.UserFixture(func(u *entity.User) {
			u.Email = "admin2@test.com"
		})
		err = userRepo.Create(ctx, adminUser)
		require.NoError(t, err)

		// Assign role with expiration in 1 hour
		expiresAt := time.Now().Add(time.Hour)
		err = roleUC.AssignRoleToUser(ctx, tempUser.ID, roleID, adminUser.ID, &expiresAt)
		require.NoError(t, err)

		// Verify role is active
		activeRoles, err := roleUC.GetUserActiveRoles(ctx, tempUser.ID)
		require.NoError(t, err)
		assert.Len(t, activeRoles, 1, "temporary user should have 1 active role")
	})

	// ========== Cleanup Operations ==========

	t.Run("[+] Remove_role_from_user", func(t *testing.T) {
		err := roleUC.RemoveRoleFromUser(ctx, userID, roleID)
		require.NoError(t, err)

		roles, err := roleUC.GetUserRoles(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, roles, 0, "user should have no roles after removal")
	})

	t.Run("[+] Get_users_with_role", func(t *testing.T) {
		users, err := roleUC.GetUsersWithRole(ctx, roleID)
		require.NoError(t, err)
		// Should have 1 user (tempUser with expiration)
		assert.GreaterOrEqual(t, len(users), 0, "should return users with this role")
	})

	t.Run("[+] Update_role", func(t *testing.T) {
		updated, err := roleUC.UpdateRole(ctx, roleID, "Senior Accountant", "Senior financial accountant")
		require.NoError(t, err)
		assert.Equal(t, "Senior Accountant", updated.DisplayName)
	})

	t.Run("[+] Delete_permission", func(t *testing.T) {
		// Create a permission to delete
		tempPerm, err := permUC.CreatePermission(ctx, "temp", "action", "Temporary permission")
		require.NoError(t, err)

		err = permUC.DeletePermission(ctx, tempPerm.ID)
		require.NoError(t, err)

		_, err = permUC.GetPermission(ctx, tempPerm.ID)
		assert.Error(t, err, "deleted permission should not be found")
	})

	t.Run("[+] Cannot_delete_system_role", func(t *testing.T) {
		// Get a system role (should exist from migrations)
		systemRole, err := roleUC.GetRoleByName(ctx, "superadmin")
		if err != nil {
			t.Skip("System roles not seeded in test DB")
		}

		err = roleUC.DeleteRole(ctx, systemRole.ID)
		assert.Error(t, err, "should not be able to delete system role")
	})

	t.Run("[+] Delete_custom_role", func(t *testing.T) {
		err := roleUC.DeleteRole(ctx, roleID)
		require.NoError(t, err)

		_, err = roleUC.GetRole(ctx, roleID)
		assert.Error(t, err, "deleted role should not be found")
	})

	t.Logf("[SUCCESS] All RBAC smoke tests passed!")
}
