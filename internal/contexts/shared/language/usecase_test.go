package language

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

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Language, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Language), args.Error(1)
}

func (m *MockRepository) GetByCode(ctx context.Context, code string) (*Language, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Language), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context) ([]*Language, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Language), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, language *Language) error {
	args := m.Called(ctx, language)
	return args.Error(0)
}

func (m *MockRepository) Update(ctx context.Context, language *Language) error {
	args := m.Called(ctx, language)
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
		expected := &Language{
			ID:         id,
			Code:       "en",
			Code3:      "eng",
			Name:       "English",
			NativeName: "English",
			IsActive:   true,
		}

		mockRepo.On("GetByID", ctx, id).Return(expected, nil).Once()

		result, err := uc.GetByID(ctx, id)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		id := uuidv7.New()
		mockRepo.On("GetByID", ctx, id).Return(nil, ErrLanguageNotFound).Once()

		result, err := uc.GetByID(ctx, id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrLanguageNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_GetByCode(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := &Language{
			ID:         uuidv7.New(),
			Code:       "en",
			Code3:      "eng",
			Name:       "English",
			NativeName: "English",
			IsActive:   true,
		}

		mockRepo.On("GetByCode", ctx, "en").Return(expected, nil).Once()

		result, err := uc.GetByCode(ctx, "en")

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
		expected := []*Language{
			{ID: uuidv7.New(), Code: "en", Name: "English", IsActive: true},
			{ID: uuidv7.New(), Code: "uk", Name: "Ukrainian", IsActive: true},
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
		language := &Language{
			ID:         uuidv7.New(),
			Code:       "en",
			Code3:      "eng",
			Name:       "English",
			NativeName: "English",
			IsActive:   true,
		}

		mockRepo.On("Create", ctx, language).Return(nil).Once()

		err := uc.Create(ctx, language)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		language := &Language{
			ID:   uuidv7.New(),
			Code: "", // Invalid: empty code
		}

		err := uc.Create(ctx, language)

		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestUseCase_Update(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := NewUseCase(mockRepo, noop.NewNoOpCache())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		language := &Language{
			ID:         uuidv7.New(),
			Code:       "en",
			Code3:      "eng",
			Name:       "English Updated",
			NativeName: "English",
			IsActive:   true,
		}

		mockRepo.On("Update", ctx, language).Return(nil).Once()

		err := uc.Update(ctx, language)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		language := &Language{
			ID:   uuidv7.New(),
			Code: "", // Invalid: empty code
		}

		err := uc.Update(ctx, language)

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
		lang := &language.Language{
			ID:   id,
			Code: "en",
			Name: "English",
		}

		// Delete method calls GetByID first for cache invalidation
		mockRepo.On("GetByID", ctx, id).Return(lang, nil).Once()
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
