package role

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Role repository
type IRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Role, error)

	// GetByName retrieves a role by name
	GetByName(ctx context.Context, name string) (*Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *Role) error

	// Delete soft deletes a role
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ExistsByName checks if a role with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)

	// ListRoles lists all roles
	ListRoles(ctx context.Context) ([]*Role, error)

	// GetUserRoles retrieves all roles for a user
	GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*Role, error)

	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(ctx context.Context, userID, roleID uuidv7.UUID, assignedBy *uuidv7.UUID) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error
}
