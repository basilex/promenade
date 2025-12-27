package language

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Language data access
type IRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Language, error)
	GetByCode(ctx context.Context, code string) (*Language, error)
	List(ctx context.Context) ([]*Language, error)
	Create(ctx context.Context, language *Language) error
	Update(ctx context.Context, language *Language) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
