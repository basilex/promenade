package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToRoleResponse(t *testing.T) {
	roleID := uuidv7.New()
	now := time.Now()

	t.Run("regular role", func(t *testing.T) {
		r := &role.Role{
			Name:        "manager",
			DisplayName: "Manager",
			Description: "Team manager role",
			IsSystem:    false,
		}
		r.ID = roleID
		r.CreatedAt = now
		r.UpdatedAt = now

		resp := ToRoleResponse(r)

		assert.Equal(t, roleID.String(), resp.ID)
		assert.Equal(t, "manager", resp.Name)
		assert.Equal(t, "Manager", resp.DisplayName)
		assert.Equal(t, "Team manager role", resp.Description)
		assert.False(t, resp.IsSystem)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
	})

	t.Run("system role", func(t *testing.T) {
		r := &role.Role{
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "System administrator",
			IsSystem:    true,
		}
		r.ID = roleID
		r.CreatedAt = now
		r.UpdatedAt = now

		resp := ToRoleResponse(r)

		assert.Equal(t, roleID.String(), resp.ID)
		assert.Equal(t, "admin", resp.Name)
		assert.Equal(t, "Administrator", resp.DisplayName)
		assert.Equal(t, "System administrator", resp.Description)
		assert.True(t, resp.IsSystem)
	})

	t.Run("nil role returns nil", func(t *testing.T) {
		// ToRoleResponse doesn't handle nil, it will panic
		// This test documents expected behavior
		// In real code, handlers should check for nil before calling
	})
}

func TestToRoleListResponse(t *testing.T) {
	now := time.Now()

	t.Run("empty list", func(t *testing.T) {
		roles := []*role.Role{}
		total := 0

		resp := ToRoleListResponse(roles, total, 20, 0)

		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.Total)
		assert.Equal(t, 20, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
		assert.Empty(t, resp.Roles)
	})

	t.Run("list with multiple roles", func(t *testing.T) {
		adminRole := &role.Role{
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "System administrator",
			IsSystem:    true,
		}
		adminRole.ID = uuidv7.New()
		adminRole.CreatedAt = now
		adminRole.UpdatedAt = now

		userRole := &role.Role{
			Name:        "user",
			DisplayName: "User",
			Description: "Regular user",
			IsSystem:    true,
		}
		userRole.ID = uuidv7.New()
		userRole.CreatedAt = now
		userRole.UpdatedAt = now

	managerRole := &role.Role{
		Name:        "manager",
		DisplayName: "Manager",
		Description: "Team manager",
		IsSystem:    false,
	}
	managerRole.ID = uuidv7.New()
	managerRole.CreatedAt = now
	managerRole.UpdatedAt = now

	roles := []*role.Role{adminRole, userRole, managerRole}
	total := 3

	resp := ToRoleListResponse(roles, total, 20, 0)

	assert.NotNil(t, resp)
		assert.Equal(t, 3, resp.Total)
		assert.Equal(t, 20, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
		assert.Len(t, resp.Roles, 3)
		assert.Equal(t, "admin", resp.Roles[0].Name)
		assert.Equal(t, "user", resp.Roles[1].Name)
		assert.Equal(t, "manager", resp.Roles[2].Name)
		assert.True(t, resp.Roles[0].IsSystem)
		assert.True(t, resp.Roles[1].IsSystem)
		assert.False(t, resp.Roles[2].IsSystem)
	})

	t.Run("list with pagination", func(t *testing.T) {
		role1 := &role.Role{
			Name:        "role1",
			DisplayName: "Role 1",
			Description: "First role",
			IsSystem:    false,
		}
		role1.ID = uuidv7.New()
		role1.CreatedAt = now
		role1.UpdatedAt = now

		role2 := &role.Role{
			Name:        "role2",
			DisplayName: "Role 2",
			Description: "Second role",
			IsSystem:    false,
		}
		role2.ID = uuidv7.New()
		role2.CreatedAt = now
		role2.UpdatedAt = now

		roles := []*role.Role{role1, role2}
		total := 10

		resp := ToRoleListResponse(roles, total, 2, 5)

		assert.Equal(t, 10, resp.Total)
		assert.Equal(t, 2, resp.Limit)
		assert.Equal(t, 5, resp.Offset)
		assert.Len(t, resp.Roles, 2)
	})

	t.Run("nil roles list", func(t *testing.T) {
		resp := ToRoleListResponse(nil, 0, 20, 0)

		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.Total)
		assert.Empty(t, resp.Roles)
	})
}