package location

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for location repository
type IRepository interface {
	// CRUD operations
	Create(ctx context.Context, location *Location) error
	Update(ctx context.Context, location *Location) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Queries
	GetByID(ctx context.Context, id uuidv7.UUID) (*Location, error)
	GetByCode(ctx context.Context, code string) (*Location, error)
	List(ctx context.Context, limit, offset int) ([]*Location, error)
	Count(ctx context.Context) (int, error)

	// Hierarchical queries
	ListByType(ctx context.Context, locType LocationType, limit, offset int) ([]*Location, error)
	CountByType(ctx context.Context, locType LocationType) (int, error)
	ListByParent(ctx context.Context, parentID uuidv7.UUID) ([]*Location, error)
	ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*Location, error)

	// Status queries
	ListByStatus(ctx context.Context, status LocationStatus, limit, offset int) ([]*Location, error)
	CountByStatus(ctx context.Context, status LocationStatus) (int, error)
	ListAvailable(ctx context.Context, limit, offset int) ([]*Location, error)
	CountAvailable(ctx context.Context) (int, error)
}
