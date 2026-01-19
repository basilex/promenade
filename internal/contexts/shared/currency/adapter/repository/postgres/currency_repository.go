package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	currencyerrors "github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/internal/contexts/shared/currency/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/currency/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type currencyRepository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL currency repository
func NewRepository(db sqlx.ExtContext) repository.ICurrencyRepository {
	return &currencyRepository{db: db}
}

func (r *currencyRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Currency, error) {
	var c aggregate.Currency
	query := `
		SELECT id, code, numeric_code, name, symbol, decimal_places, is_active, created_at, updated_at
		FROM shared_currencies
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, id)
	if err == sql.ErrNoRows {
		return nil, currencyerrors.ErrCurrencyNotFound
	}
	return &c, err
}

func (r *currencyRepository) GetByCode(ctx context.Context, code string) (*aggregate.Currency, error) {
	var c aggregate.Currency
	query := `
		SELECT id, code, numeric_code, name, symbol, decimal_places, is_active, created_at, updated_at
		FROM shared_currencies
		WHERE code = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, code)
	if err == sql.ErrNoRows {
		return nil, currencyerrors.ErrCurrencyNotFound
	}
	return &c, err
}

func (r *currencyRepository) List(ctx context.Context) ([]*aggregate.Currency, error) {
	var currencies []*aggregate.Currency
	query := `
		SELECT id, code, numeric_code, name, symbol, decimal_places, is_active, created_at, updated_at
		FROM shared_currencies
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &currencies, query)
	return currencies, err
}

func (r *currencyRepository) Create(ctx context.Context, c *aggregate.Currency) error {
	query := `
		INSERT INTO shared_currencies (id, code, numeric_code, name, symbol, decimal_places, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.NumericCode, c.Name, c.Symbol, c.DecimalPlaces, c.IsActive, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *currencyRepository) Update(ctx context.Context, c *aggregate.Currency) error {
	query := `
		UPDATE shared_currencies
		SET code = $2, numeric_code = $3, name = $4, symbol = $5, decimal_places = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.NumericCode, c.Name, c.Symbol, c.DecimalPlaces, c.IsActive, c.UpdatedAt)
	return err
}

func (r *currencyRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_currencies SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
