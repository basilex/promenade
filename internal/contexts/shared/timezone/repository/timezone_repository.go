package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/shared/timezone/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ITimezoneRepository defines the interface for Timezone data access
type ITimezoneRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Timezone, error)
	GetByName(ctx context.Context, name string) (*aggregate.Timezone, error)
	List(ctx context.Context) ([]*aggregate.Timezone, error)
	Create(ctx context.Context, timezone *aggregate.Timezone) error
	Update(ctx context.Context, timezone *aggregate.Timezone) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
