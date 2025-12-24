package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

type languageRepository struct {
	*BaseRepository
}

// NewLanguageRepository creates a new language repository
func NewLanguageRepository(db *sqlx.DB) repository.ILanguageRepository {
	return &languageRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new language
func (r *languageRepository) Create(ctx context.Context, language *entity.Language) error {
	query := `
		INSERT INTO core_languages (id, name, native_name, code, iso639_2, is_rtl, is_active, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	executor := r.getExecutor(ctx)
	return executor.QueryRowxContext(ctx, query,
		language.ID,
		language.Name,
		language.NativeName,
		language.Code,
		language.ISO639_2,
		language.IsRtl,
		language.IsActive,
		language.SortOrder,
	).Scan(&language.CreatedAt, &language.UpdatedAt)
}

// GetByID retrieves a language by ID
func (r *languageRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Language, error) {
	query := `
		SELECT id, name, native_name, code, iso639_2, is_rtl, is_active, sort_order, created_at, updated_at
		FROM core_languages
		WHERE id = $1
	`

	var language entity.Language
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &language, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &language, nil
}

// GetByCode retrieves a language by ISO 639-1 code
func (r *languageRepository) GetByCode(ctx context.Context, code string) (*entity.Language, error) {
	query := `
		SELECT id, name, native_name, code, iso639_2, is_rtl, is_active, sort_order, created_at, updated_at
		FROM core_languages
		WHERE code = $1
	`

	var language entity.Language
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &language, query, code); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &language, nil
}

// List retrieves all languages with pagination
func (r *languageRepository) List(ctx context.Context, params pagination.Params) ([]*entity.Language, *pagination.Metadata, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_languages`
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &total, countQuery); err != nil {
		return nil, nil, err
	}

	// Get paginated languages
	query := `
		SELECT id, name, native_name, code, iso639_2, is_rtl, is_active, sort_order, created_at, updated_at
		FROM core_languages
		ORDER BY sort_order ASC, name ASC
		LIMIT $1 OFFSET $2
	`

	var languages []*entity.Language
	limit := params.GetLimit()
	offset := params.GetOffset()
	if err := sqlx.SelectContext(ctx, executor, &languages, query, limit, offset); err != nil {
		return nil, nil, err
	}

	metadata := pagination.NewMetadata(total, limit, offset)
	return languages, metadata, nil
}

// ListActive retrieves all active languages
func (r *languageRepository) ListActive(ctx context.Context) ([]*entity.Language, error) {
	query := `
		SELECT id, name, native_name, code, iso639_2, is_rtl, is_active, sort_order, created_at, updated_at
		FROM core_languages
		WHERE is_active = true
		ORDER BY sort_order ASC, name ASC
	`

	var languages []*entity.Language
	executor := r.getExecutor(ctx)
	if err := sqlx.SelectContext(ctx, executor, &languages, query); err != nil {
		return nil, err
	}

	return languages, nil
}

// Update updates an existing language
func (r *languageRepository) Update(ctx context.Context, language *entity.Language) error {
	query := `
		UPDATE core_languages
		SET name = $2, native_name = $3, code = $4, iso639_2 = $5, is_rtl = $6, is_active = $7, sort_order = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	executor := r.getExecutor(ctx)
	err := executor.QueryRowxContext(ctx, query,
		language.ID,
		language.Name,
		language.NativeName,
		language.Code,
		language.ISO639_2,
		language.IsRtl,
		language.IsActive,
		language.SortOrder,
	).Scan(&language.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return entity.ErrNotFound
		}
		return err
	}

	return nil
}

// Delete deletes a language by ID
func (r *languageRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM core_languages WHERE id = $1`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// Exists checks if a language exists by code
func (r *languageRepository) Exists(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM core_languages WHERE code = $1)`

	var exists bool
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &exists, query, code); err != nil {
		return false, fmt.Errorf("failed to check language existence: %w", err)
	}

	return exists, nil
}
