package role_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	roleAggregate "github.com/basilex/promenade/internal/contexts/identity/role/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestRoleRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRoleRepository(testDB.DB)

		// Create with unique role name
		uuid := uuidv7.New().String()
		roleName := fmt.Sprintf("test_role_%s", uuid)
		r, err := roleAggregate.NewRole(roleName, "Test Role", "Test description")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, r))

		// Read by ID
		found, err := repo.GetByID(ctx, r.ID)
		require.NoError(t, err)
		assert.Equal(t, roleName, found.Name)
		assert.Equal(t, "Test Role", found.DisplayName)

		// Read by Name
		foundByName, err := repo.GetByName(ctx, roleName)
		require.NoError(t, err)
		assert.Equal(t, r.ID, foundByName.ID)

		// Update
		_ = found.UpdateDisplayName("Updated Role")
		found.UpdateDescription("Updated description")
		require.NoError(t, repo.Update(ctx, found))
		updated, _ := repo.GetByID(ctx, r.ID)
		assert.Equal(t, "Updated Role", updated.DisplayName)
		assert.Equal(t, "Updated description", updated.Description)

		// Delete
		require.NoError(t, repo.Delete(ctx, r.ID))
		_, err = repo.GetByID(ctx, r.ID)
		assert.Error(t, err)
	})
}

func TestRoleRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRoleRepository(testDB.DB)

		// Create 2 roles with unique names
		uuid := uuidv7.New().String()
		role1Name := fmt.Sprintf("role1_%s", uuid)
		role2Name := fmt.Sprintf("role2_%s", uuid)
		r1, _ := roleAggregate.NewRole(role1Name, "Role 1", "First")
		r2, _ := roleAggregate.NewRole(role2Name, "Role 2", "Second")
		require.NoError(t, repo.Create(ctx, r1))
		require.NoError(t, repo.Create(ctx, r2))

		// ExistsByName
		exists, err := repo.ExistsByName(ctx, role1Name)
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByName(ctx, "nonexistent")
		require.NoError(t, err)
		assert.False(t, exists)

		// List
		roles, total, err := repo.ListRoles(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
		assert.GreaterOrEqual(t, len(roles), 2)

		// GetUserRoles (requires user-role assignment)
		userID := uuidv7.New()
		_, err = tx.ExecContext(ctx, `INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
			userID, fmt.Sprintf("user_%s@test.com", userID), "hash", "active")
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx, `INSERT INTO identity_user_roles (user_id, role_id) VALUES ($1, $2)`, userID, r1.ID)
		require.NoError(t, err)

		userRoles, err := repo.GetUserRoles(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, userRoles, 1)
		assert.Equal(t, role1Name, userRoles[0].Name)
	})
}
