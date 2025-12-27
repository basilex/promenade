package currency

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Currency data access
type IRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Currency, error)
	GetByCode(ctx context.Context, code string) (*Currency, error)
	List(ctx context.Context) ([]*Currency, error)
	Create(ctx context.Context, currency *Currency) error
	Update(ctx context.Context, currency *Currency) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
