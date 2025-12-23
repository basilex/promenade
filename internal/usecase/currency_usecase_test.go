package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Test helpers for currency
func setupCurrencyUseCase(t *testing.T) (*currencyUseCase, *mockCurrencyRepository, *mockCountryRepository) {
	currencyRepo := new(mockCurrencyRepository)
	countryRepo := new(mockCountryRepository)
	uc := NewCurrencyUseCase(currencyRepo, countryRepo).(*currencyUseCase)
	return uc, currencyRepo, countryRepo
}

// Tests for Create
func TestCurrencyUseCase_Create_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("Create", ctx, currency).Return(nil)

	err := uc.Create(ctx, currency)

	assert.NoError(t, err)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_Create_ValidationError(t *testing.T) {
	uc, _, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	// Currency with invalid data
	currency := &entity.Currency{
		ID:     uuidv7.New(),
		Name:   "", // Empty name should fail validation
		Code:   "USD",
		Symbol: "$",
	}

	err := uc.Create(ctx, currency)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestCurrencyUseCase_Create_RepositoryError(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("Create", ctx, currency).Return(errors.New("database error"))

	err := uc.Create(ctx, currency)

	assert.Error(t, err)
	currencyRepo.AssertExpectations(t)
}

// Tests for GetByID
func TestCurrencyUseCase_GetByID_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	expectedCurrency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("GetByID", ctx, expectedCurrency.ID, false).Return(expectedCurrency, nil)

	currency, err := uc.GetByID(ctx, expectedCurrency.ID, false)

	assert.NoError(t, err)
	assert.Equal(t, expectedCurrency.ID, currency.ID)
	assert.Equal(t, "US Dollar", currency.Name)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_GetByID_WithCountries(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	expectedCurrency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("GetByID", ctx, expectedCurrency.ID, true).Return(expectedCurrency, nil)

	currency, err := uc.GetByID(ctx, expectedCurrency.ID, true)

	assert.NoError(t, err)
	assert.NotNil(t, currency)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_GetByID_NotFound(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	currencyRepo.On("GetByID", ctx, currencyID, false).Return(nil, entity.ErrNotFound)

	currency, err := uc.GetByID(ctx, currencyID, false)

	assert.Error(t, err)
	assert.Nil(t, currency)
	currencyRepo.AssertExpectations(t)
}

// Tests for GetByCode
func TestCurrencyUseCase_GetByCode_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	expectedCurrency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("GetByCode", ctx, "USD", false).Return(expectedCurrency, nil)

	currency, err := uc.GetByCode(ctx, "USD", false)

	assert.NoError(t, err)
	assert.Equal(t, "USD", currency.Code)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_GetByCode_NotFound(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyRepo.On("GetByCode", ctx, "XXX", false).Return(nil, entity.ErrNotFound)

	currency, err := uc.GetByCode(ctx, "XXX", false)

	assert.Error(t, err)
	assert.Nil(t, currency)
	currencyRepo.AssertExpectations(t)
}

// Tests for List
func TestCurrencyUseCase_List_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency1 := createTestCurrency("US Dollar", "USD")
	currency2 := createTestCurrency("Euro", "EUR")
	expectedCurrencies := []entity.Currency{*currency1, *currency2}

	currencyRepo.On("List", ctx, 0, 20, false).Return(expectedCurrencies, 2, nil)

	currencies, total, err := uc.List(ctx, 1, 20, false)

	assert.NoError(t, err)
	assert.Len(t, currencies, 2)
	assert.Equal(t, 2, total)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_List_DefaultPagination(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	expectedCurrencies := []entity.Currency{}
	currencyRepo.On("List", ctx, 0, 20, false).Return(expectedCurrencies, 0, nil)

	// Invalid page and pageSize should be normalized
	currencies, total, err := uc.List(ctx, 0, 0, false)

	assert.NoError(t, err)
	assert.Empty(t, currencies)
	assert.Equal(t, 0, total)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_List_MaxPageSize(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	expectedCurrencies := []entity.Currency{}
	currencyRepo.On("List", ctx, 0, 20, false).Return(expectedCurrencies, 0, nil)

	// PageSize > 100 should be normalized to 20
	currencies, total, err := uc.List(ctx, 1, 200, false)

	assert.NoError(t, err)
	assert.Empty(t, currencies)
	assert.Equal(t, 0, total)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_List_WithCountries(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	expectedCurrencies := []entity.Currency{*currency}

	currencyRepo.On("List", ctx, 20, 10, true).Return(expectedCurrencies, 1, nil)

	currencies, total, err := uc.List(ctx, 3, 10, true)

	assert.NoError(t, err)
	assert.Len(t, currencies, 1)
	assert.Equal(t, 1, total)
	currencyRepo.AssertExpectations(t)
}

// Tests for Update
func TestCurrencyUseCase_Update_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	currency.Name = "United States Dollar"

	currencyRepo.On("GetByID", ctx, currency.ID, false).Return(currency, nil)
	currencyRepo.On("Update", ctx, currency).Return(nil)

	err := uc.Update(ctx, currency)

	assert.NoError(t, err)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_Update_NotFound(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("GetByID", ctx, currency.ID, false).Return(nil, entity.ErrNotFound)

	err := uc.Update(ctx, currency)

	assert.Error(t, err)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_Update_ValidationError(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	currency.Name = "" // Invalid name

	currencyRepo.On("GetByID", ctx, currency.ID, false).Return(currency, nil)

	err := uc.Update(ctx, currency)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_Update_RepositoryError(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currency := createTestCurrency("US Dollar", "USD")
	currencyRepo.On("GetByID", ctx, currency.ID, false).Return(currency, nil)
	currencyRepo.On("Update", ctx, currency).Return(errors.New("database error"))

	err := uc.Update(ctx, currency)

	assert.Error(t, err)
	currencyRepo.AssertExpectations(t)
}

// Tests for Delete
func TestCurrencyUseCase_Delete_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	currencyRepo.On("Delete", ctx, currencyID).Return(nil)

	err := uc.Delete(ctx, currencyID)

	assert.NoError(t, err)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_Delete_RepositoryError(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	currencyRepo.On("Delete", ctx, currencyID).Return(errors.New("database error"))

	err := uc.Delete(ctx, currencyID)

	assert.Error(t, err)
	currencyRepo.AssertExpectations(t)
}

// Tests for AddCountry
func TestCurrencyUseCase_AddCountry_Success(t *testing.T) {
	uc, currencyRepo, countryRepo := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()
	currency := createTestCurrency("US Dollar", "USD")
	currency.ID = currencyID
	country := createTestCountry("United States", "US")
	country.ID = countryID

	currencyRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
	countryRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
	currencyRepo.On("AddCountry", ctx, currencyID, countryID, true).Return(nil)

	err := uc.AddCountry(ctx, currencyID, countryID, true)

	assert.NoError(t, err)
	currencyRepo.AssertExpectations(t)
	countryRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_AddCountry_CurrencyNotFound(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	currencyRepo.On("GetByID", ctx, currencyID, false).Return(nil, entity.ErrNotFound)

	err := uc.AddCountry(ctx, currencyID, countryID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency not found")
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_AddCountry_CountryNotFound(t *testing.T) {
	uc, currencyRepo, countryRepo := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()
	currency := createTestCurrency("US Dollar", "USD")
	currency.ID = currencyID

	currencyRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
	countryRepo.On("GetByID", ctx, countryID, false).Return(nil, entity.ErrNotFound)

	err := uc.AddCountry(ctx, currencyID, countryID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "country not found")
	currencyRepo.AssertExpectations(t)
	countryRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_AddCountry_RepositoryError(t *testing.T) {
	uc, currencyRepo, countryRepo := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()
	currency := createTestCurrency("US Dollar", "USD")
	currency.ID = currencyID
	country := createTestCountry("United States", "US")
	country.ID = countryID

	currencyRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
	countryRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
	currencyRepo.On("AddCountry", ctx, currencyID, countryID, false).Return(errors.New("database error"))

	err := uc.AddCountry(ctx, currencyID, countryID, false)

	assert.Error(t, err)
	currencyRepo.AssertExpectations(t)
	countryRepo.AssertExpectations(t)
}

// Tests for RemoveCountry
func TestCurrencyUseCase_RemoveCountry_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	currencyRepo.On("RemoveCountry", ctx, currencyID, countryID).Return(nil)

	err := uc.RemoveCountry(ctx, currencyID, countryID)

	assert.NoError(t, err)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_RemoveCountry_RepositoryError(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	currencyRepo.On("RemoveCountry", ctx, currencyID, countryID).Return(errors.New("database error"))

	err := uc.RemoveCountry(ctx, currencyID, countryID)

	assert.Error(t, err)
	currencyRepo.AssertExpectations(t)
}

// Tests for GetCountries
func TestCurrencyUseCase_GetCountries_Success(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	country1 := createTestCountry("United States", "US")
	country2 := createTestCountry("Panama", "PA")
	expectedCountries := []entity.Country{*country1, *country2}

	currencyRepo.On("GetCountries", ctx, currencyID).Return(expectedCountries, nil)

	countries, err := uc.GetCountries(ctx, currencyID)

	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_GetCountries_Empty(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	currencyRepo.On("GetCountries", ctx, currencyID).Return([]entity.Country{}, nil)

	countries, err := uc.GetCountries(ctx, currencyID)

	assert.NoError(t, err)
	assert.Empty(t, countries)
	currencyRepo.AssertExpectations(t)
}

func TestCurrencyUseCase_GetCountries_RepositoryError(t *testing.T) {
	uc, currencyRepo, _ := setupCurrencyUseCase(t)
	ctx := context.Background()

	currencyID := uuidv7.New()
	currencyRepo.On("GetCountries", ctx, currencyID).Return(nil, errors.New("database error"))

	countries, err := uc.GetCountries(ctx, currencyID)

	assert.Error(t, err)
	assert.Nil(t, countries)
	currencyRepo.AssertExpectations(t)
}
