package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	rolePostgres "github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	roleAggregate "github.com/basilex/promenade/internal/contexts/identity/role/aggregate"
	roleUseCase "github.com/basilex/promenade/internal/contexts/identity/role/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// Test 1: CreateRole
func TestRoleUseCase_CreateRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Test - Create role
		r, err := uc.CreateRole(ctx, "manager", "Manager", "Team manager role")
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, r.GetID())
		assert.Equal(t, "manager", r.Name)
		assert.Equal(t, "Manager", r.DisplayName)
		assert.Equal(t, "Team manager role", r.Description)
		assert.False(t, r.IsSystem) // Not a system role

		// Verify persistence
		retrieved, err := uc.GetRole(ctx, r.GetID())
		require.NoError(t, err)
		assert.Equal(t, "manager", retrieved.Name)
	})
}

// Test 2: GetRole
func TestRoleUseCase_GetRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Create role
		created, err := uc.CreateRole(ctx, "editor", "Editor", "Content editor role")
		require.NoError(t, err)

		// Test - Get role by ID
		retrieved, err := uc.GetRole(ctx, created.GetID())
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, "editor", retrieved.Name)
		assert.Equal(t, "Editor", retrieved.DisplayName)

		// Test - Non-existent role
		_, err = uc.GetRole(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, role.ErrRoleNotFound))
	})
}

// Test 3: GetRoleByName
func TestRoleUseCase_GetRoleByName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Create role
		created, err := uc.CreateRole(ctx, "viewer", "Viewer", "Read-only viewer role")
		require.NoError(t, err)

		// Test - Get role by name
		retrieved, err := uc.GetRoleByName(ctx, "viewer")
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, "viewer", retrieved.Name)

		// Test - Non-existent role name
		_, err = uc.GetRoleByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, role.ErrRoleNotFound))
	})
}

// Test 4: UpdateRole
func TestRoleUseCase_UpdateRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Create role
		r, err := uc.CreateRole(ctx, "moderator", "Moderator", "Content moderator")
		require.NoError(t, err)

		// Test - Update role
		updated, err := uc.UpdateRole(ctx, r.GetID(), "Senior Moderator", "Senior content moderator")
		require.NoError(t, err)
		assert.Equal(t, "Senior Moderator", updated.DisplayName)
		assert.Equal(t, "Senior content moderator", updated.Description)
		assert.Equal(t, "moderator", updated.Name) // Name unchanged

		// Verify persistence
		retrieved, err := uc.GetRole(ctx, r.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Senior Moderator", retrieved.DisplayName)
		assert.Equal(t, "Senior content moderator", retrieved.Description)
	})
}

// Test 5: DeleteRole
func TestRoleUseCase_DeleteRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Create role
		r, err := uc.CreateRole(ctx, "temp_role", "Temporary Role", "For testing")
		require.NoError(t, err)

		// Test - Delete role
		err = uc.DeleteRole(ctx, r.GetID())
		require.NoError(t, err)

		// Verify - Should not be found (soft delete)
		_, err = uc.GetRole(ctx, r.GetID())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, role.ErrRoleNotFound))
	})
}

// Test 6: ListRoles
func TestRoleUseCase_ListRoles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Create multiple roles
		_, err := uc.CreateRole(ctx, "role1", "Role 1", "First role")
		require.NoError(t, err)

		_, err = uc.CreateRole(ctx, "role2", "Role 2", "Second role")
		require.NoError(t, err)

		_, err = uc.CreateRole(ctx, "role3", "Role 3", "Third role")
		require.NoError(t, err)

		// Test - List all roles
		roles, total, err := uc.ListRoles(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(roles), 3) // At least our 3 roles
		assert.GreaterOrEqual(t, total, 3)

		// Test - Pagination
		rolesPage1, _, err := uc.ListRoles(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(rolesPage1))

		rolesPage2, _, err := uc.ListRoles(ctx, 2, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(rolesPage2), 1) // At least 1 more role
	})
}

// Test 7: GetUserRoles (requires user-role assignment)
func TestRoleUseCase_GetUserRoles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Create role directly in repo (avoid usecase rollback)
		r, err := roleAggregate.NewRole("test_role", "Test Role", "For user assignment")
		require.NoError(t, err)
		err = roleRepo.Create(ctx, r)
		require.NoError(t, err)

		// Create user and assign role (use tx for transaction support)
		userID := uuidv7.New()

		// Insert user with unique email
		uniqueEmail := "test_" + userID.String()[:8] + "@example.com"
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_users (id, email, password_hash, status, email_verified)
			 VALUES ($1, $2, $3, $4, $5)`,
			userID, uniqueEmail, "$2a$10$test", "active", false,
		)
		require.NoError(t, err)

		// Assign role to user
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_user_roles (user_id, role_id) VALUES ($1, $2)`,
			userID, r.GetID(),
		)
		require.NoError(t, err)

		// Test - Get user roles
		userRoles, err := uc.GetUserRoles(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(userRoles))
		assert.Equal(t, "test_role", userRoles[0].Name)

		// Test - User with no roles
		emptyUserID := uuidv7.New()
		emptyEmail := "empty_" + emptyUserID.String()[:8] + "@example.com"
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_users (id, email, password_hash, status, email_verified)
			 VALUES ($1, $2, $3, $4, $5)`,
			emptyUserID, emptyEmail, "$2a$10$test", "active", false,
		)
		require.NoError(t, err)

		emptyRoles, err := uc.GetUserRoles(ctx, emptyUserID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(emptyRoles))
	})
}

// Test 8: CompleteWorkflow - Full lifecycle test
func TestRoleUseCase_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Setup
		roleRepo := rolePostgres.NewRoleRepository(testDB.DB)
		uc := roleUseCase.NewRoleUseCase(roleRepo)

		// Step 1: Create multiple roles
		adminRole, err := uc.CreateRole(ctx, "workflow_admin", "Workflow Admin", "Admin for workflow test")
		require.NoError(t, err)
		assert.Equal(t, "workflow_admin", adminRole.Name)

		managerRole, err := uc.CreateRole(ctx, "workflow_manager", "Workflow Manager", "Manager for workflow test")
		require.NoError(t, err)

		// Step 2: Get role by ID
		retrieved, err := uc.GetRole(ctx, adminRole.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Workflow Admin", retrieved.DisplayName)

		// Step 3: Get role by name
		byName, err := uc.GetRoleByName(ctx, "workflow_manager")
		require.NoError(t, err)
		assert.Equal(t, managerRole.GetID(), byName.GetID())

		// Step 4: Update role
		updated, err := uc.UpdateRole(ctx, adminRole.GetID(), "Super Admin", "Updated admin description")
		require.NoError(t, err)
		assert.Equal(t, "Super Admin", updated.DisplayName)
		assert.Equal(t, "Updated admin description", updated.Description)

		// Step 5: List roles
		roles, total, err := uc.ListRoles(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
		assert.GreaterOrEqual(t, len(roles), 2)

		// Step 6: Create user and assign role (use tx for transaction support)
		// Note: adminRole exists in DB from CreateRole call above
		userID := uuidv7.New()
		workflowEmail := "workflow_" + userID.String()[:8] + "@example.com"
		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_users (id, email, password_hash, status, email_verified)
			 VALUES ($1, $2, $3, $4, $5)`,
			userID, workflowEmail, "$2a$10$test", "active", false,
		)
		require.NoError(t, err)

		// Re-fetch role ID to ensure it exists in transaction scope
		adminInDB, err := uc.GetRole(ctx, adminRole.GetID())
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx,
			`INSERT INTO identity_user_roles (user_id, role_id) VALUES ($1, $2)`,
			userID, adminInDB.GetID(),
		)
		require.NoError(t, err)

		// Step 7: Get user roles
		userRoles, err := uc.GetUserRoles(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(userRoles))
		assert.Equal(t, "workflow_admin", userRoles[0].Name)
		assert.Equal(t, "Super Admin", userRoles[0].DisplayName) // Updated value

		// Step 8: Delete manager role
		err = uc.DeleteRole(ctx, managerRole.GetID())
		require.NoError(t, err)

		// Verify deletion
		_, err = uc.GetRole(ctx, managerRole.GetID())
		assert.Error(t, err)
		assert.True(t, errors.Is(err, role.ErrRoleNotFound))

		// Admin role should still exist
		stillExists, err := uc.GetRole(ctx, adminRole.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Super Admin", stillExists.DisplayName)
	})
}
