package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock CountryRepository
type mockCountryRepository struct {
	mock.Mock
}

func (m *mockCountryRepository) Create(ctx context.Context, country *entity.Country) error {
	args := m.Called(ctx, country)
	return args.Error(0)
}

func (m *mockCountryRepository) GetByID(ctx context.Context, id uuidv7.UUID, withCurrencies bool) (*entity.Country, error) {
	args := m.Called(ctx, id, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *mockCountryRepository) GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error) {
	args := m.Called(ctx, code, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *mockCountryRepository) List(ctx context.Context, offset, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
	args := m.Called(ctx, offset, pageSize, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Country), args.Int(1), args.Error(2)
}

func (m *mockCountryRepository) ListByRegion(ctx context.Context, region string, offset, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
	args := m.Called(ctx, region, offset, pageSize, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Country), args.Int(1), args.Error(2)
}

func (m *mockCountryRepository) Update(ctx context.Context, country *entity.Country) error {
	args := m.Called(ctx, country)
	return args.Error(0)
}

func (m *mockCountryRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCountryRepository) AddCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID, isPrimary bool) error {
	args := m.Called(ctx, countryID, currencyID, isPrimary)
	return args.Error(0)
}

func (m *mockCountryRepository) RemoveCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID) error {
	args := m.Called(ctx, countryID, currencyID)
	return args.Error(0)
}

func (m *mockCountryRepository) GetCurrencies(ctx context.Context, countryID uuidv7.UUID) ([]entity.Currency, error) {
	args := m.Called(ctx, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Currency), args.Error(1)
}

// Mock CurrencyRepository
type mockCurrencyRepository struct {
	mock.Mock
}

func (m *mockCurrencyRepository) Create(ctx context.Context, currency *entity.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *mockCurrencyRepository) GetByID(ctx context.Context, id uuidv7.UUID, withCountries bool) (*entity.Currency, error) {
	args := m.Called(ctx, id, withCountries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Currency), args.Error(1)
}

func (m *mockCurrencyRepository) GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error) {
	args := m.Called(ctx, code, withCountries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Currency), args.Error(1)
}

func (m *mockCurrencyRepository) List(ctx context.Context, offset, pageSize int, withCountries bool) ([]entity.Currency, int, error) {
	args := m.Called(ctx, offset, pageSize, withCountries)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Currency), args.Int(1), args.Error(2)
}

func (m *mockCurrencyRepository) Update(ctx context.Context, currency *entity.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *mockCurrencyRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockCurrencyRepository) AddCountry(ctx context.Context, currencyID, countryID uuidv7.UUID, isPrimary bool) error {
	args := m.Called(ctx, currencyID, countryID, isPrimary)
	return args.Error(0)
}

func (m *mockCurrencyRepository) RemoveCountry(ctx context.Context, currencyID, countryID uuidv7.UUID) error {
	args := m.Called(ctx, currencyID, countryID)
	return args.Error(0)
}

func (m *mockCurrencyRepository) GetCountries(ctx context.Context, currencyID uuidv7.UUID) ([]entity.Country, error) {
	args := m.Called(ctx, currencyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Country), args.Error(1)
}

// Test helpers
func setupCountryUseCase(t *testing.T) (*countryUseCase, *mockCountryRepository, *mockCurrencyRepository) {
	countryRepo := new(mockCountryRepository)
	currencyRepo := new(mockCurrencyRepository)
	uc := NewCountryUseCase(countryRepo, currencyRepo).(*countryUseCase)
	return uc, countryRepo, currencyRepo
}

func createTestCountry(name, code string) *entity.Country {
	return &entity.Country{
		ID:     uuidv7.New(),
		Name:   name,
		Code:   code,
		ISO2:   code,
		ISO3:   code + "X",
		Region: "north_america",
	}
}

func createTestCurrency(name, code string) *entity.Currency {
	return &entity.Currency{
		ID:     uuidv7.New(),
		Name:   name,
		Code:   code,
		Symbol: "$",
	}
}

// Tests for Create
func TestCountryUseCase_Create_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	countryRepo.On("Create", ctx, country).Return(nil)

	err := uc.Create(ctx, country)

	assert.NoError(t, err)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_Create_ValidationError(t *testing.T) {
	uc, _, _ := setupCountryUseCase(t)
	ctx := context.Background()

	// Country with invalid data
	country := &entity.Country{
		ID:     uuidv7.New(),
		Name:   "", // Empty name should fail validation
		Code:   "US",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	err := uc.Create(ctx, country)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestCountryUseCase_Create_RepositoryError(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	countryRepo.On("Create", ctx, country).Return(errors.New("database error"))

	err := uc.Create(ctx, country)

	assert.Error(t, err)
	countryRepo.AssertExpectations(t)
}

// Tests for GetByID
func TestCountryUseCase_GetByID_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	expectedCountry := createTestCountry("United States", "US")
	countryRepo.On("GetByID", ctx, expectedCountry.ID, false).Return(expectedCountry, nil)

	country, err := uc.GetByID(ctx, expectedCountry.ID, false)

	assert.NoError(t, err)
	assert.Equal(t, expectedCountry.ID, country.ID)
	assert.Equal(t, "United States", country.Name)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_GetByID_WithCurrencies(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	expectedCountry := createTestCountry("United States", "US")
	countryRepo.On("GetByID", ctx, expectedCountry.ID, true).Return(expectedCountry, nil)

	country, err := uc.GetByID(ctx, expectedCountry.ID, true)

	assert.NoError(t, err)
	assert.NotNil(t, country)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_GetByID_NotFound(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	countryRepo.On("GetByID", ctx, countryID, false).Return(nil, entity.ErrNotFound)

	country, err := uc.GetByID(ctx, countryID, false)

	assert.Error(t, err)
	assert.Nil(t, country)
	countryRepo.AssertExpectations(t)
}

// Tests for GetByCode
func TestCountryUseCase_GetByCode_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	expectedCountry := createTestCountry("United States", "US")
	countryRepo.On("GetByCode", ctx, "US", false).Return(expectedCountry, nil)

	country, err := uc.GetByCode(ctx, "US", false)

	assert.NoError(t, err)
	assert.Equal(t, "US", country.Code)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_GetByCode_NotFound(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryRepo.On("GetByCode", ctx, "XX", false).Return(nil, entity.ErrNotFound)

	country, err := uc.GetByCode(ctx, "XX", false)

	assert.Error(t, err)
	assert.Nil(t, country)
	countryRepo.AssertExpectations(t)
}

// Tests for List
func TestCountryUseCase_List_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country1 := createTestCountry("United States", "US")
	country2 := createTestCountry("Canada", "CA")
	expectedCountries := []entity.Country{*country1, *country2}

	countryRepo.On("List", ctx, 0, 20, false).Return(expectedCountries, 2, nil)

	countries, total, err := uc.List(ctx, 1, 20, false)

	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Equal(t, 2, total)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_List_DefaultPagination(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	expectedCountries := []entity.Country{}
	countryRepo.On("List", ctx, 0, 20, false).Return(expectedCountries, 0, nil)

	// Invalid page and pageSize should be normalized
	countries, total, err := uc.List(ctx, 0, 0, false)

	assert.NoError(t, err)
	assert.Empty(t, countries)
	assert.Equal(t, 0, total)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_List_MaxPageSize(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	expectedCountries := []entity.Country{}
	countryRepo.On("List", ctx, 0, 20, false).Return(expectedCountries, 0, nil)

	// PageSize > 100 should be normalized to 20
	countries, total, err := uc.List(ctx, 1, 200, false)

	assert.NoError(t, err)
	assert.Empty(t, countries)
	assert.Equal(t, 0, total)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_List_WithCurrencies(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	expectedCountries := []entity.Country{*country}

	countryRepo.On("List", ctx, 20, 10, true).Return(expectedCountries, 1, nil)

	countries, total, err := uc.List(ctx, 3, 10, true)

	assert.NoError(t, err)
	assert.Len(t, countries, 1)
	assert.Equal(t, 1, total)
	countryRepo.AssertExpectations(t)
}

// Tests for ListByRegion
func TestCountryUseCase_ListByRegion_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country1 := createTestCountry("United States", "US")
	country2 := createTestCountry("Canada", "CA")
	expectedCountries := []entity.Country{*country1, *country2}

	countryRepo.On("ListByRegion", ctx, "north_america", 0, 20, false).
		Return(expectedCountries, 2, nil)

	countries, total, err := uc.ListByRegion(ctx, "north_america", 1, 20, false)

	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Equal(t, 2, total)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_ListByRegion_Empty(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryRepo.On("ListByRegion", ctx, "asia", 0, 20, false).
		Return([]entity.Country{}, 0, nil)

	countries, total, err := uc.ListByRegion(ctx, "asia", 1, 20, false)

	assert.NoError(t, err)
	assert.Empty(t, countries)
	assert.Equal(t, 0, total)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_ListByRegion_DefaultPagination(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryRepo.On("ListByRegion", ctx, "western_europe", 0, 20, false).
		Return([]entity.Country{}, 0, nil)

	// Invalid pagination should be normalized
	countries, total, err := uc.ListByRegion(ctx, "western_europe", -1, 500, false)

	assert.NoError(t, err)
	assert.Empty(t, countries)
	assert.Equal(t, 0, total)
	countryRepo.AssertExpectations(t)
}

// Tests for Update
func TestCountryUseCase_Update_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	country.Name = "United States of America"

	countryRepo.On("GetByID", ctx, country.ID, false).Return(country, nil)
	countryRepo.On("Update", ctx, country).Return(nil)

	err := uc.Update(ctx, country)

	assert.NoError(t, err)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_Update_NotFound(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	countryRepo.On("GetByID", ctx, country.ID, false).Return(nil, entity.ErrNotFound)

	err := uc.Update(ctx, country)

	assert.Error(t, err)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_Update_ValidationError(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	country.Name = "" // Invalid name

	countryRepo.On("GetByID", ctx, country.ID, false).Return(country, nil)

	err := uc.Update(ctx, country)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_Update_RepositoryError(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	country := createTestCountry("United States", "US")
	countryRepo.On("GetByID", ctx, country.ID, false).Return(country, nil)
	countryRepo.On("Update", ctx, country).Return(errors.New("database error"))

	err := uc.Update(ctx, country)

	assert.Error(t, err)
	countryRepo.AssertExpectations(t)
}

// Tests for Delete
func TestCountryUseCase_Delete_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	countryRepo.On("Delete", ctx, countryID).Return(nil)

	err := uc.Delete(ctx, countryID)

	assert.NoError(t, err)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_Delete_RepositoryError(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	countryRepo.On("Delete", ctx, countryID).Return(errors.New("database error"))

	err := uc.Delete(ctx, countryID)

	assert.Error(t, err)
	countryRepo.AssertExpectations(t)
}

// Tests for AddCurrency
func TestCountryUseCase_AddCurrency_Success(t *testing.T) {
	uc, countryRepo, currencyRepo := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currencyID := uuidv7.New()
	country := createTestCountry("United States", "US")
	country.ID = countryID
	currency := createTestCurrency("US Dollar", "USD")
	currency.ID = currencyID

	countryRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
	currencyRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
	countryRepo.On("AddCurrency", ctx, countryID, currencyID, true).Return(nil)

	err := uc.AddCurrency(ctx, countryID, currencyID, true)

	assert.NoError(t, err)
	countryRepo.AssertExpectations(t)
	currencyRepo.AssertExpectations(t)
}

func TestCountryUseCase_AddCurrency_CountryNotFound(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currencyID := uuidv7.New()

	countryRepo.On("GetByID", ctx, countryID, false).Return(nil, entity.ErrNotFound)

	err := uc.AddCurrency(ctx, countryID, currencyID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "country not found")
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_AddCurrency_CurrencyNotFound(t *testing.T) {
	uc, countryRepo, currencyRepo := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currencyID := uuidv7.New()
	country := createTestCountry("United States", "US")
	country.ID = countryID

	countryRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
	currencyRepo.On("GetByID", ctx, currencyID, false).Return(nil, entity.ErrNotFound)

	err := uc.AddCurrency(ctx, countryID, currencyID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency not found")
	countryRepo.AssertExpectations(t)
	currencyRepo.AssertExpectations(t)
}

func TestCountryUseCase_AddCurrency_RepositoryError(t *testing.T) {
	uc, countryRepo, currencyRepo := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currencyID := uuidv7.New()
	country := createTestCountry("United States", "US")
	country.ID = countryID
	currency := createTestCurrency("US Dollar", "USD")
	currency.ID = currencyID

	countryRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
	currencyRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
	countryRepo.On("AddCurrency", ctx, countryID, currencyID, false).Return(errors.New("database error"))

	err := uc.AddCurrency(ctx, countryID, currencyID, false)

	assert.Error(t, err)
	countryRepo.AssertExpectations(t)
	currencyRepo.AssertExpectations(t)
}

// Tests for RemoveCurrency
func TestCountryUseCase_RemoveCurrency_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currencyID := uuidv7.New()

	countryRepo.On("RemoveCurrency", ctx, countryID, currencyID).Return(nil)

	err := uc.RemoveCurrency(ctx, countryID, currencyID)

	assert.NoError(t, err)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_RemoveCurrency_RepositoryError(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currencyID := uuidv7.New()

	countryRepo.On("RemoveCurrency", ctx, countryID, currencyID).Return(errors.New("database error"))

	err := uc.RemoveCurrency(ctx, countryID, currencyID)

	assert.Error(t, err)
	countryRepo.AssertExpectations(t)
}

// Tests for GetCurrencies
func TestCountryUseCase_GetCurrencies_Success(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	currency1 := createTestCurrency("US Dollar", "USD")
	currency2 := createTestCurrency("Euro", "EUR")
	expectedCurrencies := []entity.Currency{*currency1, *currency2}

	countryRepo.On("GetCurrencies", ctx, countryID).Return(expectedCurrencies, nil)

	currencies, err := uc.GetCurrencies(ctx, countryID)

	assert.NoError(t, err)
	assert.Len(t, currencies, 2)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_GetCurrencies_Empty(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	countryRepo.On("GetCurrencies", ctx, countryID).Return([]entity.Currency{}, nil)

	currencies, err := uc.GetCurrencies(ctx, countryID)

	assert.NoError(t, err)
	assert.Empty(t, currencies)
	countryRepo.AssertExpectations(t)
}

func TestCountryUseCase_GetCurrencies_RepositoryError(t *testing.T) {
	uc, countryRepo, _ := setupCountryUseCase(t)
	ctx := context.Background()

	countryID := uuidv7.New()
	countryRepo.On("GetCurrencies", ctx, countryID).Return(nil, errors.New("database error"))

	currencies, err := uc.GetCurrencies(ctx, countryID)

	assert.Error(t, err)
	assert.Nil(t, currencies)
	countryRepo.AssertExpectations(t)
}
