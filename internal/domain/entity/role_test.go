package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestRole_Validate(t *testing.T) {
	t.Run("valid role", func(t *testing.T) {
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "admin",
			DisplayName: "Administrator",
			IsSystem:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		assert.NoError(t, role.Validate())
	})

	t.Run("valid system role", func(t *testing.T) {
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "superadmin",
			DisplayName: "Super Administrator",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		assert.NoError(t, role.Validate())
	})

	t.Run("empty name", func(t *testing.T) {
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "",
			DisplayName: "Test Role",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		assert.Error(t, role.Validate())
	})

	t.Run("name too short", func(t *testing.T) {
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "a",
			DisplayName: "Test Role",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		assert.Error(t, role.Validate())
	})

	t.Run("name too long", func(t *testing.T) {
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        string(make([]byte, 51)),
			DisplayName: "Test Role",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		assert.Error(t, role.Validate())
	})

	t.Run("empty display name", func(t *testing.T) {
		role := &entity.Role{
			ID:          uuidv7.New(),
			Name:        "admin",
			DisplayName: "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		assert.Error(t, role.Validate())
	})
}

func TestRole_HasPermission(t *testing.T) {
	t.Run("has exact permission", func(t *testing.T) {
		role := &entity.Role{
			Permissions: []*entity.Permission{
				{Resource: "posts", Action: "create"},
				{Resource: "posts", Action: "read"},
			},
		}
		assert.True(t, role.HasPermission("posts:create"))
		assert.True(t, role.HasPermission("posts:read"))
		assert.False(t, role.HasPermission("posts:delete"))
	})

	t.Run("has wildcard permission", func(t *testing.T) {
		role := &entity.Role{
			Permissions: []*entity.Permission{
				{Resource: "*", Action: "*"},
			},
		}
		assert.True(t, role.HasPermission("posts:create"))
		assert.True(t, role.HasPermission("users:delete"))
		assert.True(t, role.HasPermission("anything:anything"))
	})

	t.Run("has resource wildcard", func(t *testing.T) {
		role := &entity.Role{
			Permissions: []*entity.Permission{
				{Resource: "posts", Action: "*"},
			},
		}
		assert.True(t, role.HasPermission("posts:create"))
		assert.True(t, role.HasPermission("posts:delete"))
		assert.False(t, role.HasPermission("users:create"))
	})

	t.Run("no permissions", func(t *testing.T) {
		role := &entity.Role{
			Permissions: []*entity.Permission{},
		}
		assert.False(t, role.HasPermission("posts:create"))
	})

	t.Run("nil permissions", func(t *testing.T) {
		role := &entity.Role{
			Permissions: nil,
		}
		assert.False(t, role.HasPermission("posts:create"))
	})
}

func TestUserRole_IsExpired(t *testing.T) {
	now := time.Now()

	t.Run("not expired", func(t *testing.T) {
		future := now.Add(24 * time.Hour)
		userRole := &entity.UserRole{
			ExpiresAt: &future,
		}
		assert.False(t, userRole.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		past := now.Add(-24 * time.Hour)
		userRole := &entity.UserRole{
			ExpiresAt: &past,
		}
		assert.True(t, userRole.IsExpired())
	})

	t.Run("no expiration", func(t *testing.T) {
		userRole := &entity.UserRole{
			ExpiresAt: nil,
		}
		assert.False(t, userRole.IsExpired())
	})

	t.Run("expires exactly now", func(t *testing.T) {
		// This is a bit tricky - time.Now() returns a slightly different time
		// so we need to be careful with the comparison
		expiresAt := time.Now().Add(-1 * time.Millisecond)
		userRole := &entity.UserRole{
			ExpiresAt: &expiresAt,
		}
		assert.True(t, userRole.IsExpired())
	})
}

func TestUserRole_IsActive(t *testing.T) {
	now := time.Now()

	t.Run("active with future expiration", func(t *testing.T) {
		future := now.Add(24 * time.Hour)
		userRole := &entity.UserRole{
			ExpiresAt: &future,
		}
		assert.True(t, userRole.IsActive())
	})

	t.Run("inactive when expired", func(t *testing.T) {
		past := now.Add(-24 * time.Hour)
		userRole := &entity.UserRole{
			ExpiresAt: &past,
		}
		assert.False(t, userRole.IsActive())
	})

	t.Run("active with no expiration", func(t *testing.T) {
		userRole := &entity.UserRole{
			ExpiresAt: nil,
		}
		assert.True(t, userRole.IsActive())
	})
}

func TestGetSystemRoles(t *testing.T) {
	roles := entity.GetSystemRoles()

	t.Run("returns 5 system roles", func(t *testing.T) {
		assert.Len(t, roles, 5)
	})

	t.Run("all roles are marked as system", func(t *testing.T) {
		for _, role := range roles {
			assert.True(t, role.IsSystem, "role %s should be system role", role.Name)
		}
	})

	t.Run("all roles have required fields", func(t *testing.T) {
		for _, role := range roles {
			assert.NotEmpty(t, role.Name, "role should have name")
			assert.NotEmpty(t, role.DisplayName, "role should have display name")
			assert.NotNil(t, role.Permissions, "role should have permissions")
			assert.NoError(t, role.Validate(), "role %s should be valid", role.Name)
		}
	})

	t.Run("superadmin has all permissions", func(t *testing.T) {
		var superadmin *entity.Role
		for _, role := range roles {
			if role.Name == "superadmin" {
				superadmin = role
				break
			}
		}
		require.NotNil(t, superadmin, "superadmin role should exist")

		// Check for wildcard permission
		found := false
		for _, perm := range superadmin.Permissions {
			if perm.Resource == "*" && perm.Action == "*" {
				found = true
				break
			}
		}
		assert.True(t, found, "superadmin should have *:* permission")
	})

	t.Run("admin has management permissions", func(t *testing.T) {
		var admin *entity.Role
		for _, role := range roles {
			if role.Name == "admin" {
				admin = role
				break
			}
		}
		require.NotNil(t, admin, "admin role should exist")
		assert.True(t, len(admin.Permissions) > 0, "admin should have permissions")
		assert.True(t, admin.HasPermission("users:manage"))
	})

	t.Run("moderator has moderation permissions", func(t *testing.T) {
		var moderator *entity.Role
		for _, role := range roles {
			if role.Name == "moderator" {
				moderator = role
				break
			}
		}
		require.NotNil(t, moderator, "moderator role should exist")
		assert.True(t, len(moderator.Permissions) > 0, "moderator should have permissions")
		assert.True(t, moderator.HasPermission("posts:delete"))
	})

	t.Run("user has basic permissions", func(t *testing.T) {
		var user *entity.Role
		for _, role := range roles {
			if role.Name == "user" {
				user = role
				break
			}
		}
		require.NotNil(t, user, "user role should exist")
		assert.True(t, len(user.Permissions) > 0, "user should have permissions")
		assert.True(t, user.HasPermission("posts:create"))
	})

	t.Run("guest has read-only permissions", func(t *testing.T) {
		var guest *entity.Role
		for _, role := range roles {
			if role.Name == "guest" {
				guest = role
				break
			}
		}
		require.NotNil(t, guest, "guest role should exist")
		assert.True(t, len(guest.Permissions) > 0, "guest should have permissions")
		assert.True(t, guest.HasPermission("posts:read"))
		assert.False(t, guest.HasPermission("posts:create"))
	})

	t.Run("role names are unique", func(t *testing.T) {
		names := make(map[string]bool)
		for _, role := range roles {
			assert.False(t, names[role.Name], "duplicate role name: %s", role.Name)
			names[role.Name] = true
		}
	})
}
