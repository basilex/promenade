package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// LanguageUseCase interface defines operations for language management
type LanguageUseCase interface {
	Create(ctx context.Context, language *entity.Language) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Language, error)
	GetByCode(ctx context.Context, code string) (*entity.Language, error)
	List(ctx context.Context, params pagination.Params) ([]*entity.Language, *pagination.Metadata, error)
	ListActive(ctx context.Context) ([]*entity.Language, error)
	Update(ctx context.Context, language *entity.Language) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type languageUseCase struct {
	languageRepo repository.LanguageRepository
}

// NewLanguageUseCase creates a new language use case
func NewLanguageUseCase(languageRepo repository.LanguageRepository) LanguageUseCase {
	return &languageUseCase{
		languageRepo: languageRepo,
	}
}

// Create creates a new language
func (uc *languageUseCase) Create(ctx context.Context, language *entity.Language) error {
	// Validate entity
	if err := language.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if language with this code already exists
	exists, err := uc.languageRepo.Exists(ctx, language.Code)
	if err != nil {
		return fmt.Errorf("failed to check language existence: %w", err)
	}
	if exists {
		return fmt.Errorf("language with code '%s' already exists", language.Code)
	}

	// Generate UUID if not set
	if language.ID == (uuidv7.UUID{}) {
		language.ID = uuidv7.New()
	}

	return uc.languageRepo.Create(ctx, language)
}

// GetByID retrieves a language by ID
func (uc *languageUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Language, error) {
	return uc.languageRepo.GetByID(ctx, id)
}

// GetByCode retrieves a language by ISO 639-1 code
func (uc *languageUseCase) GetByCode(ctx context.Context, code string) (*entity.Language, error) {
	return uc.languageRepo.GetByCode(ctx, code)
}

// List retrieves all languages with pagination
func (uc *languageUseCase) List(ctx context.Context, params pagination.Params) ([]*entity.Language, *pagination.Metadata, error) {
	// Apply default pagination if not set
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	return uc.languageRepo.List(ctx, params)
}

// ListActive retrieves all active languages
func (uc *languageUseCase) ListActive(ctx context.Context) ([]*entity.Language, error) {
	return uc.languageRepo.ListActive(ctx)
}

// Update updates an existing language
func (uc *languageUseCase) Update(ctx context.Context, language *entity.Language) error {
	// Check if language exists
	existing, err := uc.languageRepo.GetByID(ctx, language.ID)
	if err != nil {
		return err
	}

	// If code is being changed, check if new code already exists
	if existing.Code != language.Code {
		exists, err := uc.languageRepo.Exists(ctx, language.Code)
		if err != nil {
			return fmt.Errorf("failed to check language existence: %w", err)
		}
		if exists {
			return fmt.Errorf("language with code '%s' already exists", language.Code)
		}
	}

	// Validate entity
	if err := language.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.languageRepo.Update(ctx, language)
}

// Delete deletes a language by ID
func (uc *languageUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Check if language exists
	_, err := uc.languageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.languageRepo.Delete(ctx, id)
}
