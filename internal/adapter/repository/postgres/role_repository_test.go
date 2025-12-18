package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

func TestRoleRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewRoleRepository(testDB.DB)
	ctx := context.Background()

	t.Run("creates role successfully", func(t *testing.T) {
		role := helpers.RoleFixture("editor")
		err := repo.Create(ctx, role)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Equal(t, role.Name, retrieved.Name)
		assert.Equal(t, role.DisplayName, retrieved.DisplayName)
	})

	t.Run("creates system role", func(t *testing.T) {
		role := helpers.RoleFixture("system_admin", func(r *entity.Role) {
			r.IsSystem = true
		})
		err := repo.Create(ctx, role)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, role.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsSystem)
	})
}

func TestRoleRepository_GetByName(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewRoleRepository(testDB.DB)
	ctx := context.Background()

	t.Run("finds role by name", func(t *testing.T) {
		role := helpers.RoleFixture("editor")
		err := repo.Create(ctx, role)
		require.NoError(t, err)

		retrieved, err := repo.GetByName(ctx, "editor")
		require.NoError(t, err)
		assert.Equal(t, role.ID, retrieved.ID)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		_, err := repo.GetByName(ctx, "nonexistent")
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestRoleRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewRoleRepository(testDB.DB)
	ctx := context.Background()

	t.Run("updates role successfully", func(t *testing.T) {
		role := helpers.RoleFixture("editor")
		err := repo.Create(ctx, role)
		require.NoError(t, err)

		role.DisplayName = "Content Editor"
		newDesc := "Updated description"
		role.Description = &newDesc
		err = repo.Update(ctx, role)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Equal(t, "Content Editor", retrieved.DisplayName)
		assert.Equal(t, newDesc, *retrieved.Description)
	})

	t.Run("returns error when role not found", func(t *testing.T) {
		role := helpers.RoleFixture("nonexistent")
		role.ID = uuidv7.New()
		err := repo.Update(ctx, role)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestRoleRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewRoleRepository(testDB.DB)
	ctx := context.Background()

	t.Run("deletes non-system role successfully", func(t *testing.T) {
		role := helpers.RoleFixture("editor")
		err := repo.Create(ctx, role)
		require.NoError(t, err)

		err = repo.Delete(ctx, role.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, role.ID)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})

	t.Run("returns error when role not found", func(t *testing.T) {
		err := repo.Delete(ctx, uuidv7.New())
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})

	t.Run("prevents deletion of system role", func(t *testing.T) {
		role := helpers.RoleFixture("system_admin", func(r *entity.Role) {
			r.IsSystem = true
		})
		err := repo.Create(ctx, role)
		require.NoError(t, err)

		err = repo.Delete(ctx, role.ID)
		assert.ErrorIs(t, err, entity.ErrNotFound)

		// Verify role still exists
		_, err = repo.GetByID(ctx, role.ID)
		assert.NoError(t, err)
	})
}

func TestRoleRepository_Permissions(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	roleRepo := postgres.NewRoleRepository(testDB.DB)
	permRepo := postgres.NewPermissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("adds permission to role", func(t *testing.T) {
		role := helpers.RoleFixture("editor")
		perm := helpers.PermissionFixture("posts", "create")

		require.NoError(t, roleRepo.Create(ctx, role))
		require.NoError(t, permRepo.Create(ctx, perm))

		err := roleRepo.AddPermission(ctx, role.ID, perm.ID)
		require.NoError(t, err)

		permissions, err := roleRepo.GetPermissions(ctx, role.ID)
		require.NoError(t, err)
		assert.Len(t, permissions, 1)
		assert.Equal(t, perm.ID, permissions[0].ID)
	})

	t.Run("removes permission from role", func(t *testing.T) {
		role := helpers.RoleFixture("editor1")
		perm := helpers.PermissionFixture("posts", "delete")

		require.NoError(t, roleRepo.Create(ctx, role))
		require.NoError(t, permRepo.Create(ctx, perm))
		require.NoError(t, roleRepo.AddPermission(ctx, role.ID, perm.ID))

		err := roleRepo.RemovePermission(ctx, role.ID, perm.ID)
		require.NoError(t, err)

		permissions, err := roleRepo.GetPermissions(ctx, role.ID)
		require.NoError(t, err)
		assert.Empty(t, permissions)
	})

	t.Run("syncs permissions", func(t *testing.T) {
		role := helpers.RoleFixture("editor2")
		perm1 := helpers.PermissionFixture("articles", "create")
		perm2 := helpers.PermissionFixture("articles", "read")
		perm3 := helpers.PermissionFixture("articles", "update")

		require.NoError(t, roleRepo.Create(ctx, role))
		require.NoError(t, permRepo.Create(ctx, perm1))
		require.NoError(t, permRepo.Create(ctx, perm2))
		require.NoError(t, permRepo.Create(ctx, perm3))

		// Initially add perm1 and perm2
		require.NoError(t, roleRepo.AddPermission(ctx, role.ID, perm1.ID))
		require.NoError(t, roleRepo.AddPermission(ctx, role.ID, perm2.ID))

		// Sync to perm2 and perm3 (removes perm1, adds perm3)
		err := roleRepo.SyncPermissions(ctx, role.ID, []uuidv7.UUID{perm2.ID, perm3.ID})
		require.NoError(t, err)

		permissions, err := roleRepo.GetPermissions(ctx, role.ID)
		require.NoError(t, err)
		assert.Len(t, permissions, 2)

		// Verify correct permissions
		ids := make(map[uuidv7.UUID]bool)
		for _, p := range permissions {
			ids[p.ID] = true
		}
		assert.True(t, ids[perm2.ID])
		assert.True(t, ids[perm3.ID])
		assert.False(t, ids[perm1.ID])
	})

	t.Run("syncs to empty permissions", func(t *testing.T) {
		role := helpers.RoleFixture("editor3")
		perm := helpers.PermissionFixture("documents", "create")

		require.NoError(t, roleRepo.Create(ctx, role))
		require.NoError(t, permRepo.Create(ctx, perm))
		require.NoError(t, roleRepo.AddPermission(ctx, role.ID, perm.ID))

		err := roleRepo.SyncPermissions(ctx, role.ID, []uuidv7.UUID{})
		require.NoError(t, err)

		permissions, err := roleRepo.GetPermissions(ctx, role.ID)
		require.NoError(t, err)
		assert.Empty(t, permissions)
	})
}

func TestRoleRepository_UserRoles(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	roleRepo := postgres.NewRoleRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("assigns role to user", func(t *testing.T) {
		user := helpers.UserFixture()
		role := helpers.RoleFixture("editor")
		assigner := helpers.UserFixture() // Create assigner user

		require.NoError(t, userRepo.Create(ctx, user))
		require.NoError(t, userRepo.Create(ctx, assigner))
		require.NoError(t, roleRepo.Create(ctx, role))

		userRole := helpers.UserRoleFixture(user.ID, role.ID, func(ur *entity.UserRole) {
			ur.AssignedBy = &assigner.ID
		})

		err := roleRepo.AssignToUser(ctx, userRole)
		require.NoError(t, err)

		roles, err := roleRepo.GetUserRoles(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, roles, 1)
		assert.Equal(t, role.ID, roles[0].ID)
	})

	t.Run("updates existing assignment (upsert)", func(t *testing.T) {
		user := helpers.UserFixture()
		role := helpers.RoleFixture("editor4")
		assigner1 := helpers.UserFixture()
		assigner2 := helpers.UserFixture()

		require.NoError(t, userRepo.Create(ctx, user))
		require.NoError(t, userRepo.Create(ctx, assigner1))
		require.NoError(t, userRepo.Create(ctx, assigner2))
		require.NoError(t, roleRepo.Create(ctx, role))

		// First assignment
		userRole1 := helpers.UserRoleFixture(user.ID, role.ID, func(ur *entity.UserRole) {
			ur.AssignedBy = &assigner1.ID
		})
		require.NoError(t, roleRepo.AssignToUser(ctx, userRole1))

		// Second assignment (update)
		userRole2 := helpers.UserRoleFixture(user.ID, role.ID, func(ur *entity.UserRole) {
			ur.AssignedBy = &assigner2.ID
			future := time.Now().Add(24 * time.Hour)
			ur.ExpiresAt = &future
		})
		err := roleRepo.AssignToUser(ctx, userRole2)
		require.NoError(t, err)

		roles, err := roleRepo.GetUserRoles(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, roles, 1)
	})

	t.Run("removes role from user", func(t *testing.T) {
		user := helpers.UserFixture()
		role := helpers.RoleFixture("editor5")

		require.NoError(t, userRepo.Create(ctx, user))
		require.NoError(t, roleRepo.Create(ctx, role))

		userRole := helpers.UserRoleFixture(user.ID, role.ID)
		require.NoError(t, roleRepo.AssignToUser(ctx, userRole))

		err := roleRepo.RemoveFromUser(ctx, user.ID, role.ID)
		require.NoError(t, err)

		roles, err := roleRepo.GetUserRoles(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("gets active roles only", func(t *testing.T) {
		user := helpers.UserFixture()
		role1 := helpers.RoleFixture("editor6")
		role2 := helpers.RoleFixture("viewer6")

		require.NoError(t, userRepo.Create(ctx, user))
		require.NoError(t, roleRepo.Create(ctx, role1))
		require.NoError(t, roleRepo.Create(ctx, role2))

		// Active role
		userRole1 := helpers.UserRoleFixture(user.ID, role1.ID)
		require.NoError(t, roleRepo.AssignToUser(ctx, userRole1))

		// Expired role
		past := time.Now().Add(-1 * time.Hour)
		userRole2 := helpers.UserRoleFixture(user.ID, role2.ID, func(ur *entity.UserRole) {
			ur.ExpiresAt = &past
		})
		require.NoError(t, roleRepo.AssignToUser(ctx, userRole2))

		activeRoles, err := roleRepo.GetUserActiveRoles(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, activeRoles, 1)
		assert.Equal(t, role1.ID, activeRoles[0].ID)
	})

	t.Run("gets users with role", func(t *testing.T) {
		user1 := helpers.UserFixture()
		user2 := helpers.UserFixture()
		role := helpers.RoleFixture("editor7")

		require.NoError(t, userRepo.Create(ctx, user1))
		require.NoError(t, userRepo.Create(ctx, user2))
		require.NoError(t, roleRepo.Create(ctx, role))

		userRole1 := helpers.UserRoleFixture(user1.ID, role.ID)
		userRole2 := helpers.UserRoleFixture(user2.ID, role.ID)
		require.NoError(t, roleRepo.AssignToUser(ctx, userRole1))
		require.NoError(t, roleRepo.AssignToUser(ctx, userRole2))

		userIDs, err := roleRepo.GetUsersWithRole(ctx, role.ID)
		require.NoError(t, err)
		assert.Len(t, userIDs, 2)

		// Verify IDs
		ids := make(map[uuidv7.UUID]bool)
		for _, id := range userIDs {
			ids[id] = true
		}
		assert.True(t, ids[user1.ID])
		assert.True(t, ids[user2.ID])
	})
}

func TestRoleRepository_CreateMany(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	testDB.CleanupTables(t)
	defer testDB.CleanupTables(t)

	repo := postgres.NewRoleRepository(testDB.DB)
	ctx := context.Background()

	t.Run("creates multiple roles", func(t *testing.T) {
		roles := []*entity.Role{
			helpers.RoleFixture("editor"),
			helpers.RoleFixture("viewer"),
			helpers.RoleFixture("contributor"),
		}

		err := repo.CreateMany(ctx, roles)
		require.NoError(t, err)

		for _, role := range roles {
			retrieved, err := repo.GetByID(ctx, role.ID)
			require.NoError(t, err)
			assert.Equal(t, role.Name, retrieved.Name)
		}
	})

	t.Run("creates empty list successfully", func(t *testing.T) {
		err := repo.CreateMany(ctx, []*entity.Role{})
		require.NoError(t, err)
	})
}
