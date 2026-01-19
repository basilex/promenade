package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/ui/metadata/form/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IFormRepository defines persistence operations for form definitions.
type IFormRepository interface {
	Create(ctx context.Context, form *aggregate.FormDefinition) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error)
	GetByFormID(ctx context.Context, formID string) (*aggregate.FormDefinition, error)
	Update(ctx context.Context, form *aggregate.FormDefinition) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error)
}
