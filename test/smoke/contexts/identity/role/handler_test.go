package role_test

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

	"github.com/basilex/promenade/internal/contexts/identity/role"
	roleHTTP "github.com/basilex/promenade/internal/contexts/identity/role/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockRoleUseCase is a mock implementation of role.IUseCase for smoke tests
type MockRoleUseCase struct {
	mock.Mock
}

func (m *MockRoleUseCase) CreateRole(ctx context.Context, name, displayName, description string) (*role.Role, error) {
	args := m.Called(ctx, name, displayName, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.Role), args.Error(1)
}

func (m *MockRoleUseCase) GetRole(ctx context.Context, roleID uuidv7.UUID) (*role.Role, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.Role), args.Error(1)
}

func (m *MockRoleUseCase) GetRoleByName(ctx context.Context, name string) (*role.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.Role), args.Error(1)
}

func (m *MockRoleUseCase) UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*role.Role, error) {
	args := m.Called(ctx, roleID, displayName, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.Role), args.Error(1)
}

func (m *MockRoleUseCase) DeleteRole(ctx context.Context, roleID uuidv7.UUID) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

func (m *MockRoleUseCase) ListRoles(ctx context.Context, limit, offset int) ([]*role.Role, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*role.Role), args.Int(1), args.Error(2)
}

func (m *MockRoleUseCase) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*role.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*role.Role), args.Error(1)
}

func setupRoleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestHandler_CreateRole tests role creation endpoint
func TestHandler_CreateRole(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.POST("/roles", handler.Create)

	testRole := &role.Role{
		ID:          uuidv7.New(),
		Name:        "manager",
		DisplayName: "Manager",
		Description: "Manager role",
		IsSystem:    false,
	}

	mockUC.On("CreateRole", mock.Anything, "manager", "Manager", "Manager role").Return(testRole, nil)

	requestBody := map[string]interface{}{
		"name":         "manager",
		"display_name": "Manager",
		"description":  "Manager role",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(body))
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

// TestHandler_GetRoleByID tests getting role by ID
func TestHandler_GetRoleByID(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.GET("/roles/:id", handler.GetByID)

	roleID := uuidv7.New()
	testRole := &role.Role{
		ID:          roleID,
		Name:        "manager",
		DisplayName: "Manager",
		Description: "Manager role",
		IsSystem:    false,
	}

	mockUC.On("GetRole", mock.Anything, roleID).Return(testRole, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String(), nil)
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

// TestHandler_GetRoleByName tests getting role by name
func TestHandler_GetRoleByName(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.GET("/roles/name/:name", handler.GetByName)

	testRole := &role.Role{
		ID:          uuidv7.New(),
		Name:        "manager",
		DisplayName: "Manager",
		Description: "Manager role",
		IsSystem:    false,
	}

	mockUC.On("GetRoleByName", mock.Anything, "manager").Return(testRole, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/name/manager", nil)
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

// TestHandler_UpdateRole tests role update endpoint
func TestHandler_UpdateRole(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.PUT("/roles/:id", handler.Update)

	roleID := uuidv7.New()
	updatedRole := &role.Role{
		ID:          roleID,
		Name:        "manager",
		DisplayName: "Senior Manager",
		Description: "Updated description",
		IsSystem:    false,
	}

	mockUC.On("UpdateRole", mock.Anything, roleID, "Senior Manager", "Updated description").Return(updatedRole, nil)

	requestBody := map[string]interface{}{
		"display_name": "Senior Manager",
		"description":  "Updated description",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/roles/"+roleID.String(), bytes.NewBuffer(body))
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

// TestHandler_DeleteRole tests role deletion endpoint
func TestHandler_DeleteRole(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.DELETE("/roles/:id", handler.Delete)

	roleID := uuidv7.New()

	mockUC.On("DeleteRole", mock.Anything, roleID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUC.AssertExpectations(t)
}

// TestHandler_ListRoles tests role listing endpoint
func TestHandler_ListRoles(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.GET("/roles", handler.List)

	roles := []*role.Role{
		{
			ID:          uuidv7.New(),
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "Admin role",
			IsSystem:    true,
		},
		{
			ID:          uuidv7.New(),
			Name:        "manager",
			DisplayName: "Manager",
			Description: "Manager role",
			IsSystem:    false,
		},
	}

	mockUC.On("ListRoles", mock.Anything, 20, 0).Return(roles, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
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

// TestHandler_GetUserRoles tests getting user roles endpoint
func TestHandler_GetUserRoles(t *testing.T) {
	mockUC := new(MockRoleUseCase)
	handler := roleHTTP.NewRoleHandler(mockUC)
	router := setupRoleRouter()
	router.GET("/users/:userId/roles", handler.GetUserRoles)

	userID := uuidv7.New()
	roles := []*role.Role{
		{
			ID:          uuidv7.New(),
			Name:        "user",
			DisplayName: "User",
			Description: "Regular user",
			IsSystem:    true,
		},
		{
			ID:          uuidv7.New(),
			Name:        "manager",
			DisplayName: "Manager",
			Description: "Manager role",
			IsSystem:    false,
		},
	}

	mockUC.On("GetUserRoles", mock.Anything, userID).Return(roles, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String()+"/roles", nil)
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
