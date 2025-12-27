package timezone

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Timezone data access
type IRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Timezone, error)
	GetByName(ctx context.Context, name string) (*Timezone, error)
	List(ctx context.Context) ([]*Timezone, error)
	Create(ctx context.Context, timezone *Timezone) error
	Update(ctx context.Context, timezone *Timezone) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
