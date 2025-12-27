package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type repository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL currency repository
func NewRepository(db sqlx.ExtContext) currency.IRepository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uuidv7.UUID) (*currency.Currency, error) {
	var c currency.Currency
	query := `
		SELECT id, code, numeric_code, name, symbol, decimal_places, is_active
		FROM shared_currencies
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, id)
	if err == sql.ErrNoRows {
		return nil, currency.ErrNotFound
	}
	return &c, err
}

func (r *repository) GetByCode(ctx context.Context, code string) (*currency.Currency, error) {
	var c currency.Currency
	query := `
		SELECT id, code, numeric_code, name, symbol, decimal_places, is_active
		FROM shared_currencies
		WHERE code = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, code)
	if err == sql.ErrNoRows {
		return nil, currency.ErrNotFound
	}
	return &c, err
}

func (r *repository) List(ctx context.Context) ([]*currency.Currency, error) {
	var currencies []*currency.Currency
	query := `
		SELECT id, code, numeric_code, name, symbol, decimal_places, is_active
		FROM shared_currencies
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &currencies, query)
	return currencies, err
}

func (r *repository) Create(ctx context.Context, c *currency.Currency) error {
	query := `
		INSERT INTO shared_currencies (id, code, numeric_code, name, symbol, decimal_places, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.NumericCode, c.Name, c.Symbol, c.DecimalPlaces, c.IsActive)
	return err
}

func (r *repository) Update(ctx context.Context, c *currency.Currency) error {
	query := `
		UPDATE shared_currencies
		SET code = $2, numeric_code = $3, name = $4, symbol = $5, decimal_places = $6, is_active = $7
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.NumericCode, c.Name, c.Symbol, c.DecimalPlaces, c.IsActive)
	return err
}

func (r *repository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_currencies SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
