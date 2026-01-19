package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Country data access
type ICountryRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Country, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Country, error)
	List(ctx context.Context) ([]*aggregate.Country, error)
	Create(ctx context.Context, country *aggregate.Country) error
	Update(ctx context.Context, country *aggregate.Country) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
