package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ILanguageRepository defines the interface for Language data access
type ILanguageRepository interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Language, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Language, error)
	List(ctx context.Context) ([]*aggregate.Language, error)
	Create(ctx context.Context, language *aggregate.Language) error
	Update(ctx context.Context, language *aggregate.Language) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}
