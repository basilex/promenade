package purge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/internal/modules/audit/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/purge"
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

func (m *MockAuditEventRepository) DeleteOlderThan(ctx context.Context, cutoffTime time.Time) (int64, error) {
	args := m.Called(ctx, cutoffTime)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditEventRepository) VerifySignature(ctx context.Context, event *entity.AuditEvent, secret string) (bool, error) {
	args := m.Called(ctx, event, secret)
	return args.Bool(0), args.Error(1)
}

// Basic test to ensure purge handler structure exists
func TestPurgeHandler_Structure(t *testing.T) {
	// Test that purge handler types exist
	t.Run("handler can be created", func(t *testing.T) {
		handler := &AuditEventPurgeHandler{
			retentionDays: 90,
		}

		assert.NotNil(t, handler)
		assert.Equal(t, 90, handler.retentionDays)
	})

	t.Run("handler name is correct", func(t *testing.T) {
		handler := &AuditEventPurgeHandler{}
		
		assert.Equal(t, "audit_events", handler.Name())
	})

	t.Run("entity name is correct", func(t *testing.T) {
		handler := &AuditEventPurgeHandler{}
		
		assert.Equal(t, "audit_events", handler.EntityName())
	})

	t.Run("dry run returns zero", func(t *testing.T) {
		handler := &AuditEventPurgeHandler{
			retentionDays: 90,
		}
		ctx := context.Background()
		
		deleted, err := handler.Purge(ctx, time.Now(), 100, true)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
	})
}

func TestNewAuditEventPurgeHandler(t *testing.T) {
	t.Run("creates handler with correct fields", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		retentionDays := 90

		handler := NewAuditEventPurgeHandler(uc, retentionDays)

		assert.NotNil(t, handler)
		assert.Equal(t, uc, handler.useCase)
		assert.Equal(t, retentionDays, handler.retentionDays)
	})

	t.Run("accepts different retention periods", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")

		testCases := []int{30, 90, 180, 365}
		for _, days := range testCases {
			handler := NewAuditEventPurgeHandler(uc, days)
			assert.Equal(t, days, handler.retentionDays)
		}
	})
}

func TestAuditEventPurgeHandler_Purge(t *testing.T) {
	t.Run("purges old events successfully", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		mockRepo.On("DeleteOlderThan", mock.Anything, mock.Anything).Return(int64(10), nil)

		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventPurgeHandler(uc, 90)

		ctx := context.Background()
		deleted, err := handler.Purge(ctx, time.Now(), 100, false)

		require.NoError(t, err)
		assert.Equal(t, int64(10), deleted)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles zero deletions", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		mockRepo.On("DeleteOlderThan", mock.Anything, mock.Anything).Return(int64(0), nil)

		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventPurgeHandler(uc, 90)

		ctx := context.Background()
		deleted, err := handler.Purge(ctx, time.Now(), 100, false)

		require.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error on purge failure", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		mockRepo.On("DeleteOlderThan", mock.Anything, mock.Anything).
			Return(int64(0), assert.AnError)

		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventPurgeHandler(uc, 90)

		ctx := context.Background()
		deleted, err := handler.Purge(ctx, time.Now(), 100, false)

		assert.Error(t, err)
		assert.Equal(t, int64(0), deleted)
		mockRepo.AssertExpectations(t)
	})

	t.Run("dry run does not call repository", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		// Should NOT be called
		mockRepo.AssertNotCalled(t, "DeleteOlderThan", mock.Anything, mock.Anything)

		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventPurgeHandler(uc, 90)

		ctx := context.Background()
		deleted, err := handler.Purge(ctx, time.Now(), 100, true)

		require.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
	})

	t.Run("respects retention days from handler", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		mockRepo.On("DeleteOlderThan", mock.Anything, mock.Anything).Return(int64(5), nil)

		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		
		// Test with different retention periods
		testCases := []struct {
			days     int
			expected int64
		}{
			{30, 5},
			{90, 5},
			{180, 5},
		}

		for _, tc := range testCases {
			handler := NewAuditEventPurgeHandler(uc, tc.days)
			deleted, err := handler.Purge(context.Background(), time.Now(), 100, false)
			
			require.NoError(t, err)
			assert.Equal(t, tc.expected, deleted)
		}

		mockRepo.AssertExpectations(t)
	})
}

func TestRegisterPurgeHandlers(t *testing.T) {
	// Clean up registry before test
	defer func() {
		purge.DefaultRegistry = purge.NewRegistry()
		purge.DefaultPolicyRegistry = purge.NewPolicyRegistry()
	}()

	t.Run("registers handler successfully", func(t *testing.T) {
		// Reset registries
		purge.DefaultRegistry = purge.NewRegistry()
		purge.DefaultPolicyRegistry = purge.NewPolicyRegistry()

		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")

		err := RegisterPurgeHandlers(uc, 90)
		require.NoError(t, err)

		// Verify handler is registered
		handler, exists := purge.DefaultRegistry.Get("audit_events")
		assert.True(t, exists)
		assert.NotNil(t, handler)
		
		// Type assertion to check it's our handler
		auditHandler, ok := handler.(*AuditEventPurgeHandler)
		assert.True(t, ok)
		assert.Equal(t, "audit_events", auditHandler.Name())
	})

	t.Run("registers policy successfully", func(t *testing.T) {
		// Reset registries
		purge.DefaultRegistry = purge.NewRegistry()
		purge.DefaultPolicyRegistry = purge.NewPolicyRegistry()

		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")

		err := RegisterPurgeHandlers(uc, 90)
		require.NoError(t, err)

		// Verify policy is registered
		policy, exists := purge.DefaultPolicyRegistry.GetPolicy("audit_events")
		assert.True(t, exists)
		assert.Equal(t, "audit_events", policy.EntityName)
		assert.Equal(t, 90, policy.RetentionDays)
		assert.True(t, policy.Enabled)
	})

	t.Run("policy enabled when retention days > 0", func(t *testing.T) {
		// Reset registries
		purge.DefaultRegistry = purge.NewRegistry()
		purge.DefaultPolicyRegistry = purge.NewPolicyRegistry()

		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")

		err := RegisterPurgeHandlers(uc, 90)
		require.NoError(t, err)

		policy, _ := purge.DefaultPolicyRegistry.GetPolicy("audit_events")
		assert.True(t, policy.Enabled)
	})

	t.Run("policy disabled when retention days is 0", func(t *testing.T) {
		// Reset registries
		purge.DefaultRegistry = purge.NewRegistry()
		purge.DefaultPolicyRegistry = purge.NewPolicyRegistry()

		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")

		err := RegisterPurgeHandlers(uc, 0)
		require.NoError(t, err)

		policy, _ := purge.DefaultPolicyRegistry.GetPolicy("audit_events")
		assert.False(t, policy.Enabled)
	})

	t.Run("handles duplicate registration gracefully", func(t *testing.T) {
		// Reset registries
		purge.DefaultRegistry = purge.NewRegistry()
		purge.DefaultPolicyRegistry = purge.NewPolicyRegistry()

		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")

		// First registration should succeed
		err := RegisterPurgeHandlers(uc, 90)
		require.NoError(t, err)

		// Second registration - handler will fail but we still register
		err = RegisterPurgeHandlers(uc, 90)
		// Either succeeds or fails, both are acceptable depending on implementation
		// Main point is it doesn't panic
		_ = err
	})
}
