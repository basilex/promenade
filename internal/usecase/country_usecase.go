package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICountryUseCase interface defines operations for country management
type ICountryUseCase interface {
	Create(ctx context.Context, country *entity.Country) error
	GetByID(ctx context.Context, id uuidv7.UUID, withCurrencies bool) (*entity.Country, error)
	GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error)
	List(ctx context.Context, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error)
	ListByRegion(ctx context.Context, region string, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error)
	Update(ctx context.Context, country *entity.Country) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	AddCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID, isPrimary bool) error
	RemoveCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID) error
	GetCurrencies(ctx context.Context, countryID uuidv7.UUID) ([]entity.Currency, error)
}

type countryUseCase struct {
	countryRepo  repository.ICountryRepository
	currencyRepo repository.ICurrencyRepository
}

func NewCountryUseCase(
	countryRepo repository.ICountryRepository,
	currencyRepo repository.ICurrencyRepository,
) ICountryUseCase {
	return &countryUseCase{
		countryRepo:  countryRepo,
		currencyRepo: currencyRepo,
	}
}

// Create creates a new country
func (uc *countryUseCase) Create(ctx context.Context, country *entity.Country) error {
	// Validate entity
	if err := country.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.countryRepo.Create(ctx, country)
}

// GetByID retrieves a country by ID
func (uc *countryUseCase) GetByID(ctx context.Context, id uuidv7.UUID, withCurrencies bool) (*entity.Country, error) {
	return uc.countryRepo.GetByID(ctx, id, withCurrencies)
}

// GetByCode retrieves a country by code
func (uc *countryUseCase) GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error) {
	return uc.countryRepo.GetByCode(ctx, code, withCurrencies)
}

// List retrieves all countries with pagination
func (uc *countryUseCase) List(ctx context.Context, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
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
func (uc *countryUseCase) ListByRegion(ctx context.Context, region string, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
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
func (uc *countryUseCase) Update(ctx context.Context, country *entity.Country) error {
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
func (uc *countryUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.countryRepo.Delete(ctx, id)
}

// AddCurrency adds a currency to a country
func (uc *countryUseCase) AddCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID, isPrimary bool) error {
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
func (uc *countryUseCase) RemoveCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID) error {
	return uc.countryRepo.RemoveCurrency(ctx, countryID, currencyID)
}

// GetCurrencies gets all currencies for a country
func (uc *countryUseCase) GetCurrencies(ctx context.Context, countryID uuidv7.UUID) ([]entity.Currency, error) {
	return uc.countryRepo.GetCurrencies(ctx, countryID)
}
