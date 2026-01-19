package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/identity/permission/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IPermissionRepository defines the interface for Permission repository
type IPermissionRepository interface {
	// Create creates a new permission
	Create(ctx context.Context, permission *aggregate.Permission) error

	// GetByID retrieves a permission by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Permission, error)

	// GetByName retrieves a permission by name (resource:action)
	GetByName(ctx context.Context, name string) (*aggregate.Permission, error)

	// Update updates an existing permission
	Update(ctx context.Context, permission *aggregate.Permission) error

	// Delete deletes a permission
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ExistsByName checks if a permission with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ListPermissions lists all permissions with pagination
	ListPermissions(ctx context.Context, limit, offset int) ([]*aggregate.Permission, int, error)

	// GetRolePermissions retrieves all permissions for a role
	GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*aggregate.Permission, error)

	// AssignPermissionToRole assigns a permission to a role
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error

	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error
}
