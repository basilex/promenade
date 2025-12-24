package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICityRepository defines the interface for city data operations
type ICityRepository interface {
	// Create creates a new city
	Create(ctx context.Context, city *entity.City) error

	// GetByID retrieves a city by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.City, error)

	// List retrieves all cities with pagination
	List(ctx context.Context, offset, limit int) ([]entity.City, int, error)

	// ListByCountry retrieves cities for a specific country
	ListByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.City, int, error)

	// ListByRegion retrieves cities for a specific region
	ListByRegion(ctx context.Context, regionID uuidv7.UUID, offset, limit int) ([]entity.City, int, error)

	// ListActive retrieves only active cities
	ListActive(ctx context.Context, offset, limit int) ([]entity.City, int, error)

	// ListActiveByCountry retrieves active cities for a specific country
	ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.City, int, error)

	// ListCapitals retrieves all capital cities
	ListCapitals(ctx context.Context, offset, limit int) ([]entity.City, int, error)

	// SearchByName searches cities by name (supports partial matching)
	SearchByName(ctx context.Context, query string, offset, limit int) ([]entity.City, int, error)

	// Update updates an existing city
	Update(ctx context.Context, city *entity.City) error

	// Delete deletes a city by ID
	Delete(ctx context.Context, id uuidv7.UUID) error
}
