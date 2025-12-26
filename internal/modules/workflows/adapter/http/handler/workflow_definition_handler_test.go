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

// MockWorkflowDefinitionUseCase for handler testing
type MockWorkflowDefinitionUseCase struct {
	mock.Mock
}

func (m *MockWorkflowDefinitionUseCase) Create(
	ctx context.Context,
	name, description, version string,
	schema entity.WorkflowSchema,
	tags []string,
	metadata map[string]interface{},
	createdBy uuidv7.UUID,
) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, name, description, version, schema, tags, metadata, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionUseCase) GetByName(ctx context.Context, name string) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionUseCase) Activate(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionUseCase) List(
	ctx context.Context,
	status *entity.WorkflowDefinitionStatus,
	category *string,
	creatorID *uuidv7.UUID,
	limit, offset int,
) ([]*entity.WorkflowDefinition, int64, error) {
	args := m.Called(ctx, status, category, creatorID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.WorkflowDefinition), int64(args.Int(1)), args.Error(2)
}

func (m *MockWorkflowDefinitionUseCase) Update(
	ctx context.Context,
	id uuidv7.UUID,
	name, description *string,
	schema *entity.WorkflowSchema,
	tags []string,
) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id, name, description, schema, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkflowDefinitionUseCase) Deprecate(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func (m *MockWorkflowDefinitionUseCase) Archive(ctx context.Context, id uuidv7.UUID, cancelRunningInstances bool) (*entity.WorkflowDefinition, error) {
	args := m.Called(ctx, id, cancelRunningInstances)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowDefinition), args.Error(1)
}

func TestCreateDefinition_Success(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	userID := uuidv7.New()
	definitionID := uuidv7.New()

	schema := entity.WorkflowSchema{
		States: []entity.WorkflowState{
			{
				Name: "start",
				Type: "activity",
			},
			{
				Name:    "end",
				Type:    "final",
				IsFinal: true,
			},
		},
		Transitions: []entity.WorkflowTransition{
			{
				From:  "start",
				To:    "end",
				Event: "complete",
			},
		},
		InitialState: "start",
	}

	mockUC.On("Create",
		mock.Anything,
		"Test Workflow",
		"Test description",
		"",
		schema,
		[]string{"test"},
		map[string]interface{}{"key": "value"},
		userID,
	).Return(&entity.WorkflowDefinition{
		ID:          definitionID,
		Name:        "Test Workflow",
		DisplayName: "Test Workflow",
		Description: "Test description",
		Version:     1,
		Status:      entity.WorkflowDefinitionStatusDraft,
		Category:    "general",
		Definition:  jsonb.JSON[entity.WorkflowSchema]{Data: schema},
		Tags:        []string{"test"},
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	reqBody := dto.CreateDefinitionRequest{
		Name:        "Test Workflow",
		Description: "Test description",
		Schema:      schema,
		Tags:        []string{"test"},
		Metadata:    map[string]interface{}{"key": "value"},
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.Use(authMiddleware(userID))
	router.POST("/workflows/definitions", handler.CreateDefinition)

	req := httptest.NewRequest(http.MethodPost, "/workflows/definitions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Debug: print response
	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, true, response["success"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, definitionID.String(), data["id"])
	assert.Equal(t, "Test Workflow", data["name"])

	mockUC.AssertExpectations(t)
}

func TestCreateDefinition_MissingAuth(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "activity"}},
		InitialState: "start",
	}

	reqBody := dto.CreateDefinitionRequest{
		Name:   "Test Workflow",
		Schema: schema,
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.POST("/workflows/definitions", handler.CreateDefinition)

	req := httptest.NewRequest(http.MethodPost, "/workflows/definitions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateDefinition_InvalidJSON(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)
	userID := uuidv7.New()

	router := setupTestRouter()
	router.Use(authMiddleware(userID))
	router.POST("/workflows/definitions", handler.CreateDefinition)

	req := httptest.NewRequest(http.MethodPost, "/workflows/definitions", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDefinition_NameAlreadyExists(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)
	userID := uuidv7.New()

	schema := entity.WorkflowSchema{
		States:       []entity.WorkflowState{{Name: "start", Type: "activity"}},
		InitialState: "start",
	}

	mockUC.On("Create",
		mock.Anything,
		"Existing Workflow",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		userID,
	).Return(nil, usecase.ErrWorkflowNameAlreadyExists)

	reqBody := dto.CreateDefinitionRequest{
		Name:   "Existing Workflow",
		Schema: schema,
	}
	body, _ := json.Marshal(reqBody)

	router := setupTestRouter()
	router.Use(authMiddleware(userID))
	router.POST("/workflows/definitions", handler.CreateDefinition)

	req := httptest.NewRequest(http.MethodPost, "/workflows/definitions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetDefinition_Success(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	definitionID := uuidv7.New()
	userID := uuidv7.New()

	mockUC.On("GetByID", mock.Anything, definitionID).Return(&entity.WorkflowDefinition{
		ID:          definitionID,
		Name:        "Test Workflow",
		DisplayName: "Test Workflow",
		Description: "Test description",
		Version:     1,
		Status:      entity.WorkflowDefinitionStatusActive,
		Category:    "general",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	router := setupTestRouter()
	router.GET("/workflows/definitions/:id", handler.GetDefinition)

	req := httptest.NewRequest(http.MethodGet, "/workflows/definitions/"+definitionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, true, response["success"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, definitionID.String(), data["id"])

	mockUC.AssertExpectations(t)
}

func TestGetDefinition_InvalidUUID(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	router := setupTestRouter()
	router.GET("/workflows/definitions/:id", handler.GetDefinition)

	req := httptest.NewRequest(http.MethodGet, "/workflows/definitions/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDefinition_NotFound(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	definitionID := uuidv7.New()
	mockUC.On("GetByID", mock.Anything, definitionID).Return(nil, usecase.ErrWorkflowNotFound)

	router := setupTestRouter()
	router.GET("/workflows/definitions/:id", handler.GetDefinition)

	req := httptest.NewRequest(http.MethodGet, "/workflows/definitions/"+definitionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetDefinitionByName_Success(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	definitionID := uuidv7.New()
	userID := uuidv7.New()

	mockUC.On("GetByName", mock.Anything, "test-workflow").Return(&entity.WorkflowDefinition{
		ID:          definitionID,
		Name:        "test-workflow",
		DisplayName: "Test Workflow",
		Version:     1,
		Status:      entity.WorkflowDefinitionStatusActive,
		Category:    "general",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	router := setupTestRouter()
	router.GET("/workflows/definitions/name/:name", handler.GetDefinitionByName)

	req := httptest.NewRequest(http.MethodGet, "/workflows/definitions/name/test-workflow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestActivateDefinition_Success(t *testing.T) {
	mockUC := new(MockWorkflowDefinitionUseCase)
	handler := NewWorkflowDefinitionHandler(mockUC)

	definitionID := uuidv7.New()
	userID := uuidv7.New()

	mockUC.On("Activate", mock.Anything, definitionID).Return(&entity.WorkflowDefinition{
		ID:          definitionID,
		Name:        "Test Workflow",
		DisplayName: "Test Workflow",
		Version:     1,
		Status:      entity.WorkflowDefinitionStatusActive,
		Category:    "general",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	router := setupTestRouter()
	router.POST("/workflows/definitions/:id/activate", handler.ActivateDefinition)

	req := httptest.NewRequest(http.MethodPost, "/workflows/definitions/"+definitionID.String()+"/activate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "active", data["status"])

	mockUC.AssertExpectations(t)
}
