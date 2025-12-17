package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CurrencyUseCase interface defines operations for currency management
type CurrencyUseCase interface {
	Create(ctx context.Context, currency *entity.Currency) error
	GetByID(ctx context.Context, id uuidv7.UUID, withCountries bool) (*entity.Currency, error)
	GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error)
	List(ctx context.Context, page, pageSize int, withCountries bool) ([]entity.Currency, int, error)
	Update(ctx context.Context, currency *entity.Currency) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	AddCountry(ctx context.Context, currencyID, countryID uuidv7.UUID, isPrimary bool) error
	RemoveCountry(ctx context.Context, currencyID, countryID uuidv7.UUID) error
	GetCountries(ctx context.Context, currencyID uuidv7.UUID) ([]entity.Country, error)
}

type currencyUseCase struct {
	currencyRepo repository.CurrencyRepository
	countryRepo  repository.CountryRepository
}

func NewCurrencyUseCase(
	currencyRepo repository.CurrencyRepository,
	countryRepo repository.CountryRepository,
) CurrencyUseCase {
	return &currencyUseCase{
		currencyRepo: currencyRepo,
		countryRepo:  countryRepo,
	}
}

// Create creates a new currency
func (uc *currencyUseCase) Create(ctx context.Context, currency *entity.Currency) error {
	// Validate entity
	if err := currency.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.currencyRepo.Create(ctx, currency)
}

// GetByID retrieves a currency by ID
func (uc *currencyUseCase) GetByID(ctx context.Context, id uuidv7.UUID, withCountries bool) (*entity.Currency, error) {
	return uc.currencyRepo.GetByID(ctx, id, withCountries)
}

// GetByCode retrieves a currency by code
func (uc *currencyUseCase) GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error) {
	return uc.currencyRepo.GetByCode(ctx, code, withCountries)
}

// List retrieves all currencies with pagination
func (uc *currencyUseCase) List(ctx context.Context, page, pageSize int, withCountries bool) ([]entity.Currency, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return uc.currencyRepo.List(ctx, offset, pageSize, withCountries)
}

// Update updates a currency
func (uc *currencyUseCase) Update(ctx context.Context, currency *entity.Currency) error {
	// Check if currency exists
	existing, err := uc.currencyRepo.GetByID(ctx, currency.ID, false)
	if err != nil {
		return err
	}

	if existing == nil {
		return entity.ErrNotFound
	}

	// Validate entity
	if err := currency.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return uc.currencyRepo.Update(ctx, currency)
}

// Delete deletes a currency
func (uc *currencyUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.currencyRepo.Delete(ctx, id)
}

// AddCountry adds a country to a currency
func (uc *currencyUseCase) AddCountry(ctx context.Context, currencyID, countryID uuidv7.UUID, isPrimary bool) error {
	// Verify currency exists
	_, err := uc.currencyRepo.GetByID(ctx, currencyID, false)
	if err != nil {
		return fmt.Errorf("currency not found: %w", err)
	}

	// Verify country exists
	_, err = uc.countryRepo.GetByID(ctx, countryID, false)
	if err != nil {
		return fmt.Errorf("country not found: %w", err)
	}

	return uc.currencyRepo.AddCountry(ctx, currencyID, countryID, isPrimary)
}

// RemoveCountry removes a country from a currency
func (uc *currencyUseCase) RemoveCountry(ctx context.Context, currencyID, countryID uuidv7.UUID) error {
	return uc.currencyRepo.RemoveCountry(ctx, currencyID, countryID)
}

// GetCountries gets all countries for a currency
func (uc *currencyUseCase) GetCountries(ctx context.Context, currencyID uuidv7.UUID) ([]entity.Country, error) {
	return uc.currencyRepo.GetCountries(ctx, currencyID)
}
