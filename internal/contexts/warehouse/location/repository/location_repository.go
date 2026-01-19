package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/warehouse/location/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ILocationRepository defines the interface for location repository
type ILocationRepository interface {
	// CRUD operations
	Create(ctx context.Context, location *aggregate.Location) error
	Update(ctx context.Context, location *aggregate.Location) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Queries
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Location, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Location, error)
	List(ctx context.Context, limit, offset int) ([]*aggregate.Location, error)
	Count(ctx context.Context) (int, error)

	// Hierarchical queries
	ListByType(ctx context.Context, locType aggregate.LocationType, limit, offset int) ([]*aggregate.Location, error)
	CountByType(ctx context.Context, locType aggregate.LocationType) (int, error)
	ListByParent(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Location, error)
	ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Location, error)

	// Status queries
	ListByStatus(ctx context.Context, status aggregate.LocationStatus, limit, offset int) ([]*aggregate.Location, error)
	CountByStatus(ctx context.Context, status aggregate.LocationStatus) (int, error)
	ListAvailable(ctx context.Context, limit, offset int) ([]*aggregate.Location, error)
	CountAvailable(ctx context.Context) (int, error)
}
