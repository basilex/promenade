package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockAuditEventRepository for testing
type MockAuditEventRepository struct {
	mock.Mock
}

func (m *MockAuditEventRepository) Create(ctx context.Context, event *entity.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditEventRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.AuditEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AuditEvent), args.Error(1)
}

func (m *MockAuditEventRepository) List(ctx context.Context, filters repository.AuditEventFilters, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	args := m.Called(ctx, filters, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.AuditEvent), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockAuditEventRepository) GetByEntityID(ctx context.Context, entityType string, entityID uuidv7.UUID) ([]*entity.AuditEvent, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditEvent), args.Error(1)
}

func (m *MockAuditEventRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.AuditEvent), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockAuditEventRepository) GetByAction(ctx context.Context, action string, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	args := m.Called(ctx, action, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.AuditEvent), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockAuditEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditEventRepository) VerifySignature(ctx context.Context, event *entity.AuditEvent, secret string) (bool, error) {
	args := m.Called(ctx, event, secret)
	return args.Bool(0), args.Error(1)
}

func TestAuditEventUseCase_CreateAuditEvent(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditEventRepository)
	secret := "test-secret"
	uc := NewAuditEventUseCase(mockRepo, secret)

	t.Run("successful creation", func(t *testing.T) {
		userID := uuidv7.New()
		entityID := uuidv7.New()

		data := &entity.AuditEventCreate{
			UserID:     userID,
			Action:     "user.created",
			EntityType: "user",
			EntityID:   entityID,
			IPAddress:  "127.0.0.1",
			UserAgent:  "TestAgent",
			RequestID:  "req-123",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AuditEvent")).Return(nil)

		event, err := uc.CreateAuditEvent(ctx, data)

		require.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, "user.created", event.Action)
		assert.NotEmpty(t, event.Signature)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		data := &entity.AuditEventCreate{
			Action:     "user.created",
			EntityType: "user",
			EntityID:   uuidv7.New(),
		}

		event, err := uc.CreateAuditEvent(ctx, data)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Equal(t, entity.ErrUserIDRequired, err)
	})
}

func TestAuditEventUseCase_GetAuditEvent(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditEventRepository)
	uc := NewAuditEventUseCase(mockRepo, "secret")

	t.Run("get existing event", func(t *testing.T) {
		eventID := uuidv7.New()
		expectedEvent := &entity.AuditEvent{
			ID:         eventID,
			UserID:     uuidv7.New(),
			Action:     "user.created",
			EntityType: "user",
			EntityID:   uuidv7.New(),
			Signature:  "sig",
			CreatedAt:  time.Now(),
		}

		mockRepo.On("GetByID", ctx, eventID).Return(expectedEvent, nil)

		event, err := uc.GetAuditEvent(ctx, eventID)

		require.NoError(t, err)
		assert.Equal(t, expectedEvent, event)
		mockRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		eventID := uuidv7.New()
		mockRepo.On("GetByID", ctx, eventID).Return(nil, errors.New("not found"))

		event, err := uc.GetAuditEvent(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, event)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_ListAuditEvents(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditEventRepository)
	uc := NewAuditEventUseCase(mockRepo, "secret")

	t.Run("list with filters", func(t *testing.T) {
		userID := uuidv7.New()
		filters := repository.AuditEventFilters{UserID: &userID}
		params := pagination.Params{Page: 1, Limit: 10}

		expectedEvents := []*entity.AuditEvent{
			{ID: uuidv7.New(), UserID: userID, Action: "user.login"},
			{ID: uuidv7.New(), UserID: userID, Action: "user.logout"},
		}
		expectedMeta := &pagination.Metadata{Total: 2, CurrentPage: 1}

		mockRepo.On("List", ctx, filters, params).Return(expectedEvents, expectedMeta, nil)

		events, meta, err := uc.ListAuditEvents(ctx, filters, params)

		require.NoError(t, err)
		assert.Equal(t, expectedEvents, events)
		assert.Equal(t, expectedMeta, meta)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		filters := repository.AuditEventFilters{}
		params := pagination.Params{Page: 1, Limit: 10}

		mockRepo.On("List", ctx, filters, params).Return([]*entity.AuditEvent{}, &pagination.Metadata{Total: 0}, nil)

		events, meta, err := uc.ListAuditEvents(ctx, filters, params)

		require.NoError(t, err)
		assert.Empty(t, events)
		assert.Equal(t, 0, meta.Total)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_GetAuditEventsByEntity(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditEventRepository)
	uc := NewAuditEventUseCase(mockRepo, "secret")

	t.Run("get events for entity", func(t *testing.T) {
		entityID := uuidv7.New()
		expectedEvents := []*entity.AuditEvent{
			{ID: uuidv7.New(), Action: "post.created", EntityType: "post", EntityID: entityID},
			{ID: uuidv7.New(), Action: "post.updated", EntityType: "post", EntityID: entityID},
		}

		mockRepo.On("GetByEntityID", ctx, "post", entityID).Return(expectedEvents, nil)

		events, err := uc.GetAuditEventsByEntity(ctx, "post", entityID)

		require.NoError(t, err)
		assert.Equal(t, expectedEvents, events)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_PurgeOldAuditEvents(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditEventRepository)
	uc := NewAuditEventUseCase(mockRepo, "secret")

	t.Run("successful purge", func(t *testing.T) {
		retentionDays := 90

		mockRepo.On("DeleteOlderThan", ctx, mock.AnythingOfType("time.Time")).Return(int64(42), nil)

		deleted, err := uc.PurgeOldAuditEvents(ctx, retentionDays)

		require.NoError(t, err)
		assert.Equal(t, int64(42), deleted)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no records to delete", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")
		mockRepo.On("DeleteOlderThan", ctx, mock.AnythingOfType("time.Time")).Return(int64(0), nil)

		deleted, err := uc.PurgeOldAuditEvents(ctx, 90)

		require.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		expectedErr := errors.New("database error")
		mockRepo.On("DeleteOlderThan", ctx, mock.AnythingOfType("time.Time")).Return(int64(0), expectedErr)

		deleted, err := uc.PurgeOldAuditEvents(ctx, 90)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, int64(0), deleted)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_GetAuditEventsByUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		userID := uuidv7.New()
		params := pagination.Params{Page: 1, Limit: 10}
		expectedEvents := []*entity.AuditEvent{
			{ID: uuidv7.New(), UserID: userID, Action: "user.update"},
		}
		expectedMeta := &pagination.Metadata{Total: 1, CurrentPage: 1}

		mockRepo.On("GetByUserID", ctx, userID, params).Return(expectedEvents, expectedMeta, nil)

		events, meta, err := uc.GetAuditEventsByUser(ctx, userID, params)

		require.NoError(t, err)
		assert.Equal(t, expectedEvents, events)
		assert.Equal(t, expectedMeta, meta)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		userID := uuidv7.New()
		params := pagination.Params{Page: 1, Limit: 10}
		expectedErr := errors.New("database error")

		mockRepo.On("GetByUserID", ctx, userID, params).Return(nil, nil, expectedErr)

		events, meta, err := uc.GetAuditEventsByUser(ctx, userID, params)

		require.Error(t, err)
		assert.Nil(t, events)
		assert.Nil(t, meta)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_GetAuditEventsByAction(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		action := "user.login"
		params := pagination.Params{Page: 1, Limit: 20}
		expectedEvents := []*entity.AuditEvent{
			{ID: uuidv7.New(), Action: action},
			{ID: uuidv7.New(), Action: action},
		}
		expectedMeta := &pagination.Metadata{Total: 2, CurrentPage: 1}

		mockRepo.On("GetByAction", ctx, action, params).Return(expectedEvents, expectedMeta, nil)

		events, meta, err := uc.GetAuditEventsByAction(ctx, action, params)

		require.NoError(t, err)
		assert.Equal(t, expectedEvents, events)
		assert.Equal(t, expectedMeta, meta)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		action := "user.unknown"
		params := pagination.Params{Page: 1, Limit: 10}
		expectedMeta := &pagination.Metadata{Total: 0, CurrentPage: 1}

		mockRepo.On("GetByAction", ctx, action, params).Return([]*entity.AuditEvent{}, expectedMeta, nil)

		events, meta, err := uc.GetAuditEventsByAction(ctx, action, params)

		require.NoError(t, err)
		assert.Empty(t, events)
		assert.Equal(t, expectedMeta, meta)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_VerifyAuditEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("valid signature", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		eventID := uuidv7.New()
		event := &entity.AuditEvent{
			ID:        eventID,
			UserID:    uuidv7.New(),
			Action:    "user.login",
			Signature: "valid_signature",
		}

		mockRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockRepo.On("VerifySignature", ctx, event, "secret").Return(true, nil)

		valid, err := uc.VerifyAuditEvent(ctx, eventID)

		require.NoError(t, err)
		assert.True(t, valid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid signature", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		eventID := uuidv7.New()
		event := &entity.AuditEvent{
			ID:        eventID,
			UserID:    uuidv7.New(),
			Action:    "user.login",
			Signature: "invalid_signature",
		}

		mockRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockRepo.On("VerifySignature", ctx, event, "secret").Return(false, nil)

		valid, err := uc.VerifyAuditEvent(ctx, eventID)

		require.NoError(t, err)
		assert.False(t, valid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		eventID := uuidv7.New()
		expectedErr := entity.ErrAuditEventNotFound

		mockRepo.On("GetByID", ctx, eventID).Return(nil, expectedErr)

		valid, err := uc.VerifyAuditEvent(ctx, eventID)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("verification error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		eventID := uuidv7.New()
		event := &entity.AuditEvent{
			ID:        eventID,
			UserID:    uuidv7.New(),
			Action:    "user.login",
			Signature: "some_signature",
		}
		expectedErr := errors.New("verification failed")

		mockRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockRepo.On("VerifySignature", ctx, event, "secret").Return(false, expectedErr)

		valid, err := uc.VerifyAuditEvent(ctx, eventID)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_CreateAuditEvent_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		data := &entity.AuditEventCreate{
			UserID:     uuidv7.New(),
			Action:     "user.login",
			EntityType: "user",
			EntityID:   uuidv7.New(),
		}

		expectedErr := errors.New("database connection failed")
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AuditEvent")).Return(expectedErr)

		event, err := uc.CreateAuditEvent(ctx, data)

		require.Error(t, err)
		assert.Nil(t, event)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_ListAuditEvents_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		filters := repository.AuditEventFilters{}
		params := pagination.Params{Page: 1, Limit: 10}
		expectedErr := errors.New("database error")

		mockRepo.On("List", ctx, filters, params).Return(nil, nil, expectedErr)

		events, meta, err := uc.ListAuditEvents(ctx, filters, params)

		require.Error(t, err)
		assert.Nil(t, events)
		assert.Nil(t, meta)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventUseCase_GetAuditEventsByEntity_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := NewAuditEventUseCase(mockRepo, "secret")

		entityType := "user"
		entityID := uuidv7.New()
		expectedErr := errors.New("database error")

		mockRepo.On("GetByEntityID", ctx, entityType, entityID).Return(nil, expectedErr)

		events, err := uc.GetAuditEventsByEntity(ctx, entityType, entityID)

		require.Error(t, err)
		assert.Nil(t, events)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
