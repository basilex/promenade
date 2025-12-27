package reference

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Repository provides read-only access to reference data.
// This is intentionally read-only - reference data is managed via migrations.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new reference repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Countries

// GetCountryByID retrieves a country by ID.
func (r *Repository) GetCountryByID(ctx context.Context, id uuidv7.UUID) (*Country, error) {
	var country Country
	query := `SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active 
			  FROM shared_countries WHERE id = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &country, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get country: %w", err)
	}

	return &country, nil
}

// GetCountryByCode retrieves a country by ISO code.
func (r *Repository) GetCountryByCode(ctx context.Context, code string) (*Country, error) {
	var country Country
	query := `SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active 
			  FROM shared_countries WHERE code = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &country, query, code)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get country: %w", err)
	}

	return &country, nil
}

// ListCountries retrieves all active countries.
func (r *Repository) ListCountries(ctx context.Context) ([]Country, error) {
	var countries []Country
	query := `SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active 
			  FROM shared_countries WHERE is_active = true ORDER BY name`

	err := r.db.SelectContext(ctx, &countries, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list countries: %w", err)
	}

	return countries, nil
}

// Currencies

// GetCurrencyByID retrieves a currency by ID.
func (r *Repository) GetCurrencyByID(ctx context.Context, id uuidv7.UUID) (*Currency, error) {
	var currency Currency
	query := `SELECT id, code, numeric_code, name, symbol, decimal_places, is_active 
			  FROM shared_currencies WHERE id = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &currency, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("currency not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}

	return &currency, nil
}

// GetCurrencyByCode retrieves a currency by ISO code.
func (r *Repository) GetCurrencyByCode(ctx context.Context, code string) (*Currency, error) {
	var currency Currency
	query := `SELECT id, code, numeric_code, name, symbol, decimal_places, is_active 
			  FROM shared_currencies WHERE code = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &currency, query, code)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("currency not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}

	return &currency, nil
}

// ListCurrencies retrieves all active currencies.
func (r *Repository) ListCurrencies(ctx context.Context) ([]Currency, error) {
	var currencies []Currency
	query := `SELECT id, code, numeric_code, name, symbol, decimal_places, is_active 
			  FROM shared_currencies WHERE is_active = true ORDER BY code`

	err := r.db.SelectContext(ctx, &currencies, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}

	return currencies, nil
}

// Languages

// GetLanguageByID retrieves a language by ID.
func (r *Repository) GetLanguageByID(ctx context.Context, id uuidv7.UUID) (*Language, error) {
	var language Language
	query := `SELECT id, code, code3, name, native_name, is_active 
			  FROM shared_languages WHERE id = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &language, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("language not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %w", err)
	}

	return &language, nil
}

// GetLanguageByCode retrieves a language by ISO code.
func (r *Repository) GetLanguageByCode(ctx context.Context, code string) (*Language, error) {
	var language Language
	query := `SELECT id, code, code3, name, native_name, is_active 
			  FROM shared_languages WHERE code = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &language, query, code)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("language not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %w", err)
	}

	return &language, nil
}

// ListLanguages retrieves all active languages.
func (r *Repository) ListLanguages(ctx context.Context) ([]Language, error) {
	var languages []Language
	query := `SELECT id, code, code3, name, native_name, is_active 
			  FROM shared_languages WHERE is_active = true ORDER BY name`

	err := r.db.SelectContext(ctx, &languages, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list languages: %w", err)
	}

	return languages, nil
}

// Timezones

// GetTimezoneByID retrieves a timezone by ID.
func (r *Repository) GetTimezoneByID(ctx context.Context, id uuidv7.UUID) (*Timezone, error) {
	var timezone Timezone
	query := `SELECT id, name, abbreviation, utc_offset, is_active 
			  FROM shared_timezones WHERE id = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &timezone, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("timezone not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get timezone: %w", err)
	}

	return &timezone, nil
}

// GetTimezoneByName retrieves a timezone by IANA name.
func (r *Repository) GetTimezoneByName(ctx context.Context, name string) (*Timezone, error) {
	var timezone Timezone
	query := `SELECT id, name, abbreviation, utc_offset, is_active 
			  FROM shared_timezones WHERE name = $1 AND is_active = true`

	err := r.db.GetContext(ctx, &timezone, query, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("timezone not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get timezone: %w", err)
	}

	return &timezone, nil
}

// ListTimezones retrieves all active timezones.
func (r *Repository) ListTimezones(ctx context.Context) ([]Timezone, error) {
	var timezones []Timezone
	query := `SELECT id, name, abbreviation, utc_offset, is_active 
			  FROM shared_timezones WHERE is_active = true ORDER BY name`

	err := r.db.SelectContext(ctx, &timezones, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list timezones: %w", err)
	}

	return timezones, nil
}
