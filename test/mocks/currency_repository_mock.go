package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCurrencyRepository is a mock implementation of repository.CurrencyRepository
type MockCurrencyRepository struct {
	mock.Mock
}

func (m *MockCurrencyRepository) Create(ctx context.Context, currency *entity.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *MockCurrencyRepository) GetByID(ctx context.Context, id uuidv7.UUID, withCountries bool) (*entity.Currency, error) {
	args := m.Called(ctx, id, withCountries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Currency), args.Error(1)
}

func (m *MockCurrencyRepository) GetByCode(ctx context.Context, code string, withCountries bool) (*entity.Currency, error) {
	args := m.Called(ctx, code, withCountries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Currency), args.Error(1)
}

func (m *MockCurrencyRepository) List(ctx context.Context, offset, limit int, withCountries bool) ([]entity.Currency, int, error) {
	args := m.Called(ctx, offset, limit, withCountries)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Currency), args.Int(1), args.Error(2)
}

func (m *MockCurrencyRepository) Update(ctx context.Context, currency *entity.Currency) error {
	args := m.Called(ctx, currency)
	return args.Error(0)
}

func (m *MockCurrencyRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCurrencyRepository) AddCountry(ctx context.Context, currencyID, countryID uuidv7.UUID, isPrimary bool) error {
	args := m.Called(ctx, currencyID, countryID, isPrimary)
	return args.Error(0)
}

func (m *MockCurrencyRepository) RemoveCountry(ctx context.Context, currencyID, countryID uuidv7.UUID) error {
	args := m.Called(ctx, currencyID, countryID)
	return args.Error(0)
}

func (m *MockCurrencyRepository) GetCountries(ctx context.Context, currencyID uuidv7.UUID) ([]entity.Country, error) {
	args := m.Called(ctx, currencyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Country), args.Error(1)
}
