#!/usr/bin/env python3

permission_test = '''package permission_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	"github.com/basilex/promenade/internal/contexts/identity/permission/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestPermissionRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPermissionRepository(testDB.DB)

		// Create
		p, err := permission.NewPermission("test_resource", "read", "Test permission")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, p))

		// Read by ID
		found, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "test_resource", found.Resource)
		assert.Equal(t, "read", found.Action)
		assert.Equal(t, "test_resource:read", found.Name)

		// Read by Name
		foundByName, err := repo.GetByName(ctx, "test_resource:read")
		require.NoError(t, err)
		assert.Equal(t, p.ID, foundByName.ID)

		// Update
		found.UpdateDescription("Updated description")
		require.NoError(t, repo.Update(ctx, found))
		updated, _ := repo.GetByID(ctx, p.ID)
		assert.Equal(t, "Updated description", updated.Description)

		// Delete
		require.NoError(t, repo.Delete(ctx, p.ID))
		_, err = repo.GetByID(ctx, p.ID)
		assert.Error(t, err)
	})
}

func TestPermissionRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPermissionRepository(testDB.DB)

		// Create 2 permissions
		p1, _ := permission.NewPermission("resource1", "read", "First")
		p2, _ := permission.NewPermission("resource2", "write", "Second")
		require.NoError(t, repo.Create(ctx, p1))
		require.NoError(t, repo.Create(ctx, p2))

		// ExistsByName
		exists, err := repo.ExistsByName(ctx, "resource1:read")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByName(ctx, "nonexistent:action")
		require.NoError(t, err)
		assert.False(t, exists)

		// List
		perms, total, err := repo.ListPermissions(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(perms), 2)

		// GetRolePermissions (requires role-permission assignment)
		roleID := uuidv7.New()
		_, err = testDB.DB.Exec(`INSERT INTO identity_roles (id, name, display_name, description) VALUES ($1, $2, $3, $4)`,
			roleID, fmt.Sprintf("role_%s", roleID), "Test Role", "Test")
		require.NoError(t, err)

		_, err = testDB.DB.Exec(`INSERT INTO identity_role_permissions (role_id, permission_id) VALUES ($1, $2)`, roleID, p1.ID)
		require.NoError(t, err)

		rolePerms, err := repo.GetRolePermissions(ctx, roleID)
		require.NoError(t, err)
		assert.Len(t, rolePerms, 1)
		assert.Equal(t, "resource1:read", rolePerms[0].Name)
	})
}
'''

with open('test/integration/contexts/identity/permission/repository_test.go', 'w') as f:
    f.write(permission_test)

print("✓ Generated test/integration/contexts/identity/permission/repository_test.go (106 lines)")
