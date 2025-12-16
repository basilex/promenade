package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/google/uuid"
)

type CountryUseCase struct {
	countryRepo  repository.CountryRepository
	currencyRepo repository.CurrencyRepository
}

func NewCountryUseCase(
	countryRepo repository.CountryRepository,
	currencyRepo repository.CurrencyRepository,
) *CountryUseCase {
	return &CountryUseCase{
		countryRepo:  countryRepo,
		currencyRepo: currencyRepo,
	}
}

// Create creates a new country
func (uc *CountryUseCase) Create(ctx context.Context, country *entity.Country) error {
	// Validate entity
	if err := country.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.countryRepo.Create(ctx, country)
}

// GetByID retrieves a country by ID
func (uc *CountryUseCase) GetByID(ctx context.Context, id uuid.UUID, withCurrencies bool) (*entity.Country, error) {
	return uc.countryRepo.GetByID(ctx, id, withCurrencies)
}

// GetByCode retrieves a country by code
func (uc *CountryUseCase) GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error) {
	return uc.countryRepo.GetByCode(ctx, code, withCurrencies)
}

// List retrieves all countries with pagination
func (uc *CountryUseCase) List(ctx context.Context, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return uc.countryRepo.List(ctx, offset, pageSize, withCurrencies)
}

// ListByRegion retrieves countries by region with pagination
func (uc *CountryUseCase) ListByRegion(ctx context.Context, region string, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return uc.countryRepo.ListByRegion(ctx, region, offset, pageSize, withCurrencies)
}

// Update updates a country
func (uc *CountryUseCase) Update(ctx context.Context, country *entity.Country) error {
	// Check if country exists
	existing, err := uc.countryRepo.GetByID(ctx, country.ID, false)
	if err != nil {
		return err
	}

	if existing == nil {
		return entity.ErrNotFound
	}

	// Validate entity
	if err := country.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.countryRepo.Update(ctx, country)
}

// Delete deletes a country
func (uc *CountryUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.countryRepo.Delete(ctx, id)
}

// AddCurrency adds a currency to a country
func (uc *CountryUseCase) AddCurrency(ctx context.Context, countryID, currencyID uuid.UUID, isPrimary bool) error {
	// Verify country exists
	_, err := uc.countryRepo.GetByID(ctx, countryID, false)
	if err != nil {
		return fmt.Errorf("country not found: %w", err)
	}

	// Verify currency exists
	_, err = uc.currencyRepo.GetByID(ctx, currencyID, false)
	if err != nil {
		return fmt.Errorf("currency not found: %w", err)
	}

	return uc.countryRepo.AddCurrency(ctx, countryID, currencyID, isPrimary)
}

// RemoveCurrency removes a currency from a country
func (uc *CountryUseCase) RemoveCurrency(ctx context.Context, countryID, currencyID uuid.UUID) error {
	return uc.countryRepo.RemoveCurrency(ctx, countryID, currencyID)
}

// GetCurrencies gets all currencies for a country
func (uc *CountryUseCase) GetCurrencies(ctx context.Context, countryID uuid.UUID) ([]entity.Currency, error) {
	return uc.countryRepo.GetCurrencies(ctx, countryID)
}
