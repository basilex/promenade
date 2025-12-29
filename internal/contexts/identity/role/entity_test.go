package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRole(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		displayName string
		description string
		wantErr     bool
	}{
		{
			name:        "valid role",
			roleName:    "manager",
			displayName: "Manager",
			description: "Manager role",
			wantErr:     false,
		},
		{
			name:        "role name with spaces",
			roleName:    "  admin  ",
			displayName: "Administrator",
			description: "Admin role",
			wantErr:     false,
		},
		{
			name:        "empty role name",
			roleName:    "",
			displayName: "Test",
			description: "Test role",
			wantErr:     true,
		},
		{
			name:        "empty display name",
			roleName:    "admin",
			displayName: "",
			description: "Test role",
			wantErr:     true,
		},
		{
			name:        "role name too short",
			roleName:    "a",
			displayName: "Short",
			description: "Test role",
			wantErr:     true,
		},
		{
			name:        "role name too long",
			roleName:    "this_is_a_very_long_role_name_that_exceeds_fifty_characters_limit",
			displayName: "Too Long",
			description: "Test role",
			wantErr:     true,
		},
		{
			name:        "role name with invalid characters",
			roleName:    "admin@role",
			displayName: "Invalid",
			description: "Test role",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := NewRole(tt.roleName, tt.displayName, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.False(t, role.IsSystem)
				assert.Equal(t, tt.displayName, role.DisplayName)
			}
		})
	}
}

func TestNewSystemRole(t *testing.T) {
	role, err := NewSystemRole("admin", "System Administrator", "System administrator")

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.True(t, role.IsSystem)
	assert.Equal(t, "admin", role.Name)
	assert.Equal(t, "System Administrator", role.DisplayName)
}

func TestRole_UpdateDescription(t *testing.T) {
	role, _ := NewRole("manager", "Manager", "Manager role")
	originalUpdatedAt := role.UpdatedAt

	role.UpdateDescription("Updated manager role")

	assert.Equal(t, "Updated manager role", role.Description)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
}

func TestRole_UpdateDisplayName(t *testing.T) {
	role, _ := NewRole("manager", "Manager", "Manager role")
	originalUpdatedAt := role.UpdatedAt

	err := role.UpdateDisplayName("Senior Manager")

	assert.NoError(t, err)
	assert.Equal(t, "Senior Manager", role.DisplayName)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))

	// Test empty display name
	err = role.UpdateDisplayName("")
	assert.Error(t, err)
}

func TestRole_CanDelete(t *testing.T) {
	t.Run("system role cannot be deleted", func(t *testing.T) {
		role, _ := NewSystemRole("admin", "Administrator", "Admin")
		err := role.CanDelete()
		assert.Error(t, err)
		assert.Equal(t, ErrCannotDeleteSystem, err)
	})

	t.Run("regular role can be deleted", func(t *testing.T) {
		role, _ := NewRole("custom", "Custom Role", "Custom role")
		err := role.CanDelete()
		assert.NoError(t, err)
	})
}

func TestRole_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    *Role
		wantErr bool
	}{
		{
			name: "valid role",
			role: &Role{
				Name:        "admin",
				DisplayName: "Administrator",
				Description: "Administrator",
			},
			wantErr: false,
		},
		{
			name: "empty display name",
			role: &Role{
				Name:        "admin",
				DisplayName: "",
				Description: "Test",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			role: &Role{
				Name:        "admin",
				DisplayName: "Administrator",
				Description: "",
			},
			wantErr: true,
		},
		{
			name: "invalid name",
			role: &Role{
				Name:        "",
				DisplayName: "Test",
				Description: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
