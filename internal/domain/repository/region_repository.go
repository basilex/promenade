package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRegionRepository defines the interface for region data operations
type IRegionRepository interface {
	// Create creates a new region
	Create(ctx context.Context, region *entity.Region) error

	// GetByID retrieves a region by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Region, error)

	// GetByCode retrieves a region by country and code
	GetByCode(ctx context.Context, countryID uuidv7.UUID, code string) (*entity.Region, error)

	// List retrieves all regions with pagination
	List(ctx context.Context, offset, limit int) ([]entity.Region, int, error)

	// ListByCountry retrieves regions for a specific country
	ListByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.Region, int, error)

	// ListActive retrieves only active regions
	ListActive(ctx context.Context, offset, limit int) ([]entity.Region, int, error)

	// ListActiveByCountry retrieves active regions for a specific country
	ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.Region, int, error)

	// Update updates an existing region
	Update(ctx context.Context, region *entity.Region) error

	// Delete deletes a region by ID
	Delete(ctx context.Context, id uuidv7.UUID) error
}
