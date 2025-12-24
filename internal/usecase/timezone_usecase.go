package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ITimezoneUseCase interface defines operations for timezone management
type ITimezoneUseCase interface {
	Create(ctx context.Context, timezone *entity.Timezone) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Timezone, error)
	GetByName(ctx context.Context, name string) (*entity.Timezone, error)
	List(ctx context.Context, params pagination.Params) ([]*entity.Timezone, *pagination.Metadata, error)
	ListActive(ctx context.Context) ([]*entity.Timezone, error)
	Update(ctx context.Context, timezone *entity.Timezone) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type timezoneUseCase struct {
	timezoneRepo repository.ITimezoneRepository
}

// NewTimezoneUseCase creates a new timezone use case
func NewTimezoneUseCase(timezoneRepo repository.ITimezoneRepository) ITimezoneUseCase {
	return &timezoneUseCase{
		timezoneRepo: timezoneRepo,
	}
}

// Create creates a new timezone
func (uc *timezoneUseCase) Create(ctx context.Context, timezone *entity.Timezone) error {
	// Validate entity
	if err := timezone.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if timezone with this name already exists
	exists, err := uc.timezoneRepo.Exists(ctx, timezone.Name)
	if err != nil {
		return fmt.Errorf("failed to check timezone existence: %w", err)
	}
	if exists {
		return fmt.Errorf("timezone with name '%s' already exists", timezone.Name)
	}

	// Generate UUID if not set
	if timezone.ID == (uuidv7.UUID{}) {
		timezone.ID = uuidv7.New()
	}

	return uc.timezoneRepo.Create(ctx, timezone)
}

// GetByID retrieves a timezone by ID
func (uc *timezoneUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Timezone, error) {
	return uc.timezoneRepo.GetByID(ctx, id)
}

// GetByName retrieves a timezone by name (e.g., "Europe/Moscow")
func (uc *timezoneUseCase) GetByName(ctx context.Context, name string) (*entity.Timezone, error) {
	return uc.timezoneRepo.GetByName(ctx, name)
}

// List retrieves all timezones with pagination
func (uc *timezoneUseCase) List(ctx context.Context, params pagination.Params) ([]*entity.Timezone, *pagination.Metadata, error) {
	// Apply default pagination if not set
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	return uc.timezoneRepo.List(ctx, params)
}

// ListActive retrieves all active timezones
func (uc *timezoneUseCase) ListActive(ctx context.Context) ([]*entity.Timezone, error) {
	return uc.timezoneRepo.ListActive(ctx)
}

// Update updates an existing timezone
func (uc *timezoneUseCase) Update(ctx context.Context, timezone *entity.Timezone) error {
	// Check if timezone exists
	existing, err := uc.timezoneRepo.GetByID(ctx, timezone.ID)
	if err != nil {
		return err
	}

	// If name is being changed, check if new name already exists
	if existing.Name != timezone.Name {
		exists, err := uc.timezoneRepo.Exists(ctx, timezone.Name)
		if err != nil {
			return fmt.Errorf("failed to check timezone existence: %w", err)
		}
		if exists {
			return fmt.Errorf("timezone with name '%s' already exists", timezone.Name)
		}
	}

	// Validate entity
	if err := timezone.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.timezoneRepo.Update(ctx, timezone)
}

// Delete deletes a timezone by ID
func (uc *timezoneUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Check if timezone exists
	_, err := uc.timezoneRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.timezoneRepo.Delete(ctx, id)
}
