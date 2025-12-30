package timezone

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/cache/noop"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockRepository is a mock implementation of IRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Timezone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Timezone), args.Error(1)
}

func (m *MockRepository) GetByName(ctx context.Context, name string) (*Timezone, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Timezone), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context) ([]*Timezone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Timezone), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, timezone *Timezone) error {
	args := m.Called(ctx, timezone)
	return args.Error(0)
}

func (m *MockRepository) Update(ctx context.Context, timezone *Timezone) error {
	args := m.Called(ctx, timezone)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUseCase_GetByID(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		expected := &Timezone{
			ID:           id,
			Name:         "Europe/Kyiv",
			Abbreviation: "EET",
			UTCOffset:    7200, // +02:00
			IsActive:     true,
		}

		mockRepo.On("GetByID", ctx, id).Return(expected, nil).Once()

		result, err := uc.GetByID(ctx, id)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		id := uuidv7.New()
		mockRepo.On("GetByID", ctx, id).Return(nil, ErrTimezoneNotFound).Once()

		result, err := uc.GetByID(ctx, id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrTimezoneNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_GetByName(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := &Timezone{
			ID:           uuidv7.New(),
			Name:         "Europe/Kyiv",
			Abbreviation: "EET",
			UTCOffset:    7200, // +02:00
			IsActive:     true,
		}

		mockRepo.On("GetByName", ctx, "Europe/Kyiv").Return(expected, nil).Once()

		result, err := uc.GetByName(ctx, "Europe/Kyiv")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_List(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*Timezone{
			{ID: uuidv7.New(), Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200, IsActive: true},
			{ID: uuidv7.New(), Name: "America/New_York", Abbreviation: "EST", UTCOffset: -18000, IsActive: true},
		}

		mockRepo.On("List", ctx).Return(expected, nil).Once()

		result, err := uc.List(ctx)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_Create(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		timezone := &Timezone{
			ID:           uuidv7.New(),
			Name:         "Europe/Kyiv",
			Abbreviation: "EET",
			UTCOffset:    7200, // +02:00
			IsActive:     true,
		}

		mockRepo.On("Create", ctx, timezone).Return(nil).Once()

		err := uc.Create(ctx, timezone)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		timezone := &Timezone{
			ID:   uuidv7.New(),
			Name: "", // Invalid: empty name
		}

		err := uc.Create(ctx, timezone)

		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestUseCase_Update(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		timezone := &Timezone{
			ID:           uuidv7.New(),
			Name:         "Europe/Kyiv",
			Abbreviation: "EEST",
			UTCOffset:    10800, // +03:00 (summer time)
			IsActive:     true,
		}

		mockRepo.On("Update", ctx, timezone).Return(nil).Once()

		err := uc.Update(ctx, timezone)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		timezone := &Timezone{
			ID:   uuidv7.New(),
			Name: "", // Invalid: empty name
		}

		err := uc.Update(ctx, timezone)

		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "Update")
	})
}

func TestUseCase_Delete(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()

		mockRepo.On("Delete", ctx, id).Return(nil).Once()

		err := uc.Delete(ctx, id)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		id := uuidv7.New()
		expectedErr := errors.New("delete failed")

		mockRepo.On("Delete", ctx, id).Return(expectedErr).Once()

		err := uc.Delete(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
