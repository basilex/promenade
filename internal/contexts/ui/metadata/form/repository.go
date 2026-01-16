package form

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines persistence operations for form definitions.
type IRepository interface {
	Create(ctx context.Context, form *FormDefinition) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*FormDefinition, error)
	GetByFormID(ctx context.Context, formID string) (*FormDefinition, error)
	Update(ctx context.Context, form *FormDefinition) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, entityType string, limit, offset int) ([]*FormDefinition, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*FormDefinition, int, error)
}
