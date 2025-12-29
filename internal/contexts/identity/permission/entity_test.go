package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPermission(t *testing.T) {
	tests := []struct {
		name        string
		resource    string
		action      string
		description string
		wantErr     bool
		wantName    string
	}{
		{
			name:        "valid permission",
			resource:    "users",
			action:      "read",
			description: "Read users",
			wantErr:     false,
			wantName:    "users:read",
		},
		{
			name:        "uppercase resource and action",
			resource:    "CUSTOMERS",
			action:      "WRITE",
			description: "Write customers",
			wantErr:     false,
			wantName:    "customers:write",
		},
		{
			name:        "empty resource",
			resource:    "",
			action:      "read",
			description: "Test",
			wantErr:     true,
		},
		{
			name:        "empty action",
			resource:    "users",
			action:      "",
			description: "Test",
			wantErr:     true,
		},
		{
			name:        "invalid action with special chars",
			resource:    "users",
			action:      "read@write",
			description: "Test",
			wantErr:     true,
		},
		{
			name:        "resource too short",
			resource:    "u",
			action:      "read",
			description: "Test",
			wantErr:     true,
		},
		{
			name:        "resource too long",
			resource:    "this_is_a_very_long_resource_name_that_exceeds_limit",
			action:      "read",
			description: "Test",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, err := NewPermission(tt.resource, tt.action, tt.description)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, perm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, perm)
				assert.Equal(t, tt.wantName, perm.Name)
			}
		})
	}
}

func TestPermission_Validate(t *testing.T) {
	tests := []struct {
		name    string
		perm    *Permission
		wantErr bool
	}{
		{
			name: "valid permission",
			perm: &Permission{
				Name:        "users:read",
				Resource:    "users",
				Action:      "read",
				Description: "Read users",
			},
			wantErr: false,
		},
		{
			name: "empty description",
			perm: &Permission{
				Name:        "users:read",
				Resource:    "users",
				Action:      "read",
				Description: "",
			},
			wantErr: true,
		},
		{
			name: "invalid resource",
			perm: &Permission{
				Name:        "users:read",
				Resource:    "",
				Action:      "read",
				Description: "Test",
			},
			wantErr: true,
		},
		{
			name: "invalid action with special chars",
			perm: &Permission{
				Name:        "users:read@write",
				Resource:    "users",
				Action:      "read@write",
				Description: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.perm.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAction(t *testing.T) {
	// Test valid actions (lowercase alphanumeric + underscore + hyphen + wildcard)
	validActions := []string{
		"read", "write", "delete", "admin", "create", "update", "list",
		"*",                        // wildcard
		"custom_action",            // underscore
		"custom-action",            // hyphen
		"execute", "modify", "any", // extensible - any lowercase is valid
	}
	for _, action := range validActions {
		t.Run("valid action: "+action, func(t *testing.T) {
			err := validateAction(action)
			assert.NoError(t, err)
		})
	}

	// Test invalid actions (uppercase, special chars, spaces)
	invalidActions := []struct {
		action string
		name   string
	}{
		{action: "Read", name: "uppercase"},
		{action: "read@write", name: "special_char_@"},
		{action: "read write", name: "space"},
		{action: "", name: "empty"},
	}
	for _, tt := range invalidActions {
		t.Run("invalid action: "+tt.name, func(t *testing.T) {
			err := validateAction(tt.action)
			assert.Error(t, err)
		})
	}
}
