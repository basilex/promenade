package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ITimezoneRepository defines the interface for timezone data operations
type ITimezoneRepository interface {
	// Create creates a new timezone
	Create(ctx context.Context, timezone *entity.Timezone) error

	// GetByID retrieves a timezone by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Timezone, error)

	// GetByName retrieves a timezone by name (e.g., "Europe/Moscow")
	GetByName(ctx context.Context, name string) (*entity.Timezone, error)

	// List retrieves all timezones with pagination
	List(ctx context.Context, params pagination.Params) ([]*entity.Timezone, *pagination.Metadata, error)

	// ListActive retrieves all active timezones
	ListActive(ctx context.Context) ([]*entity.Timezone, error)

	// Update updates an existing timezone
	Update(ctx context.Context, timezone *entity.Timezone) error

	// Delete deletes a timezone by ID
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Exists checks if a timezone exists by name
	Exists(ctx context.Context, name string) (bool, error)
}
