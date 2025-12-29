package permission_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	permissionHTTP "github.com/basilex/promenade/internal/contexts/identity/permission/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockPermissionUseCase is a mock implementation of permission.IUseCase for smoke tests
type MockPermissionUseCase struct {
	mock.Mock
}

func (m *MockPermissionUseCase) CreatePermission(ctx context.Context, resource, action, description string) (*permission.Permission, error) {
	args := m.Called(ctx, resource, action, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permission.Permission), args.Error(1)
}

func (m *MockPermissionUseCase) GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*permission.Permission, error) {
	args := m.Called(ctx, permissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permission.Permission), args.Error(1)
}

func (m *MockPermissionUseCase) GetPermissionByName(ctx context.Context, name string) (*permission.Permission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permission.Permission), args.Error(1)
}

func (m *MockPermissionUseCase) UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*permission.Permission, error) {
	args := m.Called(ctx, permissionID, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permission.Permission), args.Error(1)
}

func (m *MockPermissionUseCase) DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error {
	args := m.Called(ctx, permissionID)
	return args.Error(0)
}

func (m *MockPermissionUseCase) ListPermissions(ctx context.Context, limit, offset int) ([]*permission.Permission, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*permission.Permission), args.Int(1), args.Error(2)
}

func (m *MockPermissionUseCase) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*permission.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*permission.Permission), args.Error(1)
}

func setupPermissionRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestHandler_CreatePermission tests permission creation endpoint
func TestHandler_CreatePermission(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.POST("/permissions", handler.Create)

	testPermission := &permission.Permission{
		ID:          uuidv7.New(),
		Name:        "users:create",
		Resource:    "users",
		Action:      "create",
		Description: "Create users",
	}

	mockUC.On("CreatePermission", mock.Anything, "users", "create", "Create users").Return(testPermission, nil)

	requestBody := map[string]interface{}{
		"resource":    "users",
		"action":      "create",
		"description": "Create users",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
}

// TestHandler_GetPermissionByID tests getting permission by ID
func TestHandler_GetPermissionByID(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.GET("/permissions/:id", handler.GetByID)

	permissionID := uuidv7.New()
	testPermission := &permission.Permission{
		ID:          permissionID,
		Name:        "users:read",
		Resource:    "users",
		Action:      "read",
		Description: "Read users",
	}

	mockUC.On("GetPermission", mock.Anything, permissionID).Return(testPermission, nil)

	req := httptest.NewRequest(http.MethodGet, "/permissions/"+permissionID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
}

// TestHandler_GetPermissionByName tests getting permission by name
func TestHandler_GetPermissionByName(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.GET("/permissions/name/:name", handler.GetByName)

	testPermission := &permission.Permission{
		ID:          uuidv7.New(),
		Name:        "users:update",
		Resource:    "users",
		Action:      "update",
		Description: "Update users",
	}

	mockUC.On("GetPermissionByName", mock.Anything, "users:update").Return(testPermission, nil)

	req := httptest.NewRequest(http.MethodGet, "/permissions/name/users:update", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
}

// TestHandler_UpdatePermission tests permission update endpoint
func TestHandler_UpdatePermission(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.PUT("/permissions/:id", handler.Update)

	permissionID := uuidv7.New()
	updatedPermission := &permission.Permission{
		ID:          permissionID,
		Name:        "users:delete",
		Resource:    "users",
		Action:      "delete",
		Description: "Updated description for delete",
	}

	mockUC.On("UpdatePermission", mock.Anything, permissionID, "Updated description for delete").Return(updatedPermission, nil)

	requestBody := map[string]interface{}{
		"description": "Updated description for delete",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/permissions/"+permissionID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
}

// TestHandler_DeletePermission tests permission deletion endpoint
func TestHandler_DeletePermission(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.DELETE("/permissions/:id", handler.Delete)

	permissionID := uuidv7.New()

	mockUC.On("DeletePermission", mock.Anything, permissionID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/permissions/"+permissionID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUC.AssertExpectations(t)
}

// TestHandler_ListPermissions tests permission listing endpoint
func TestHandler_ListPermissions(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.GET("/permissions", handler.List)

	permissions := []*permission.Permission{
		{
			ID:          uuidv7.New(),
			Name:        "users:create",
			Resource:    "users",
			Action:      "create",
			Description: "Create users",
		},
		{
			ID:          uuidv7.New(),
			Name:        "users:read",
			Resource:    "users",
			Action:      "read",
			Description: "Read users",
		},
	}

	mockUC.On("ListPermissions", mock.Anything, 20, 0).Return(permissions, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
}

// TestHandler_GetRolePermissions tests getting role permissions endpoint
func TestHandler_GetRolePermissions(t *testing.T) {
	mockUC := new(MockPermissionUseCase)
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router := setupPermissionRouter()
	router.GET("/roles/:roleId/permissions", handler.GetRolePermissions)

	roleID := uuidv7.New()
	permissions := []*permission.Permission{
		{
			ID:          uuidv7.New(),
			Name:        "users:read",
			Resource:    "users",
			Action:      "read",
			Description: "Read users",
		},
		{
			ID:          uuidv7.New(),
			Name:        "users:update",
			Resource:    "users",
			Action:      "update",
			Description: "Update users",
		},
	}

	mockUC.On("GetRolePermissions", mock.Anything, roleID).Return(permissions, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String()+"/permissions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
}
