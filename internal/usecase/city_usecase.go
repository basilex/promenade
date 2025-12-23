package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CityUseCase interface defines operations for city management
type CityUseCase interface {
	Create(ctx context.Context, city *entity.City) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.City, error)
	List(ctx context.Context, page, pageSize int) ([]entity.City, int, error)
	ListByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.City, int, error)
	ListByRegion(ctx context.Context, regionID uuidv7.UUID, page, pageSize int) ([]entity.City, int, error)
	ListActive(ctx context.Context, page, pageSize int) ([]entity.City, int, error)
	ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.City, int, error)
	ListCapitals(ctx context.Context, page, pageSize int) ([]entity.City, int, error)
	SearchByName(ctx context.Context, query string, page, pageSize int) ([]entity.City, int, error)
	Update(ctx context.Context, city *entity.City) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type cityUseCase struct {
	cityRepo repository.CityRepository
}

func NewCityUseCase(cityRepo repository.CityRepository) CityUseCase {
	return &cityUseCase{
		cityRepo: cityRepo,
	}
}

// Create creates a new city
func (uc *cityUseCase) Create(ctx context.Context, city *entity.City) error {
	// Validate entity
	if err := city.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.cityRepo.Create(ctx, city)
}

// GetByID retrieves a city by ID
func (uc *cityUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.City, error) {
	return uc.cityRepo.GetByID(ctx, id)
}

// List retrieves all cities with pagination
func (uc *cityUseCase) List(ctx context.Context, page, pageSize int) ([]entity.City, int, error) {
	offset := (page - 1) * pageSize
	return uc.cityRepo.List(ctx, offset, pageSize)
}

// ListByCountry retrieves cities for a specific country
func (uc *cityUseCase) ListByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.City, int, error) {
	offset := (page - 1) * pageSize
	return uc.cityRepo.ListByCountry(ctx, countryID, offset, pageSize)
}

// ListByRegion retrieves cities for a specific region
func (uc *cityUseCase) ListByRegion(ctx context.Context, regionID uuidv7.UUID, page, pageSize int) ([]entity.City, int, error) {
	offset := (page - 1) * pageSize
	return uc.cityRepo.ListByRegion(ctx, regionID, offset, pageSize)
}

// ListActive retrieves only active cities
func (uc *cityUseCase) ListActive(ctx context.Context, page, pageSize int) ([]entity.City, int, error) {
	offset := (page - 1) * pageSize
	return uc.cityRepo.ListActive(ctx, offset, pageSize)
}

// ListActiveByCountry retrieves active cities for a specific country
func (uc *cityUseCase) ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.City, int, error) {
	offset := (page - 1) * pageSize
	return uc.cityRepo.ListActiveByCountry(ctx, countryID, offset, pageSize)
}

// ListCapitals retrieves all capital cities
func (uc *cityUseCase) ListCapitals(ctx context.Context, page, pageSize int) ([]entity.City, int, error) {
	offset := (page - 1) * pageSize
	return uc.cityRepo.ListCapitals(ctx, offset, pageSize)
}

// SearchByName searches cities by name
func (uc *cityUseCase) SearchByName(ctx context.Context, query string, page, pageSize int) ([]entity.City, int, error) {
	if query == "" {
		return nil, 0, fmt.Errorf("search query cannot be empty")
	}
	offset := (page - 1) * pageSize
	return uc.cityRepo.SearchByName(ctx, query, offset, pageSize)
}

// Update updates an existing city
func (uc *cityUseCase) Update(ctx context.Context, city *entity.City) error {
	// Validate entity
	if err := city.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.cityRepo.Update(ctx, city)
}

// Delete deletes a city by ID
func (uc *cityUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.cityRepo.Delete(ctx, id)
}
