package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermission_Constants(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		want       string
	}{
		{
			name:       "wildcard all",
			permission: PermissionAll,
			want:       "*:*",
		},
		{
			name:       "users create",
			permission: PermissionUsersCreate,
			want:       "users:create",
		},
		{
			name:       "posts all",
			permission: PermissionPostsAll,
			want:       "posts:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.permission)
		})
	}
}

func TestPermission_Format(t *testing.T) {
	// Test that all permission constants follow resource:action format
	permissions := []string{
		PermissionAll,
		PermissionUsersCreate,
		PermissionUsersRead,
		PermissionUsersUpdate,
		PermissionUsersDelete,
		PermissionPostsCreate,
		PermissionPostsAll,
	}

	for _, perm := range permissions {
		t.Run(perm, func(t *testing.T) {
			// All permissions should have a colon
			assert.Contains(t, perm, ":")
		})
	}
}
