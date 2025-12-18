package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// RoleRepository defines operations for role management
type RoleRepository interface {
	// Basic CRUD
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	List(ctx context.Context) ([]*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Role permissions management
	AddPermission(ctx context.Context, roleID, permissionID uuidv7.UUID) error
	RemovePermission(ctx context.Context, roleID, permissionID uuidv7.UUID) error
	GetPermissions(ctx context.Context, roleID uuidv7.UUID) ([]*entity.Permission, error)
	SyncPermissions(ctx context.Context, roleID uuidv7.UUID, permissionIDs []uuidv7.UUID) error

	// Role assignment to users
	AssignToUser(ctx context.Context, userRole *entity.UserRole) error
	RemoveFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error
	GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error)
	GetUserActiveRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error)
	GetUsersWithRole(ctx context.Context, roleID uuidv7.UUID) ([]uuidv7.UUID, error)

	// Bulk operations
	CreateMany(ctx context.Context, roles []*entity.Role) error
}
