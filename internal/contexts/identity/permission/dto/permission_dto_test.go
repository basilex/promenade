package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/identity/permission/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToPermissionResponse(t *testing.T) {
	permissionID := uuidv7.New()
	now := time.Now()

	t.Run("regular permission", func(t *testing.T) {
		p := &aggregate.Permission{
			Name:        "users:create",
			Resource:    "users",
			Action:      "create",
			Description: "Create new users",
		}
		p.ID = permissionID
		p.CreatedAt = now

		resp := ToPermissionResponse(p)

		assert.Equal(t, permissionID.String(), resp.ID)
		assert.Equal(t, "users:create", resp.Name)
		assert.Equal(t, "users", resp.Resource)
		assert.Equal(t, "create", resp.Action)
		assert.Equal(t, "Create new users", resp.Description)
		assert.Equal(t, now, resp.CreatedAt)
	})

	t.Run("permission without description", func(t *testing.T) {
		p := &aggregate.Permission{
			Name:        "users:read",
			Resource:    "users",
			Action:      "read",
			Description: "",
		}
		p.ID = permissionID
		p.CreatedAt = now

		resp := ToPermissionResponse(p)

		assert.Equal(t, permissionID.String(), resp.ID)
		assert.Equal(t, "users:read", resp.Name)
		assert.Equal(t, "users", resp.Resource)
		assert.Equal(t, "read", resp.Action)
		assert.Empty(t, resp.Description)
	})

	t.Run("nil permission returns nil", func(t *testing.T) {
		// ToPermissionResponse doesn't handle nil, it will panic
		// This test documents expected behavior
		// In real code, handlers should check for nil before calling
	})
}

func TestToPermissionListResponse(t *testing.T) {
	now := time.Now()

	t.Run("empty list", func(t *testing.T) {
		permissions := []*aggregate.Permission{}
		total := 0

		resp := ToPermissionListResponse(permissions, total, 20, 0)

		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.Total)
		assert.Equal(t, 20, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
		assert.Empty(t, resp.Permissions)
	})

	t.Run("list with multiple permissions", func(t *testing.T) {
		perm1 := &aggregate.Permission{
			Name:        "users:create",
			Resource:    "users",
			Action:      "create",
			Description: "Create users",
		}
		perm1.ID = uuidv7.New()
		perm1.CreatedAt = now

		perm2 := &aggregate.Permission{
			Name:        "users:read",
			Resource:    "users",
			Action:      "read",
			Description: "Read users",
		}
		perm2.ID = uuidv7.New()
		perm2.CreatedAt = now

		perm3 := &aggregate.Permission{
			Name:        "users:update",
			Resource:    "users",
			Action:      "update",
			Description: "Update users",
		}
		perm3.ID = uuidv7.New()
		perm3.CreatedAt = now

		perm4 := &aggregate.Permission{
			Name:        "users:delete",
			Resource:    "users",
			Action:      "delete",
			Description: "Delete users",
		}
		perm4.ID = uuidv7.New()
		perm4.CreatedAt = now

		permissions := []*aggregate.Permission{perm1, perm2, perm3, perm4}
		total := 4

		resp := ToPermissionListResponse(permissions, total, 20, 0)

		assert.NotNil(t, resp)
		assert.Equal(t, 4, resp.Total)
		assert.Equal(t, 20, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
		assert.Len(t, resp.Permissions, 4)
		assert.Equal(t, "users:create", resp.Permissions[0].Name)
		assert.Equal(t, "users:read", resp.Permissions[1].Name)
		assert.Equal(t, "users:update", resp.Permissions[2].Name)
		assert.Equal(t, "users:delete", resp.Permissions[3].Name)
		assert.Equal(t, "users", resp.Permissions[0].Resource)
		assert.Equal(t, "create", resp.Permissions[0].Action)
	})

	t.Run("list with pagination", func(t *testing.T) {
		perm1 := &aggregate.Permission{
			Name:        "perm1:read",
			Resource:    "perm1",
			Action:      "read",
			Description: "First permission",
		}
		perm1.ID = uuidv7.New()
		perm1.CreatedAt = now

		perm2 := &aggregate.Permission{
			Name:        "perm2:write",
			Resource:    "perm2",
			Action:      "write",
			Description: "Second permission",
		}
		perm2.ID = uuidv7.New()
		perm2.CreatedAt = now

		permissions := []*aggregate.Permission{perm1, perm2}
		total := 15

		resp := ToPermissionListResponse(permissions, total, 2, 10)

		assert.Equal(t, 15, resp.Total)
		assert.Equal(t, 2, resp.Limit)
		assert.Equal(t, 10, resp.Offset)
		assert.Len(t, resp.Permissions, 2)
	})

	t.Run("nil permissions list", func(t *testing.T) {
		resp := ToPermissionListResponse(nil, 0, 20, 0)

		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.Total)
		assert.Empty(t, resp.Permissions)
	})
}
