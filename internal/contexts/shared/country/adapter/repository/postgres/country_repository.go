package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type repository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL country repository
func NewRepository(db sqlx.ExtContext) country.IRepository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uuidv7.UUID) (*country.Country, error) {
	var c country.Country
	query := `
		SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active
		FROM shared_countries
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, id)
	if err == sql.ErrNoRows {
		return nil, country.ErrNotFound
	}
	return &c, err
}

func (r *repository) GetByCode(ctx context.Context, code string) (*country.Country, error) {
	var c country.Country
	query := `
		SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active
		FROM shared_countries
		WHERE code = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, code)
	if err == sql.ErrNoRows {
		return nil, country.ErrNotFound
	}
	return &c, err
}

func (r *repository) List(ctx context.Context) ([]*country.Country, error) {
	var countries []*country.Country
	query := `
		SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active
		FROM shared_countries
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &countries, query)
	return countries, err
}

func (r *repository) Create(ctx context.Context, c *country.Country) error {
	query := `
		INSERT INTO shared_countries (id, code, code3, numeric_code, name, name_local, phone_code, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.Code3, c.NumericCode, c.Name, c.NameLocal, c.PhoneCode, c.IsActive)
	return err
}

func (r *repository) Update(ctx context.Context, c *country.Country) error {
	query := `
		UPDATE shared_countries
		SET code = $2, code3 = $3, numeric_code = $4, name = $5, name_local = $6, phone_code = $7, is_active = $8
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.Code3, c.NumericCode, c.Name, c.NameLocal, c.PhoneCode, c.IsActive)
	return err
}

func (r *repository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_countries SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
