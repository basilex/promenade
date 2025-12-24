package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/internal/modules/audit/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuditEventRepository is a mock for repository testing
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

// MockAuditEventUseCase is a mock implementation of audit event usecase
type MockAuditEventUseCase struct {
	mock.Mock
}

func (m *MockAuditEventUseCase) CreateAuditEvent(ctx interface{}, data *entity.AuditEventCreate) (*entity.AuditEvent, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AuditEvent), args.Error(1)
}

func (m *MockAuditEventUseCase) GetAuditEvent(ctx interface{}, id uuidv7.UUID) (*entity.AuditEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AuditEvent), args.Error(1)
}

func (m *MockAuditEventUseCase) ListAuditEvents(ctx interface{}, filters repository.AuditEventFilters, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	args := m.Called(ctx, filters, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.AuditEvent), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *MockAuditEventUseCase) GetAuditEventsByEntity(ctx interface{}, entityType string, entityID uuidv7.UUID) ([]*entity.AuditEvent, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditEvent), args.Error(1)
}

func (m *MockAuditEventUseCase) VerifyAuditEvent(ctx interface{}, id uuidv7.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func TestAuditEventHandler_CreateAuditEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		userID := uuidv7.New()
		entityID := uuidv7.New()

		req := CreateAuditEventRequest{
			UserID:     userID,
			Action:     "user.login",
			EntityType: "user",
			EntityID:   entityID,
			IPAddress:  "127.0.0.1",
			UserAgent:  "test-agent",
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AuditEvent")).Return(nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/audit/events", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.CreateAuditEvent(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error - missing user_id", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		req := CreateAuditEventRequest{
			Action:     "user.login",
			EntityType: "user",
			EntityID:   uuidv7.New(),
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/audit/events", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.CreateAuditEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response["success"].(bool))
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/audit/events", bytes.NewBuffer([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.CreateAuditEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		userID := uuidv7.New()
		entityID := uuidv7.New()

		req := CreateAuditEventRequest{
			UserID:     userID,
			Action:     "user.login",
			EntityType: "user",
			EntityID:   entityID,
		}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AuditEvent")).Return(errors.New("database error"))

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/audit/events", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.CreateAuditEvent(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventHandler_GetAuditEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("event found", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		eventID := uuidv7.New()
		expectedEvent := &entity.AuditEvent{
			ID:         eventID,
			UserID:     uuidv7.New(),
			Action:     "user.login",
			EntityType: "user",
			EntityID:   uuidv7.New(),
			IPAddress:  "127.0.0.1",
			Signature:  "test-signature",
			CreatedAt:  time.Now(),
		}

		mockRepo.On("GetByID", mock.Anything, eventID).Return(expectedEvent, nil)

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/%s", eventID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: eventID.String()}}

		handler.GetAuditEvent(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		eventID := uuidv7.New()

		mockRepo.On("GetByID", mock.Anything, eventID).Return(nil, entity.ErrAuditEventNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/%s", eventID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: eventID.String()}}

		handler.GetAuditEvent(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events/invalid-uuid", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

		handler.GetAuditEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuditEventHandler_ListAuditEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list with pagination", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		expectedEvents := []*entity.AuditEvent{
			{ID: uuidv7.New(), Action: "user.login"},
			{ID: uuidv7.New(), Action: "user.logout"},
		}
		expectedMeta := &pagination.Metadata{Total: 2, CurrentPage: 1}

		mockRepo.On("List", mock.Anything, mock.AnythingOfType("repository.AuditEventFilters"), mock.AnythingOfType("pagination.Params")).Return(expectedEvents, expectedMeta, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events?page=1&limit=10", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.ListAuditEvents(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		expectedMeta := &pagination.Metadata{Total: 0, CurrentPage: 1}

		mockRepo.On("List", mock.Anything, mock.AnythingOfType("repository.AuditEventFilters"), mock.AnythingOfType("pagination.Params")).Return([]*entity.AuditEvent{}, expectedMeta, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.ListAuditEvents(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventHandler_GetAuditEventsByEntity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get events for entity", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		entityID := uuidv7.New()
		expectedEvents := []*entity.AuditEvent{
			{ID: uuidv7.New(), EntityType: "post", EntityID: entityID, Action: "post.created"},
			{ID: uuidv7.New(), EntityType: "post", EntityID: entityID, Action: "post.updated"},
		}

		mockRepo.On("GetByEntityID", mock.Anything, "post", entityID).Return(expectedEvents, nil)

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/entity/post/%s", entityID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{
			{Key: "entity_type", Value: "post"},
			{Key: "entity_id", Value: entityID.String()},
		}

		handler.GetAuditEventsByEntity(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid entity_id", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events/entity/post/invalid", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{
			{Key: "entity_type", Value: "post"},
			{Key: "entity_id", Value: "invalid"},
		}

		handler.GetAuditEventsByEntity(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		entityID := uuidv7.New()

		mockRepo.On("GetByEntityID", mock.Anything, "post", entityID).Return(nil, errors.New("database error"))

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/entity/post/%s", entityID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{
			{Key: "entity_type", Value: "post"},
			{Key: "entity_id", Value: entityID.String()},
		}

		handler.GetAuditEventsByEntity(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuditEventHandler_VerifyAuditEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid signature", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		eventID := uuidv7.New()
		event := &entity.AuditEvent{
			ID:        eventID,
			Signature: "valid-signature",
		}

		mockRepo.On("GetByID", mock.Anything, eventID).Return(event, nil)
		mockRepo.On("VerifySignature", mock.Anything, event, "test-secret").Return(true, nil)

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/%s/verify", eventID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: eventID.String()}}

		handler.VerifyAuditEvent(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		assert.True(t, data["valid"].(bool))

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid signature", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		eventID := uuidv7.New()
		event := &entity.AuditEvent{
			ID:        eventID,
			Signature: "invalid-signature",
		}

		mockRepo.On("GetByID", mock.Anything, eventID).Return(event, nil)
		mockRepo.On("VerifySignature", mock.Anything, event, "test-secret").Return(false, nil)

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/%s/verify", eventID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: eventID.String()}}

		handler.VerifyAuditEvent(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		assert.False(t, data["valid"].(bool))

		mockRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		eventID := uuidv7.New()

		mockRepo.On("GetByID", mock.Anything, eventID).Return(nil, entity.ErrAuditEventNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/audit/events/%s/verify", eventID), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: eventID.String()}}

		handler.VerifyAuditEvent(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		mockRepo := new(MockAuditEventRepository)
		uc := usecase.NewAuditEventUseCase(mockRepo, "test-secret")
		handler := NewAuditEventHandler(uc)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events/invalid/verify", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.VerifyAuditEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuditEventHandler_toResponse(t *testing.T) {
	t.Run("converts audit event to response", func(t *testing.T) {
		handler := &AuditEventHandler{}

		eventID := uuidv7.New()
		userID := uuidv7.New()
		entityID := uuidv7.New()

		event := &entity.AuditEvent{
			ID:         eventID,
			UserID:     userID,
			Action:     "user.login",
			EntityType: "user",
			EntityID:   entityID,
			IPAddress:  "127.0.0.1",
			UserAgent:  "test-agent",
			Signature:  "test-signature",
		}

		resp := handler.toResponse(event)

		assert.NotNil(t, resp)
		assert.Equal(t, eventID, resp.ID)
		assert.Equal(t, userID, resp.UserID)
		assert.Equal(t, "user.login", resp.Action)
		assert.Equal(t, "user", resp.EntityType)
		assert.Equal(t, entityID, resp.EntityID)
		assert.Equal(t, "127.0.0.1", resp.IPAddress)
		assert.Equal(t, "test-agent", resp.UserAgent)
		assert.Equal(t, "test-signature", resp.Signature)
	})

	t.Run("handles optional fields", func(t *testing.T) {
		handler := &AuditEventHandler{}

		oldData := `{"name":"old"}`
		newData := `{"name":"new"}`
		metadata := `{"key":"value"}`

		event := &entity.AuditEvent{
			ID:         uuidv7.New(),
			UserID:     uuidv7.New(),
			Action:     "user.update",
			EntityType: "user",
			EntityID:   uuidv7.New(),
			OldData:    &oldData,
			NewData:    &newData,
			Metadata:   &metadata,
		}

		resp := handler.toResponse(event)

		assert.NotNil(t, resp)
		assert.Equal(t, &oldData, resp.OldData)
		assert.Equal(t, &newData, resp.NewData)
		assert.Equal(t, &metadata, resp.Metadata)
	})
}
