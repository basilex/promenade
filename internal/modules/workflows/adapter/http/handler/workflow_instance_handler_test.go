package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/internal/modules/workflows/usecase"
	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockWorkflowInstanceUseCase for handler testing
type MockWorkflowInstanceUseCase struct {
	mock.Mock
}

func (m *MockWorkflowInstanceUseCase) StartWorkflow(
	ctx context.Context,
	definitionID uuidv7.UUID,
	externalReference *string,
	priority int,
	workflowContext map[string]interface{},
	assignedTo *uuidv7.UUID,
	startedBy uuidv7.UUID,
) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, definitionID, externalReference, priority, workflowContext, assignedTo, startedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceUseCase) GetInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceUseCase) ListInstances(
	ctx context.Context,
	definitionID *uuidv7.UUID,
	status string,
	assignedTo *uuidv7.UUID,
	limit, offset int,
) ([]*entity.WorkflowInstance, error) {
	args := m.Called(ctx, definitionID, status, assignedTo, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceUseCase) PauseInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceUseCase) ResumeInstance(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceUseCase) CancelInstance(ctx context.Context, id uuidv7.UUID, reason string) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, id, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func (m *MockWorkflowInstanceUseCase) UpdateInstance(
	ctx context.Context,
	id uuidv7.UUID,
	priority *int,
	assignedTo *uuidv7.UUID,
) (*entity.WorkflowInstance, error) {
	args := m.Called(ctx, id, priority, assignedTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowInstance), args.Error(1)
}

func TestStartWorkflow_Success(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	userID := uuidv7.New()
	definitionID := uuidv7.New()
	instanceID := uuidv7.New()

	mockUC.On("StartWorkflow",
		mock.Anything,
		definitionID,
		(*string)(nil),
		5,
		map[string]interface{}{"key": "value"},
		(*uuidv7.UUID)(nil),
		userID,
	).Return(&entity.WorkflowInstance{
		ID:             instanceID,
		DefinitionID:   definitionID,
		Status:         entity.WorkflowInstanceStatusRunning,
		CurrentState:   "start",
		Priority:       5,
		Context:        jsonb.Map{},
		Input:          jsonb.Map{},
		Output:         jsonb.Map{},
		ErrorDetails:   jsonb.Map{},
		StartedBy:      userID,
		StartedAt:      &[]time.Time{time.Now()}[0],
		StateEnteredAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil)

	reqBody := dto.StartInstanceRequest{
		DefinitionID: definitionID.String(),
		Priority:     5,
		Context:      map[string]interface{}{"key": "value"},
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.Use(authMiddleware(userID))
	router.POST("/workflows/instances", handler.StartWorkflow)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, true, response["success"])
	mockUC.AssertExpectations(t)
}

func TestStartWorkflow_MissingAuth(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	definitionID := uuidv7.New()
	reqBody := dto.StartInstanceRequest{
		DefinitionID: definitionID.String(),
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.POST("/workflows/instances", handler.StartWorkflow)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStartWorkflow_WorkflowNotFound(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	userID := uuidv7.New()
	definitionID := uuidv7.New()

	mockUC.On("StartWorkflow",
		mock.Anything,
		definitionID,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		userID,
	).Return(nil, usecase.ErrWorkflowNotFound)

	reqBody := dto.StartInstanceRequest{
		DefinitionID: definitionID.String(),
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.Use(authMiddleware(userID))
	router.POST("/workflows/instances", handler.StartWorkflow)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestStartWorkflow_WorkflowNotActive(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	userID := uuidv7.New()
	definitionID := uuidv7.New()

	mockUC.On("StartWorkflow",
		mock.Anything,
		definitionID,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		userID,
	).Return(nil, usecase.ErrWorkflowNotActive)

	reqBody := dto.StartInstanceRequest{
		DefinitionID: definitionID.String(),
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.Use(authMiddleware(userID))
	router.POST("/workflows/instances", handler.StartWorkflow)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetInstance_Success(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	instanceID := uuidv7.New()
	definitionID := uuidv7.New()
	userID := uuidv7.New()

	mockUC.On("GetInstance", mock.Anything, instanceID).Return(&entity.WorkflowInstance{
		ID:             instanceID,
		DefinitionID:   definitionID,
		Status:         entity.WorkflowInstanceStatusRunning,
		CurrentState:   "start",
		StartedBy:      userID,
		StartedAt:      &[]time.Time{time.Now()}[0],
		StateEnteredAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil)

	router := setupTestRouter()
	router.GET("/workflows/instances/:id", handler.GetInstance)

	req := httptest.NewRequest(http.MethodGet, "/workflows/instances/"+instanceID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetInstance_NotFound(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	instanceID := uuidv7.New()
	mockUC.On("GetInstance", mock.Anything, instanceID).Return(nil, usecase.ErrInstanceNotFound)

	router := setupTestRouter()
	router.GET("/workflows/instances/:id", handler.GetInstance)

	req := httptest.NewRequest(http.MethodGet, "/workflows/instances/"+instanceID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestPauseInstance_Success(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	instanceID := uuidv7.New()
	mockUC.On("PauseInstance", mock.Anything, instanceID).Return(&entity.WorkflowInstance{
		ID:             instanceID,
		Status:         entity.WorkflowInstanceStatusPaused,
		StateEnteredAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil)

	router := setupTestRouter()
	router.POST("/workflows/instances/:id/pause", handler.PauseInstance)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances/"+instanceID.String()+"/pause", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestResumeInstance_Success(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	instanceID := uuidv7.New()
	mockUC.On("ResumeInstance", mock.Anything, instanceID).Return(&entity.WorkflowInstance{
		ID:             instanceID,
		Status:         entity.WorkflowInstanceStatusRunning,
		StateEnteredAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil)

	router := setupTestRouter()
	router.POST("/workflows/instances/:id/resume", handler.ResumeInstance)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances/"+instanceID.String()+"/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCancelInstance_Success(t *testing.T) {
	mockUC := new(MockWorkflowInstanceUseCase)
	handler := NewWorkflowInstanceHandler(mockUC)

	instanceID := uuidv7.New()
	mockUC.On("CancelInstance", mock.Anything, instanceID, "test reason").Return(&entity.WorkflowInstance{
		ID:             instanceID,
		Status:         entity.WorkflowInstanceStatusCancelled,
		StateEnteredAt: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil)

	router := setupTestRouter()
	router.POST("/workflows/instances/:id/cancel", handler.CancelInstance)

	req := httptest.NewRequest(http.MethodPost, "/workflows/instances/"+instanceID.String()+"/cancel?reason=test+reason", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}
