package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IPermissionRepository defines operations for permission management
type IPermissionRepository interface {
	// Basic CRUD
	Create(ctx context.Context, permission *entity.Permission) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) (*entity.Permission, error)
	List(ctx context.Context) ([]*entity.Permission, error)
	Update(ctx context.Context, permission *entity.Permission) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Bulk operations
	CreateMany(ctx context.Context, permissions []*entity.Permission) error
	GetByIDs(ctx context.Context, ids []uuidv7.UUID) ([]*entity.Permission, error)

	// Search
	FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error)
}
