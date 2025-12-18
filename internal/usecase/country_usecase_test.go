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

func TestCountryUseCase_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		country := &entity.Country{
			ID:     uuidv7.New(),
			Name:   "United States",
			ISO2:   "US",
			ISO3:   "USA",
			Code:   "USA",
			Region: "north_america",
		}

		mockRepo.On("Create", ctx, country).Return(nil)

		err := uc.Create(ctx, country)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		country := &entity.Country{
			ID:   uuidv7.New(),
			Name: "", // Invalid: empty name
			ISO2: "US",
		}

		err := uc.Create(ctx, country)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestCountryUseCase_GetByID(t *testing.T) {
	ctx := context.Background()
	countryID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		expectedCountry := &entity.Country{
			ID:   countryID,
			Name: "United States",
			ISO2: "US",
		}

		mockRepo.On("GetByID", ctx, countryID, false).Return(expectedCountry, nil)

		country, err := uc.GetByID(ctx, countryID, false)
		require.NoError(t, err)
		assert.Equal(t, expectedCountry, country)
		mockRepo.AssertExpectations(t)
	})

	t.Run("country not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		mockRepo.On("GetByID", ctx, countryID, false).Return(nil, entity.ErrNotFound)

		_, err := uc.GetByID(ctx, countryID, false)
		assert.ErrorIs(t, err, entity.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_GetByCode(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval by ISO2", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		expectedCountry := &entity.Country{
			ID:   uuidv7.New(),
			Name: "United States",
			ISO2: "US",
		}

		mockRepo.On("GetByCode", ctx, "US", false).Return(expectedCountry, nil)

		country, err := uc.GetByCode(ctx, "US", false)
		require.NoError(t, err)
		assert.Equal(t, expectedCountry, country)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list with pagination", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		countries := []entity.Country{
			{ID: uuidv7.New(), Name: "Country 1", ISO2: "C1"},
			{ID: uuidv7.New(), Name: "Country 2", ISO2: "C2"},
		}

		mockRepo.On("List", ctx, 0, 20, false).Return(countries, 2, nil)

		result, total, err := uc.List(ctx, 1, 20, false)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("default pagination values", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		mockRepo.On("List", ctx, 0, 20, false).Return([]entity.Country{}, 0, nil)

		// Invalid page and pageSize should be corrected
		_, _, err := uc.List(ctx, 0, 0, false)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		country := &entity.Country{
			ID:     uuidv7.New(),
			Name:   "Updated Name",
			ISO2:   "US",
			ISO3:   "USA",
			Code:   "USA",
			Region: "north_america",
		}

		mockRepo.On("GetByID", ctx, country.ID, false).Return(country, nil)
		mockRepo.On("Update", ctx, country).Return(nil)

		err := uc.Update(ctx, country)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("country not found", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		country := &entity.Country{
			ID:   uuidv7.New(),
			Name: "Updated Name",
			ISO2: "US",
		}

		mockRepo.On("GetByID", ctx, country.ID, false).Return(nil, entity.ErrNotFound)

		err := uc.Update(ctx, country)
		assert.ErrorIs(t, err, entity.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	countryID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		mockRepo.On("Delete", ctx, countryID).Return(nil)

		err := uc.Delete(ctx, countryID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deletion fails", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		mockRepo.On("Delete", ctx, countryID).Return(errors.New("delete failed"))

		err := uc.Delete(ctx, countryID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_AddCurrency(t *testing.T) {
	ctx := context.Background()
	countryID := uuidv7.New()
	currencyID := uuidv7.New()

	t.Run("successful add", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		country := &entity.Country{ID: countryID}
		currency := &entity.Currency{ID: currencyID}

		mockRepo.On("GetByID", ctx, countryID, false).Return(country, nil)
		mockCurrencyRepo.On("GetByID", ctx, currencyID, false).Return(currency, nil)
		mockRepo.On("AddCurrency", ctx, countryID, currencyID, true).Return(nil)

		err := uc.AddCurrency(ctx, countryID, currencyID, true)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCurrencyRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_RemoveCurrency(t *testing.T) {
	ctx := context.Background()
	countryID := uuidv7.New()
	currencyID := uuidv7.New()

	t.Run("successful removal", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		mockRepo.On("RemoveCurrency", ctx, countryID, currencyID).Return(nil)

		err := uc.RemoveCurrency(ctx, countryID, currencyID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_GetCurrencies(t *testing.T) {
	ctx := context.Background()
	countryID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		currencies := []entity.Currency{
			{ID: uuidv7.New(), Code: "USD", Name: "US Dollar"},
		}

		mockRepo.On("GetCurrencies", ctx, countryID).Return(currencies, nil)

		result, err := uc.GetCurrencies(ctx, countryID)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestCountryUseCase_ListByRegion(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list by region", func(t *testing.T) {
		mockRepo := new(mocks.MockCountryRepository)
		mockCurrencyRepo := new(mocks.MockCurrencyRepository)
		uc := NewCountryUseCase(mockRepo, mockCurrencyRepo)

		countries := []entity.Country{
			{ID: uuidv7.New(), Name: "USA", Region: "Americas"},
			{ID: uuidv7.New(), Name: "Canada", Region: "Americas"},
		}

		mockRepo.On("ListByRegion", ctx, "Americas", 0, 20, false).Return(countries, 2, nil)

		result, total, err := uc.ListByRegion(ctx, "Americas", 1, 20, false)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		mockRepo.AssertExpectations(t)
	})
}
