package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/purge"
)

// Mock Event Bus
type mockEventBus struct {
	mock.Mock
}

func (m *mockEventBus) Publish(ctx context.Context, topic string, event bus.Event) error {
	args := m.Called(ctx, topic, event)
	return args.Error(0)
}

func (m *mockEventBus) Subscribe(topic string, handler bus.Handler) error {
	args := m.Called(topic, handler)
	return args.Error(0)
}

func (m *mockEventBus) Unsubscribe(topic string, handler bus.Handler) error {
	args := m.Called(topic, handler)
	return args.Error(0)
}

func (m *mockEventBus) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockEventBus) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock purge handler
type mockPurgeHandler struct {
	mock.Mock
	entityName string
}

func (m *mockPurgeHandler) EntityName() string {
	if m.entityName != "" {
		return m.entityName
	}
	return "test_entity"
}

func (m *mockPurgeHandler) Purge(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error) {
	args := m.Called(ctx, cutoffDate, batchSize, dryRun)
	return args.Get(0).(int64), args.Error(1)
}

// Test helpers
func setupPurgeUseCase(t *testing.T, policies []entity.RetentionPolicy) (*purgeUseCase, *purge.Registry, *mockEventBus) {
	registry := purge.NewRegistry()
	eventBus := new(mockEventBus)
	batchSize := 100

	uc := NewPurgeUseCase(registry, policies, batchSize, eventBus).(*purgeUseCase)
	return uc, registry, eventBus
}

func createTestPolicy(entityName string, retentionDays int, enabled bool) entity.RetentionPolicy {
	return entity.RetentionPolicy{
		EntityName:    entityName,
		RetentionDays: retentionDays,
		Enabled:       enabled,
	}
}

// Tests for PurgeEntity
func TestPurgeUseCase_PurgeEntity_Success(t *testing.T) {
	policy := createTestPolicy("test_entity", 30, true)
	uc, registry, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy})
	ctx := context.Background()

	handler := &mockPurgeHandler{entityName: "test_entity"}
	registry.Register(handler)

	handler.On("Purge", ctx, mock.AnythingOfType("time.Time"), 100, false).Return(int64(10), nil)

	result, err := uc.PurgeEntity(ctx, "test_entity", false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test_entity", result.EntityName)
	assert.Equal(t, int64(10), result.RecordsPurged)
	assert.False(t, result.DryRun)
	handler.AssertExpectations(t)
}

func TestPurgeUseCase_PurgeEntity_PolicyNotFound(t *testing.T) {
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{})
	ctx := context.Background()

	result, err := uc.PurgeEntity(ctx, "nonexistent_entity", false)

	assert.Error(t, err)
	assert.Equal(t, ErrPolicyNotFound, err)
	assert.Nil(t, result)
}

func TestPurgeUseCase_PurgeEntity_PolicyDisabled(t *testing.T) {
	policy := createTestPolicy("test_entity", 30, false)
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy})
	ctx := context.Background()

	result, err := uc.PurgeEntity(ctx, "test_entity", false)

	assert.Error(t, err)
	assert.Equal(t, ErrPolicyDisabled, err)
	assert.Nil(t, result)
}

func TestPurgeUseCase_PurgeEntity_InvalidRetention(t *testing.T) {
	policy := createTestPolicy("test_entity", -1, true)
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy})
	ctx := context.Background()

	result, err := uc.PurgeEntity(ctx, "test_entity", false)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRetention, err)
	assert.Nil(t, result)
}

// Tests for PurgeAll
func TestPurgeUseCase_PurgeAll_Success(t *testing.T) {
	policy1 := createTestPolicy("entity1", 30, true)
	policy2 := createTestPolicy("entity2", 60, true)
	uc, registry, eventBus := setupPurgeUseCase(t, []entity.RetentionPolicy{policy1, policy2})
	ctx := context.Background()

	handler1 := &mockPurgeHandler{entityName: "entity1"}
	handler2 := &mockPurgeHandler{entityName: "entity2"}
	registry.Register(handler1)
	registry.Register(handler2)

	handler1.On("Purge", ctx, mock.AnythingOfType("time.Time"), 100, false).Return(int64(10), nil)
	handler2.On("Purge", ctx, mock.AnythingOfType("time.Time"), 100, false).Return(int64(20), nil)
	eventBus.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)

	summary, err := uc.PurgeAll(ctx, false)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Len(t, summary.Results, 2)
	assert.Equal(t, int64(30), summary.TotalRecordsPurged)
	handler1.AssertExpectations(t)
	handler2.AssertExpectations(t)
}

// Tests for GetRetentionPolicies
func TestPurgeUseCase_GetRetentionPolicies_Success(t *testing.T) {
	policy1 := createTestPolicy("entity1", 30, true)
	policy2 := createTestPolicy("entity2", 60, false)
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy1, policy2})
	ctx := context.Background()

	policies, err := uc.GetRetentionPolicies(ctx)

	assert.NoError(t, err)
	assert.Len(t, policies, 2)
}

func TestPurgeUseCase_GetRetentionPolicies_Empty(t *testing.T) {
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{})
	ctx := context.Background()

	policies, err := uc.GetRetentionPolicies(ctx)

	assert.NoError(t, err)
	assert.Empty(t, policies)
}

// Tests for GetRetentionPolicy
func TestPurgeUseCase_GetRetentionPolicy_Success(t *testing.T) {
	policy := createTestPolicy("test_entity", 30, true)
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy})
	ctx := context.Background()

	result, err := uc.GetRetentionPolicy(ctx, "test_entity")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test_entity", result.EntityName)
	assert.Equal(t, 30, result.RetentionDays)
}

func TestPurgeUseCase_GetRetentionPolicy_NotFound(t *testing.T) {
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{})
	ctx := context.Background()

	result, err := uc.GetRetentionPolicy(ctx, "nonexistent_entity")

	assert.Error(t, err)
	assert.Equal(t, ErrPolicyNotFound, err)
	assert.Nil(t, result)
}

// Tests for PreviewPurge
func TestPurgeUseCase_PreviewPurge_Success(t *testing.T) {
	policy := createTestPolicy("test_entity", 30, true)
	uc, registry, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy})
	ctx := context.Background()

	handler := &mockPurgeHandler{entityName: "test_entity"}
	registry.Register(handler)

	handler.On("Purge", ctx, mock.AnythingOfType("time.Time"), 100, true).Return(int64(100), nil)

	count, err := uc.PreviewPurge(ctx, "test_entity")

	assert.NoError(t, err)
	assert.Equal(t, int64(100), count)
	handler.AssertExpectations(t)
}

func TestPurgeUseCase_PreviewPurge_PolicyNotFound(t *testing.T) {
	uc, _, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{})
	ctx := context.Background()

	count, err := uc.PreviewPurge(ctx, "nonexistent_entity")

	assert.Error(t, err)
	assert.Equal(t, ErrPolicyNotFound, err)
	assert.Equal(t, int64(0), count)
}

func TestPurgeUseCase_PreviewPurge_PolicyDisabled(t *testing.T) {
	policy := createTestPolicy("test_entity", 30, true) // enabled=true since disabled check is in PurgeEntity, not PreviewPurge
	uc, registry, _ := setupPurgeUseCase(t, []entity.RetentionPolicy{policy})
	ctx := context.Background()

	handler := &mockPurgeHandler{entityName: "test_entity"}
	registry.Register(handler)

	// PreviewPurge calls Purge with dryRun=true
	handler.On("Purge", ctx, mock.AnythingOfType("time.Time"), 100, true).Return(int64(0), ErrPolicyDisabled)

	count, err := uc.PreviewPurge(ctx, "test_entity")

	assert.Error(t, err)
	assert.Equal(t, ErrPolicyDisabled, err)
	assert.Equal(t, int64(0), count)
	handler.AssertExpectations(t)
}
