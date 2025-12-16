package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type currencyRepository struct {
	*BaseRepository
}

// NewCurrencyRepository creates a new currency repository
func NewCurrencyRepository(db *sqlx.DB) repository.CurrencyRepository {
	return &currencyRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new currency
func (r *currencyRepository) Create(ctx context.Context, currency *entity.Currency) error {
	query := `
		INSERT INTO currencies (name, code, symbol)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowxContext(ctx, query,
		currency.Name,
		currency.Code,
		currency.Symbol,
	).Scan(&currency.ID, &currency.CreatedAt, &currency.UpdatedAt)
}

// GetByID retrieves a currency by ID
func (r *currencyRepository) GetByID(ctx context.Context, id uuid.UUID, withCountries bool) (*entity.Currency, error) {
	query := `
		SELECT id, name, code, symbol, created_at, updated_at
		FROM currencies
		WHERE id = $1
	`

	var currency entity.Currency
	if err := r.db.GetContext(ctx, &currency, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	if withCountries {
		countries, err := r.GetCountries(ctx, currency.ID)
		if err != nil {
			return nil, err
		}
		currency.Countries = countries
	}

	return &currency, nil
}

// GetByCode retrieves a currency by code
func (r *currencyRepository) GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error) {
	query := `
		SELECT id, name, code, symbol, created_at, updated_at
		FROM currencies
		WHERE code = $1
	`

	var currency entity.Currency
	if err := r.db.GetContext(ctx, &currency, query, code); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	if withCountries {
		countries, err := r.GetCountries(ctx, currency.ID)
		if err != nil {
			return nil, err
		}
		currency.Countries = countries
	}

	return &currency, nil
}

// List retrieves all currencies with pagination
func (r *currencyRepository) List(ctx context.Context, offset, limit int, withCountries bool) ([]entity.Currency, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM currencies`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	// Get currencies
	query := `
		SELECT id, name, code, symbol, created_at, updated_at
		FROM currencies
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	var currencies []entity.Currency
	if err := r.db.SelectContext(ctx, &currencies, query, limit, offset); err != nil {
		return nil, 0, err
	}

	if withCountries {
		for i := range currencies {
			countries, err := r.GetCountries(ctx, currencies[i].ID)
			if err != nil {
				return nil, 0, err
			}
			currencies[i].Countries = countries
		}
	}

	return currencies, total, nil
}

// Update updates a currency
func (r *currencyRepository) Update(ctx context.Context, currency *entity.Currency) error {
	query := `
		UPDATE currencies
		SET name = $1, code = $2, symbol = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRowxContext(ctx, query,
		currency.Name,
		currency.Code,
		currency.Symbol,
		currency.ID,
	).Scan(&currency.UpdatedAt)

	if err == sql.ErrNoRows {
		return entity.ErrNotFound
	}

	return err
}

// Delete deletes a currency
func (r *currencyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM currencies WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// AddCountry adds a country to a currency
func (r *currencyRepository) AddCountry(ctx context.Context, currencyID, countryID uuid.UUID, isPrimary bool) error {
	query := `
		INSERT INTO country_currencies (country_id, currency_id, is_primary)
		VALUES ($1, $2, $3)
		ON CONFLICT (country_id, currency_id) DO UPDATE
		SET is_primary = EXCLUDED.is_primary
	`

	_, err := r.db.ExecContext(ctx, query, countryID, currencyID, isPrimary)
	return err
}

// RemoveCountry removes a country from a currency
func (r *currencyRepository) RemoveCountry(ctx context.Context, currencyID, countryID uuid.UUID) error {
	query := `DELETE FROM country_currencies WHERE currency_id = $1 AND country_id = $2`
	result, err := r.db.ExecContext(ctx, query, currencyID, countryID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("relationship not found")
	}

	return nil
}

// GetCountries gets all countries for a currency
func (r *currencyRepository) GetCountries(ctx context.Context, currencyID uuid.UUID) ([]entity.Country, error) {
	query := `
		SELECT c.id, c.name, c.code, c.iso2, c.iso3, c.created_at, c.updated_at
		FROM countries c
		INNER JOIN country_currencies cc ON c.id = cc.country_id
		WHERE cc.currency_id = $1
		ORDER BY cc.is_primary DESC, c.name ASC
	`

	var countries []entity.Country
	if err := r.db.SelectContext(ctx, &countries, query, currencyID); err != nil {
		return nil, err
	}

	return countries, nil
}
