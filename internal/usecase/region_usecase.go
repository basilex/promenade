package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRegionUseCase interface defines operations for region management
type IRegionUseCase interface {
	Create(ctx context.Context, region *entity.Region) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Region, error)
	GetByCode(ctx context.Context, countryID uuidv7.UUID, code string) (*entity.Region, error)
	List(ctx context.Context, page, pageSize int) ([]entity.Region, int, error)
	ListByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.Region, int, error)
	ListActive(ctx context.Context, page, pageSize int) ([]entity.Region, int, error)
	ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.Region, int, error)
	Update(ctx context.Context, region *entity.Region) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type regionUseCase struct {
	regionRepo repository.IRegionRepository
}

func NewRegionUseCase(regionRepo repository.IRegionRepository) IRegionUseCase {
	return &regionUseCase{
		regionRepo: regionRepo,
	}
}

// Create creates a new region
func (uc *regionUseCase) Create(ctx context.Context, region *entity.Region) error {
	// Validate entity
	if err := region.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.regionRepo.Create(ctx, region)
}

// GetByID retrieves a region by ID
func (uc *regionUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Region, error) {
	return uc.regionRepo.GetByID(ctx, id)
}

// GetByCode retrieves a region by country and code
func (uc *regionUseCase) GetByCode(ctx context.Context, countryID uuidv7.UUID, code string) (*entity.Region, error) {
	return uc.regionRepo.GetByCode(ctx, countryID, code)
}

// List retrieves all regions with pagination
func (uc *regionUseCase) List(ctx context.Context, page, pageSize int) ([]entity.Region, int, error) {
	offset := (page - 1) * pageSize
	return uc.regionRepo.List(ctx, offset, pageSize)
}

// ListByCountry retrieves regions for a specific country
func (uc *regionUseCase) ListByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.Region, int, error) {
	offset := (page - 1) * pageSize
	return uc.regionRepo.ListByCountry(ctx, countryID, offset, pageSize)
}

// ListActive retrieves only active regions
func (uc *regionUseCase) ListActive(ctx context.Context, page, pageSize int) ([]entity.Region, int, error) {
	offset := (page - 1) * pageSize
	return uc.regionRepo.ListActive(ctx, offset, pageSize)
}

// ListActiveByCountry retrieves active regions for a specific country
func (uc *regionUseCase) ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, page, pageSize int) ([]entity.Region, int, error) {
	offset := (page - 1) * pageSize
	return uc.regionRepo.ListActiveByCountry(ctx, countryID, offset, pageSize)
}

// Update updates an existing region
func (uc *regionUseCase) Update(ctx context.Context, region *entity.Region) error {
	// Validate entity
	if err := region.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.regionRepo.Update(ctx, region)
}

// Delete deletes a region by ID
func (uc *regionUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.regionRepo.Delete(ctx, id)
}
