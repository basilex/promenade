package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/identity/role/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRoleRepository defines the interface for Role repository
type IRoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *aggregate.Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Role, error)

	// GetByName retrieves a role by name
	GetByName(ctx context.Context, name string) (*aggregate.Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *aggregate.Role) error

	// Delete soft deletes a role
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ExistsByName checks if a role with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ListRoles lists all roles with pagination
	ListRoles(ctx context.Context, limit, offset int) ([]*aggregate.Role, int, error)

	// GetUserRoles retrieves all roles for a user
	GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Role, error)

	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(ctx context.Context, userID, roleID uuidv7.UUID, assignedBy *uuidv7.UUID) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error
}
