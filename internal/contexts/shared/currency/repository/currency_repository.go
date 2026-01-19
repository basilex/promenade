package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/shared/currency/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICurrencyRepository defines the interface for Currency data access
type ICurrencyRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Currency, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Currency, error)
	List(ctx context.Context) ([]*aggregate.Currency, error)
	Create(ctx context.Context, currency *aggregate.Currency) error
	Update(ctx context.Context, currency *aggregate.Currency) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
