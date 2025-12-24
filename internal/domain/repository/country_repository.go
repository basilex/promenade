package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICountryRepository defines the interface for country data access
type ICountryRepository interface {
	// Create creates a new country
	Create(ctx context.Context, country *entity.Country) error

	// GetByID retrieves a country by ID with optional currencies
	GetByID(ctx context.Context, id uuidv7.UUID, withCurrencies bool) (*entity.Country, error)

	// GetByCode retrieves a country by code (iso2 or iso3) with optional currencies
	GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error)

	// List retrieves all countries with pagination and optional currencies
	List(ctx context.Context, offset, limit int, withCurrencies bool) ([]entity.Country, int, error)

	// ListByRegion retrieves countries by region with pagination and optional currencies
	ListByRegion(ctx context.Context, region string, offset, limit int, withCurrencies bool) ([]entity.Country, int, error)

	// Update updates a country
	Update(ctx context.Context, country *entity.Country) error

	// Delete deletes a country
	Delete(ctx context.Context, id uuidv7.UUID) error

	// AddCurrency adds a currency to a country
	AddCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID, isPrimary bool) error

	// RemoveCurrency removes a currency from a country
	RemoveCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID) error

	// GetCurrencies gets all currencies for a country
	GetCurrencies(ctx context.Context, countryID uuidv7.UUID) ([]entity.Currency, error)
}
