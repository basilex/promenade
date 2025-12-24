package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ILanguageRepository defines the interface for language data operations
type ILanguageRepository interface {
	// Create creates a new language
	Create(ctx context.Context, language *entity.Language) error

	// GetByID retrieves a language by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Language, error)

	// GetByCode retrieves a language by ISO 639-1 code (e.g., "en", "ru")
	GetByCode(ctx context.Context, code string) (*entity.Language, error)

	// List retrieves all languages with pagination
	List(ctx context.Context, params pagination.Params) ([]*entity.Language, *pagination.Metadata, error)

	// ListActive retrieves all active languages
	ListActive(ctx context.Context) ([]*entity.Language, error)

	// Update updates an existing language
	Update(ctx context.Context, language *entity.Language) error

	// Delete deletes a language by ID
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Exists checks if a language exists by code
	Exists(ctx context.Context, code string) (bool, error)
}
