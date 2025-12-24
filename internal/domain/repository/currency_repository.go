package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICurrencyRepository defines the interface for currency data access
type ICurrencyRepository interface {
	// Create creates a new currency
	Create(ctx context.Context, currency *entity.Currency) error

	// GetByID retrieves a currency by ID with optional countries
	GetByID(ctx context.Context, id uuidv7.UUID, withCountries bool) (*entity.Currency, error)

	// GetByCode retrieves a currency by code with optional countries
	GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error)

	// List retrieves all currencies with pagination and optional countries
	List(ctx context.Context, offset, limit int, withCountries bool) ([]entity.Currency, int, error)

	// Update updates a currency
	Update(ctx context.Context, currency *entity.Currency) error

	// Delete deletes a currency
	Delete(ctx context.Context, id uuidv7.UUID) error

	// AddCountry adds a country to a currency
	AddCountry(ctx context.Context, currencyID, countryID uuidv7.UUID, isPrimary bool) error

	// RemoveCountry removes a country from a currency
	RemoveCountry(ctx context.Context, currencyID, countryID uuidv7.UUID) error

	// GetCountries gets all countries for a currency
	GetCountries(ctx context.Context, currencyID uuidv7.UUID) ([]entity.Country, error)
}
