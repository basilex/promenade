package mocks

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/mock"
)

// MockCurrencyUseCase is a mock implementation of usecase.CurrencyUseCase
type MockCurrencyUseCase struct {
	mock.Mock
}

func (m *MockCurrencyUseCase) Create(ctx context.Context, currency *entity.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *MockCurrencyUseCase) GetByID(ctx context.Context, id uuidv7.UUID, withCountries bool) (*entity.Currency, error) {
	args := m.Called(ctx, id, withCountries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Currency), args.Error(1)
}

func (m *MockCurrencyUseCase) GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error) {
	args := m.Called(ctx, code, withCountries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Currency), args.Error(1)
}

func (m *MockCurrencyUseCase) List(ctx context.Context, page, pageSize int, withCountries bool) ([]entity.Currency, int, error) {
	args := m.Called(ctx, page, pageSize, withCountries)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Currency), args.Int(1), args.Error(2)
}

func (m *MockCurrencyUseCase) Update(ctx context.Context, currency *entity.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *MockCurrencyUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCurrencyUseCase) GetCountries(ctx context.Context, currencyID uuidv7.UUID) ([]entity.Country, error) {
	args := m.Called(ctx, currencyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Country), args.Error(1)
}

func (m *MockCurrencyUseCase) AddCountry(ctx context.Context, currencyID, countryID uuidv7.UUID, isPrimary bool) error {
	args := m.Called(ctx, currencyID, countryID, isPrimary)
	return args.Error(0)
}

func (m *MockCurrencyUseCase) RemoveCountry(ctx context.Context, currencyID, countryID uuidv7.UUID) error {
	args := m.Called(ctx, currencyID, countryID)
	return args.Error(0)
}
