package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
)

func TestNewPermission(t *testing.T) {
	t.Run("creates permission from valid string", func(t *testing.T) {
		perm, err := entity.NewPermission("posts:create")
		require.NoError(t, err)
		assert.Equal(t, "posts", perm.Resource)
		assert.Equal(t, "create", perm.Action)
	})

	t.Run("creates wildcard permission", func(t *testing.T) {
		perm, err := entity.NewPermission("*:*")
		require.NoError(t, err)
		assert.Equal(t, "*", perm.Resource)
		assert.Equal(t, "*", perm.Action)
	})

	t.Run("creates resource wildcard", func(t *testing.T) {
		perm, err := entity.NewPermission("posts:*")
		require.NoError(t, err)
		assert.Equal(t, "posts", perm.Resource)
		assert.Equal(t, "*", perm.Action)
	})

	t.Run("creates action wildcard", func(t *testing.T) {
		perm, err := entity.NewPermission("*:read")
		require.NoError(t, err)
		assert.Equal(t, "*", perm.Resource)
		assert.Equal(t, "read", perm.Action)
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		testCases := []string{
			"posts",              // missing action
			"posts:create:extra", // too many parts
			"",                   // empty
			"posts:",             // missing action
			":create",            // missing resource
		}

		for _, tc := range testCases {
			_, err := entity.NewPermission(tc)
			assert.Error(t, err, "should error for: %s", tc)
		}
	})
}

func TestPermission_String(t *testing.T) {
	t.Run("returns formatted string", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "posts",
			Action:   "create",
		}
		assert.Equal(t, "posts:create", perm.String())
	})

	t.Run("returns wildcard string", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "*",
			Action:   "*",
		}
		assert.Equal(t, "*:*", perm.String())
	})
}

func TestPermission_Matches(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		perm := &entity.Permission{Resource: "posts", Action: "create"}
		assert.True(t, perm.Matches("posts:create"))
		assert.False(t, perm.Matches("posts:read"))
		assert.False(t, perm.Matches("users:create"))
	})

	t.Run("full wildcard matches everything", func(t *testing.T) {
		perm := &entity.Permission{Resource: "*", Action: "*"}
		assert.True(t, perm.Matches("posts:create"))
		assert.True(t, perm.Matches("users:delete"))
		assert.True(t, perm.Matches("anything:anything"))
	})

	t.Run("resource wildcard matches any resource", func(t *testing.T) {
		perm := &entity.Permission{Resource: "*", Action: "read"}
		assert.True(t, perm.Matches("posts:read"))
		assert.True(t, perm.Matches("users:read"))
		assert.False(t, perm.Matches("posts:create"))
	})

	t.Run("action wildcard matches any action", func(t *testing.T) {
		perm := &entity.Permission{Resource: "posts", Action: "*"}
		assert.True(t, perm.Matches("posts:read"))
		assert.True(t, perm.Matches("posts:create"))
		assert.True(t, perm.Matches("posts:delete"))
		assert.False(t, perm.Matches("users:read"))
	})

	t.Run("invalid required permission format", func(t *testing.T) {
		perm := &entity.Permission{Resource: "posts", Action: "create"}
		assert.False(t, perm.Matches("invalid"))
		assert.False(t, perm.Matches(""))
	})
}

func TestPermission_Validate(t *testing.T) {
	t.Run("valid permission", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "posts",
			Action:   "create",
		}
		assert.NoError(t, perm.Validate())
	})

	t.Run("valid wildcard permission", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "*",
			Action:   "*",
		}
		assert.NoError(t, perm.Validate())
	})

	t.Run("empty resource", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "",
			Action:   "create",
		}
		assert.Error(t, perm.Validate())
	})

	t.Run("empty action", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "posts",
			Action:   "",
		}
		assert.Error(t, perm.Validate())
	})

	t.Run("resource too long", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: string(make([]byte, 51)), // 51 characters
			Action:   "create",
		}
		assert.Error(t, perm.Validate())
	})

	t.Run("action too long", func(t *testing.T) {
		perm := &entity.Permission{
			Resource: "posts",
			Action:   string(make([]byte, 51)), // 51 characters
		}
		assert.Error(t, perm.Validate())
	})
}

func TestPermissionConstants(t *testing.T) {
	t.Run("all permission constants are valid", func(t *testing.T) {
		permissions := []string{
			entity.PermissionAll,
			entity.PermissionUsersAll,
			entity.PermissionUsersCreate,
			entity.PermissionUsersRead,
			entity.PermissionUsersUpdate,
			entity.PermissionUsersDelete,
			entity.PermissionUsersList,
			entity.PermissionUsersBan,
			entity.PermissionPostsAll,
			entity.PermissionPostsCreate,
			entity.PermissionPostsRead,
			entity.PermissionPostsUpdate,
			entity.PermissionPostsDelete,
			entity.PermissionPostsList,
			entity.PermissionCommentsAll,
			entity.PermissionCommentsCreate,
			entity.PermissionCommentsRead,
			entity.PermissionCommentsUpdate,
			entity.PermissionCommentsDelete,
			entity.PermissionCommentsList,
			entity.PermissionProfilesAll,
			entity.PermissionProfilesCreate,
			entity.PermissionProfilesRead,
			entity.PermissionProfilesUpdate,
			entity.PermissionProfilesDelete,
			entity.PermissionProfilesList,
			entity.PermissionRolesAll,
			entity.PermissionRolesCreate,
			entity.PermissionRolesRead,
			entity.PermissionRolesUpdate,
			entity.PermissionRolesDelete,
			entity.PermissionRolesList,
			entity.PermissionRolesAssign,
		}

		for _, permStr := range permissions {
			perm, err := entity.NewPermission(permStr)
			assert.NoError(t, err, "permission constant %s should be valid", permStr)
			assert.NotNil(t, perm)
		}
	})
}
