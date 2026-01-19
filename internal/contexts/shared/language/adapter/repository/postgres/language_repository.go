package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	languageerrors "github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/language/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type languageRepository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL language repository
func NewRepository(db sqlx.ExtContext) repository.ILanguageRepository {
	return &languageRepository{db: db}
}

func (r *languageRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Language, error) {
	var l aggregate.Language
	query := `
		SELECT id, code, code3, name, native_name, is_active, created_at, updated_at
		FROM shared_languages
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &l, query, id)
	if err == sql.ErrNoRows {
		return nil, languageerrors.ErrLanguageNotFound
	}
	return &l, err
}

func (r *languageRepository) GetByCode(ctx context.Context, code string) (*aggregate.Language, error) {
	var l aggregate.Language
	query := `
		SELECT id, code, code3, name, native_name, is_active, created_at, updated_at
		FROM shared_languages
		WHERE code = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &l, query, code)
	if err == sql.ErrNoRows {
		return nil, languageerrors.ErrLanguageNotFound
	}
	return &l, err
}

func (r *languageRepository) List(ctx context.Context) ([]*aggregate.Language, error) {
	var languages []*aggregate.Language
	query := `
		SELECT id, code, code3, name, native_name, is_active, created_at, updated_at
		FROM shared_languages
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &languages, query)
	return languages, err
}

func (r *languageRepository) Create(ctx context.Context, l *aggregate.Language) error {
	query := `
		INSERT INTO shared_languages (id, code, code3, name, native_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.Code, l.Code3, l.Name, l.NativeName, l.IsActive, l.CreatedAt, l.UpdatedAt)
	return err
}

func (r *languageRepository) Update(ctx context.Context, l *aggregate.Language) error {
	query := `
		UPDATE shared_languages
		SET code = $2, code3 = $3, name = $4, native_name = $5, is_active = $6, updated_at = $7
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.Code, l.Code3, l.Name, l.NativeName, l.IsActive, l.UpdatedAt)
	return err
}

func (r *languageRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_languages SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
