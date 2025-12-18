package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
)

func TestCurrencyUseCase_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currency := &entity.Currency{
			ID:     uuidv7.New(),
			Code:   "USD",
			Name:   "US Dollar",
			Symbol: "$",
		}

		mockRepo.On("Create", ctx, currency).Return(nil)

		err := uc.Create(ctx, currency)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currency := &entity.Currency{
			ID:   uuidv7.New(),
			Code: "", // Invalid: empty code
			Name: "US Dollar",
		}

		err := uc.Create(ctx, currency)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestCurrencyUseCase_GetByID(t *testing.T) {
	ctx := context.Background()
	currencyID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		expectedCurrency := &entity.Currency{
			ID:   currencyID,
			Code: "USD",
			Name: "US Dollar",
		}

		mockRepo.On("GetByID", ctx, currencyID, false).Return(expectedCurrency, nil)

		currency, err := uc.GetByID(ctx, currencyID, false)
		require.NoError(t, err)
		assert.Equal(t, expectedCurrency, currency)
		mockRepo.AssertExpectations(t)
	})

	t.Run("currency not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		mockRepo.On("GetByID", ctx, currencyID, false).Return(nil, entity.ErrNotFound)

		_, err := uc.GetByID(ctx, currencyID, false)
		assert.ErrorIs(t, err, entity.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_GetByCode(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval by code", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		expectedCurrency := &entity.Currency{
			ID:   uuidv7.New(),
			Code: "USD",
			Name: "US Dollar",
		}

		mockRepo.On("GetByCode", ctx, "USD", false).Return(expectedCurrency, nil)

		currency, err := uc.GetByCode(ctx, "USD", false)
		require.NoError(t, err)
		assert.Equal(t, expectedCurrency, currency)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list with pagination", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currencies := []entity.Currency{
			{ID: uuidv7.New(), Code: "USD", Name: "US Dollar"},
			{ID: uuidv7.New(), Code: "EUR", Name: "Euro"},
		}

		mockRepo.On("List", ctx, 0, 20, false).Return(currencies, 2, nil)

		result, total, err := uc.List(ctx, 1, 20, false)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("default pagination values", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		mockRepo.On("List", ctx, 0, 20, false).Return([]entity.Currency{}, 0, nil)

		// Invalid page and pageSize should be corrected
		_, _, err := uc.List(ctx, 0, 0, false)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("max page size limit", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		// Should cap at 20 (default)
		mockRepo.On("List", ctx, 0, 20, false).Return([]entity.Currency{}, 0, nil)

		_, _, err := uc.List(ctx, 1, 150, false)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currency := &entity.Currency{
			ID:     uuidv7.New(),
			Code:   "USD",
			Name:   "Updated Name",
			Symbol: "$",
		}

		mockRepo.On("GetByID", ctx, currency.ID, false).Return(currency, nil)
		mockRepo.On("Update", ctx, currency).Return(nil)

		err := uc.Update(ctx, currency)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("currency not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currency := &entity.Currency{
			ID:   uuidv7.New(),
			Code: "USD",
			Name: "Updated Name",
		}

		mockRepo.On("GetByID", ctx, currency.ID, false).Return(nil, entity.ErrNotFound)

		err := uc.Update(ctx, currency)
		assert.ErrorIs(t, err, entity.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error on update", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currency := &entity.Currency{
			ID:   uuidv7.New(),
			Code: "", // Invalid
			Name: "US Dollar",
		}

		mockRepo.On("GetByID", ctx, currency.ID, false).Return(currency, nil)

		err := uc.Update(ctx, currency)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	currencyID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		mockRepo.On("Delete", ctx, currencyID).Return(nil)

		err := uc.Delete(ctx, currencyID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deletion fails", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		mockRepo.On("Delete", ctx, currencyID).Return(errors.New("delete failed"))

		err := uc.Delete(ctx, currencyID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_AddCountry(t *testing.T) {
	ctx := context.Background()
	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	t.Run("successful add", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		currency := &entity.Currency{ID: currencyID}
		country := &entity.Country{ID: countryID}

		mockRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
		mockCountryRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
		mockRepo.On("AddCountry", ctx, currencyID, countryID, true).Return(nil)

		err := uc.AddCountry(ctx, currencyID, countryID, true)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCountryRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_RemoveCountry(t *testing.T) {
	ctx := context.Background()
	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	t.Run("successful removal", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		mockRepo.On("RemoveCountry", ctx, currencyID, countryID).Return(nil)

		err := uc.RemoveCountry(ctx, currencyID, countryID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurrencyUseCase_GetCountries(t *testing.T) {
	ctx := context.Background()
	currencyID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		countries := []entity.Country{
			{ID: uuidv7.New(), Name: "United States", ISO2: "US"},
			{ID: uuidv7.New(), Name: "Canada", ISO2: "CA"},
		}

		mockRepo.On("GetCountries", ctx, currencyID).Return(countries, nil)

		result, err := uc.GetCountries(ctx, currencyID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo := new(mocks.MockCurrencyRepository)
		mockCountryRepo := new(mocks.MockCountryRepository)
		uc := NewCurrencyUseCase(mockRepo, mockCountryRepo)

		mockRepo.On("GetCountries", ctx, currencyID).Return([]entity.Country{}, nil)

		result, err := uc.GetCountries(ctx, currencyID)
		require.NoError(t, err)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})
}
