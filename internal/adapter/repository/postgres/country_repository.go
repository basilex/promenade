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

type countryRepository struct {
	*BaseRepository
}

// NewCountryRepository creates a new country repository
func NewCountryRepository(db *sqlx.DB) repository.CountryRepository {
	return &countryRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new country
func (r *countryRepository) Create(ctx context.Context, country *entity.Country) error {
	query := `
		INSERT INTO countries (name, code, iso2, iso3, region)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowxContext(ctx, query,
		country.Name,
		country.Code,
		country.ISO2,
		country.ISO3,
		country.Region,
	).Scan(&country.ID, &country.CreatedAt, &country.UpdatedAt)
}

// GetByID retrieves a country by ID
func (r *countryRepository) GetByID(ctx context.Context, id uuid.UUID, withCurrencies bool) (*entity.Country, error) {
	query := `
		SELECT id, name, code, iso2, iso3, region, created_at, updated_at
		FROM countries
		WHERE id = $1
	`

	var country entity.Country
	if err := r.db.GetContext(ctx, &country, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	if withCurrencies {
		currencies, err := r.GetCurrencies(ctx, country.ID)
		if err != nil {
			return nil, err
		}
		country.Currencies = currencies
	}

	return &country, nil
}

// GetByCode retrieves a country by code (iso2 or iso3)
func (r *countryRepository) GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error) {
	query := `
		SELECT id, name, code, iso2, iso3, region, created_at, updated_at
		FROM countries
		WHERE code = $1 OR iso2 = $1 OR iso3 = $1
	`

	var country entity.Country
	if err := r.db.GetContext(ctx, &country, query, code); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	if withCurrencies {
		currencies, err := r.GetCurrencies(ctx, country.ID)
		if err != nil {
			return nil, err
		}
		country.Currencies = currencies
	}

	return &country, nil
}

// List retrieves all countries with pagination
func (r *countryRepository) List(ctx context.Context, offset, limit int, withCurrencies bool) ([]entity.Country, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM countries`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	// Get countries
	query := `
		SELECT id, name, code, iso2, iso3, region, created_at, updated_at
		FROM countries
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	var countries []entity.Country
	if err := r.db.SelectContext(ctx, &countries, query, limit, offset); err != nil {
		return nil, 0, err
	}

	if withCurrencies {
		for i := range countries {
			currencies, err := r.GetCurrencies(ctx, countries[i].ID)
			if err != nil {
				return nil, 0, err
			}
			countries[i].Currencies = currencies
		}
	}

	return countries, total, nil
}

// ListByRegion retrieves countries by region with pagination
func (r *countryRepository) ListByRegion(ctx context.Context, region string, offset, limit int, withCurrencies bool) ([]entity.Country, int, error) {
	// Get total count for region
	var total int
	countQuery := `SELECT COUNT(*) FROM countries WHERE region = $1`
	if err := r.db.GetContext(ctx, &total, countQuery, region); err != nil {
		return nil, 0, err
	}

	// Get countries by region
	query := `
		SELECT id, name, code, iso2, iso3, region, created_at, updated_at
		FROM countries
		WHERE region = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	var countries []entity.Country
	if err := r.db.SelectContext(ctx, &countries, query, region, limit, offset); err != nil {
		return nil, 0, err
	}

	if withCurrencies {
		for i := range countries {
			currencies, err := r.GetCurrencies(ctx, countries[i].ID)
			if err != nil {
				return nil, 0, err
			}
			countries[i].Currencies = currencies
		}
	}

	return countries, total, nil
}

// Update updates a country
func (r *countryRepository) Update(ctx context.Context, country *entity.Country) error {
	query := `
		UPDATE countries
		SET name = $1, code = $2, iso2 = $3, iso3 = $4, region = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRowxContext(ctx, query,
		country.Name,
		country.Code,
		country.ISO2,
		country.ISO3,
		country.Region,
		country.ID,
	).Scan(&country.UpdatedAt)

	if err == sql.ErrNoRows {
		return entity.ErrNotFound
	}

	return err
}

// Delete deletes a country
func (r *countryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM countries WHERE id = $1`
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

// AddCurrency adds a currency to a country
func (r *countryRepository) AddCurrency(ctx context.Context, countryID, currencyID uuid.UUID, isPrimary bool) error {
	query := `
		INSERT INTO country_currencies (country_id, currency_id, is_primary)
		VALUES ($1, $2, $3)
		ON CONFLICT (country_id, currency_id) DO UPDATE
		SET is_primary = EXCLUDED.is_primary
	`

	_, err := r.db.ExecContext(ctx, query, countryID, currencyID, isPrimary)
	return err
}

// RemoveCurrency removes a currency from a country
func (r *countryRepository) RemoveCurrency(ctx context.Context, countryID, currencyID uuid.UUID) error {
	query := `DELETE FROM country_currencies WHERE country_id = $1 AND currency_id = $2`
	result, err := r.db.ExecContext(ctx, query, countryID, currencyID)
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

// GetCurrencies gets all currencies for a country
func (r *countryRepository) GetCurrencies(ctx context.Context, countryID uuid.UUID) ([]entity.Currency, error) {
	query := `
		SELECT c.id, c.name, c.code, c.symbol, c.created_at, c.updated_at
		FROM currencies c
		INNER JOIN country_currencies cc ON c.id = cc.currency_id
		WHERE cc.country_id = $1
		ORDER BY cc.is_primary DESC, c.name ASC
	`

	var currencies []entity.Currency
	if err := r.db.SelectContext(ctx, &currencies, query, countryID); err != nil {
		return nil, err
	}

	return currencies, nil
}
