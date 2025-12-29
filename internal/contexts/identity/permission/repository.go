package permission

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Permission repository
type IRepository interface {
	// Create creates a new permission
	Create(ctx context.Context, permission *Permission) error

	// GetByID retrieves a permission by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Permission, error)

	// GetByName retrieves a permission by name (resource:action)
	GetByName(ctx context.Context, name string) (*Permission, error)

	// Update updates an existing permission
	Update(ctx context.Context, permission *Permission) error

	// Delete deletes a permission
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ExistsByName checks if a permission with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ListPermissions lists all permissions with pagination
	ListPermissions(ctx context.Context, limit, offset int) ([]*Permission, int, error)

	// GetRolePermissions retrieves all permissions for a role
	GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*Permission, error)

	// AssignPermissionToRole assigns a permission to a role
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error

	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error
}
