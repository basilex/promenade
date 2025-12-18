package entity

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Role represents a collection of permissions
type Role struct {
	ID          uuidv7.UUID   `db:"id" json:"id"`
	Name        string        `db:"name" json:"name"`                 // Unique name (e.g., "admin", "moderator")
	DisplayName string        `db:"display_name" json:"display_name"` // Human-readable name
	Description *string       `db:"description" json:"description"`   // Role description
	IsSystem    bool          `db:"is_system" json:"is_system"`       // Cannot be deleted if true
	Permissions []*Permission `db:"-" json:"permissions,omitempty"`   // Loaded separately
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	UserID     uuidv7.UUID  `db:"user_id" json:"user_id"`
	RoleID     uuidv7.UUID  `db:"role_id" json:"role_id"`
	Role       *Role        `db:"-" json:"role,omitempty"` // Loaded separately
	AssignedAt time.Time    `db:"assigned_at" json:"assigned_at"`
	AssignedBy *uuidv7.UUID `db:"assigned_by" json:"assigned_by"` // Who assigned this role
	ExpiresAt  *time.Time   `db:"expires_at" json:"expires_at"`   // Optional expiration
}

// IsExpired checks if the role assignment has expired
func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*ur.ExpiresAt)
}

// IsActive checks if the role assignment is currently active
func (ur *UserRole) IsActive() bool {
	return !ur.IsExpired()
}

// Validate validates role fields
func (r *Role) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("role name is required")
	}
	if len(r.Name) < 2 {
		return fmt.Errorf("role name must be at least 2 characters")
	}
	if len(r.Name) > 50 {
		return fmt.Errorf("role name must be 50 characters or less")
	}
	if r.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}
	if len(r.DisplayName) > 100 {
		return fmt.Errorf("display name must be 100 characters or less")
	}
	return nil
}

// HasPermission checks if role has a specific permission
func (r *Role) HasPermission(required string) bool {
	for _, perm := range r.Permissions {
		if perm.Matches(required) {
			return true
		}
	}
	return false
}

// System role constants
const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
	RoleModerator  = "moderator"
	RoleUser       = "user"
	RoleGuest      = "guest"
)

// GetSystemRoles returns predefined system roles with their permissions
func GetSystemRoles() []*Role {
	now := time.Now()

	return []*Role{
		{
			ID:          uuidv7.New(),
			Name:        RoleSuperAdmin,
			DisplayName: "Super Administrator",
			Description: strPtr("Full system access with all permissions"),
			IsSystem:    true,
			Permissions: []*Permission{
				MustNewPermission(PermissionAll),
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          uuidv7.New(),
			Name:        RoleAdmin,
			DisplayName: "Administrator",
			Description: strPtr("Administrative access to manage users and content"),
			IsSystem:    true,
			Permissions: []*Permission{
				MustNewPermission(PermissionUsersAll),
				MustNewPermission(PermissionPostsAll),
				MustNewPermission(PermissionCommentsAll),
				MustNewPermission(PermissionProfilesAll),
				MustNewPermission(PermissionRolesRead),
				MustNewPermission(PermissionRolesList),
				MustNewPermission(PermissionRolesAssign),
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          uuidv7.New(),
			Name:        RoleModerator,
			DisplayName: "Moderator",
			Description: strPtr("Can moderate user content and comments"),
			IsSystem:    true,
			Permissions: []*Permission{
				MustNewPermission(PermissionUsersRead),
				MustNewPermission(PermissionUsersList),
				MustNewPermission(PermissionPostsRead),
				MustNewPermission(PermissionPostsUpdate),
				MustNewPermission(PermissionPostsDelete),
				MustNewPermission(PermissionPostsList),
				MustNewPermission(PermissionCommentsAll),
				MustNewPermission(PermissionProfilesRead),
				MustNewPermission(PermissionProfilesList),
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          uuidv7.New(),
			Name:        RoleUser,
			DisplayName: "User",
			Description: strPtr("Regular user with basic permissions"),
			IsSystem:    true,
			Permissions: []*Permission{
				MustNewPermission(PermissionPostsCreate),
				MustNewPermission(PermissionPostsRead),
				MustNewPermission(PermissionCommentsCreate),
				MustNewPermission(PermissionCommentsRead),
				MustNewPermission(PermissionProfilesCreate),
				MustNewPermission(PermissionProfilesRead),
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          uuidv7.New(),
			Name:        RoleGuest,
			DisplayName: "Guest",
			Description: strPtr("Limited read-only access"),
			IsSystem:    true,
			Permissions: []*Permission{
				MustNewPermission(PermissionPostsRead),
				MustNewPermission(PermissionCommentsRead),
				MustNewPermission(PermissionProfilesRead),
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func strPtr(s string) *string {
	return &s
}
