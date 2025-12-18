package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCountryRepository is a mock implementation of repository.CountryRepository
type MockCountryRepository struct {
	mock.Mock
}

func (m *MockCountryRepository) Create(ctx context.Context, country *entity.Country) error {
	args := m.Called(ctx, country)
	return args.Error(0)
}

func (m *MockCountryRepository) GetByID(ctx context.Context, id uuidv7.UUID, withCurrencies bool) (*entity.Country, error) {
	args := m.Called(ctx, id, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *MockCountryRepository) GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error) {
	args := m.Called(ctx, code, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *MockCountryRepository) List(ctx context.Context, offset, limit int, withCurrencies bool) ([]entity.Country, int, error) {
	args := m.Called(ctx, offset, limit, withCurrencies)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Country), args.Int(1), args.Error(2)
}

func (m *MockCountryRepository) ListByRegion(ctx context.Context, region string, offset, limit int, withCurrencies bool) ([]entity.Country, int, error) {
	args := m.Called(ctx, region, offset, limit, withCurrencies)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Country), args.Int(1), args.Error(2)
}

func (m *MockCountryRepository) Update(ctx context.Context, country *entity.Country) error {
	args := m.Called(ctx, country)
	return args.Error(0)
}

func (m *MockCountryRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCountryRepository) AddCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID, isPrimary bool) error {
	args := m.Called(ctx, countryID, currencyID, isPrimary)
	return args.Error(0)
}

func (m *MockCountryRepository) RemoveCurrency(ctx context.Context, countryID, currencyID uuidv7.UUID) error {
	args := m.Called(ctx, countryID, currencyID)
	return args.Error(0)
}

func (m *MockCountryRepository) GetCurrencies(ctx context.Context, countryID uuidv7.UUID) ([]entity.Currency, error) {
	args := m.Called(ctx, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Currency), args.Error(1)
}
