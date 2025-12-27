package country

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Country data access
type IRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Country, error)
	GetByCode(ctx context.Context, code string) (*Country, error)
	List(ctx context.Context) ([]*Country, error)
	Create(ctx context.Context, country *Country) error
	Update(ctx context.Context, country *Country) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
