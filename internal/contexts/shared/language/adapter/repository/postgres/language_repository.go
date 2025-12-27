package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type repository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL language repository
func NewRepository(db sqlx.ExtContext) language.IRepository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uuidv7.UUID) (*language.Language, error) {
	var l language.Language
	query := `
		SELECT id, code, code3, name, native_name, is_active
		FROM shared_languages
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &l, query, id)
	if err == sql.ErrNoRows {
		return nil, language.ErrNotFound
	}
	return &l, err
}

func (r *repository) GetByCode(ctx context.Context, code string) (*language.Language, error) {
	var l language.Language
	query := `
		SELECT id, code, code3, name, native_name, is_active
		FROM shared_languages
		WHERE code = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &l, query, code)
	if err == sql.ErrNoRows {
		return nil, language.ErrNotFound
	}
	return &l, err
}

func (r *repository) List(ctx context.Context) ([]*language.Language, error) {
	var languages []*language.Language
	query := `
		SELECT id, code, code3, name, native_name, is_active
		FROM shared_languages
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &languages, query)
	return languages, err
}

func (r *repository) Create(ctx context.Context, l *language.Language) error {
	query := `
		INSERT INTO shared_languages (id, code, code3, name, native_name, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.Code, l.Code3, l.Name, l.NativeName, l.IsActive)
	return err
}

func (r *repository) Update(ctx context.Context, l *language.Language) error {
	query := `
		UPDATE shared_languages
		SET code = $2, code3 = $3, name = $4, native_name = $5, is_active = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.Code, l.Code3, l.Name, l.NativeName, l.IsActive)
	return err
}

func (r *repository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_languages SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
