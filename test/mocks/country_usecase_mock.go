package mocks

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCountryUseCase is a mock implementation of usecase.CountryUseCase
type MockCountryUseCase struct {
	mock.Mock
}

func (m *MockCountryUseCase) Create(ctx context.Context, country *entity.Country) error {
	args := m.Called(ctx, country)
	return args.Error(0)
}

func (m *MockCountryUseCase) GetByID(ctx context.Context, id uuid.UUID, withCurrencies bool) (*entity.Country, error) {
	args := m.Called(ctx, id, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *MockCountryUseCase) GetByCode(ctx context.Context, code string, withCurrencies bool) (*entity.Country, error) {
	args := m.Called(ctx, code, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Country), args.Error(1)
}

func (m *MockCountryUseCase) List(ctx context.Context, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
	args := m.Called(ctx, page, pageSize, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Country), args.Int(1), args.Error(2)
}

func (m *MockCountryUseCase) ListByRegion(ctx context.Context, region string, page, pageSize int, withCurrencies bool) ([]entity.Country, int, error) {
	args := m.Called(ctx, region, page, pageSize, withCurrencies)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Country), args.Int(1), args.Error(2)
}

func (m *MockCountryUseCase) Update(ctx context.Context, country *entity.Country) error {
	args := m.Called(ctx, country)
	return args.Error(0)
}

func (m *MockCountryUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCountryUseCase) GetCurrencies(ctx context.Context, countryID uuid.UUID) ([]entity.Currency, error) {
	args := m.Called(ctx, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Currency), args.Error(1)
}

func (m *MockCountryUseCase) AddCurrency(ctx context.Context, countryID, currencyID uuid.UUID, isPrimary bool) error {
	args := m.Called(ctx, countryID, currencyID, isPrimary)
	return args.Error(0)
}

func (m *MockCountryUseCase) RemoveCurrency(ctx context.Context, countryID, currencyID uuid.UUID) error {
	args := m.Called(ctx, countryID, currencyID)
	return args.Error(0)
}
