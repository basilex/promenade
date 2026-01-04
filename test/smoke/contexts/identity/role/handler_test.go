package role_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	roleHTTP "github.com/basilex/promenade/internal/contexts/identity/role/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/stretchr/testify/assert"
)

// MockRoleUseCase mocks role.IUseCase
type MockRoleUseCase struct {
	CreateRoleFunc    func(ctx context.Context, name, displayName, description string) (*role.Role, error)
	GetRoleFunc       func(ctx context.Context, roleID uuidv7.UUID) (*role.Role, error)
	GetRoleByNameFunc func(ctx context.Context, name string) (*role.Role, error)
	UpdateRoleFunc    func(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*role.Role, error)
	DeleteRoleFunc    func(ctx context.Context, roleID uuidv7.UUID) error
	ListRolesFunc     func(ctx context.Context, limit, offset int) ([]*role.Role, int, error)
	GetUserRolesFunc  func(ctx context.Context, userID uuidv7.UUID) ([]*role.Role, error)
}

func (m *MockRoleUseCase) CreateRole(ctx context.Context, name, displayName, description string) (*role.Role, error) {
	if m.CreateRoleFunc != nil {
		return m.CreateRoleFunc(ctx, name, displayName, description)
	}
	return nil, nil
}

func (m *MockRoleUseCase) GetRole(ctx context.Context, roleID uuidv7.UUID) (*role.Role, error) {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc(ctx, roleID)
	}
	return nil, nil
}

func (m *MockRoleUseCase) GetRoleByName(ctx context.Context, name string) (*role.Role, error) {
	if m.GetRoleByNameFunc != nil {
		return m.GetRoleByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockRoleUseCase) UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*role.Role, error) {
	if m.UpdateRoleFunc != nil {
		return m.UpdateRoleFunc(ctx, roleID, displayName, description)
	}
	return nil, nil
}

func (m *MockRoleUseCase) DeleteRole(ctx context.Context, roleID uuidv7.UUID) error {
	if m.DeleteRoleFunc != nil {
		return m.DeleteRoleFunc(ctx, roleID)
	}
	return nil
}

func (m *MockRoleUseCase) ListRoles(ctx context.Context, limit, offset int) ([]*role.Role, int, error) {
	if m.ListRolesFunc != nil {
		return m.ListRolesFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockRoleUseCase) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*role.Role, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID)
	}
	return nil, nil
}

func fakeRole() *role.Role {
	r, _ := role.NewRole("test-role", "Test Role", "Test Description")
	return r
}

func TestRoleHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		CreateRoleFunc: func(ctx context.Context, name, displayName, description string) (*role.Role, error) {
			return fakeRole(), nil
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.POST("/roles", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/roles", map[string]any{
		"name":         "test-role",
		"display_name": "Test Role",
		"description":  "Test Description",
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestRoleHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()
	mockUC := &MockRoleUseCase{}
	handler := roleHTTP.NewRoleHandler(mockUC)
	router.POST("/roles", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/roles", map[string]any{
		// Missing required fields
	})

	smoke.AssertErrorResponse(t, resp, 400, "BAD_REQUEST")
}

func TestRoleHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		GetRoleFunc: func(ctx context.Context, roleID uuidv7.UUID) (*role.Role, error) {
			return fakeRole(), nil
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.GET("/roles/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/roles/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestRoleHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		GetRoleFunc: func(ctx context.Context, roleID uuidv7.UUID) (*role.Role, error) {
			return nil, role.ErrRoleNotFound
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.GET("/roles/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/roles/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}

func TestRoleHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		ListRolesFunc: func(ctx context.Context, limit, offset int) ([]*role.Role, int, error) {
			return []*role.Role{fakeRole()}, 1, nil
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.GET("/roles", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/roles?limit=10&offset=0", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestRoleHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		ListRolesFunc: func(ctx context.Context, limit, offset int) ([]*role.Role, int, error) {
			return []*role.Role{}, 0, nil
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.GET("/roles", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/roles?limit=10&offset=0", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestRoleHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		UpdateRoleFunc: func(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*role.Role, error) {
			return fakeRole(), nil
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.PUT("/roles/:id", handler.Update)

	resp := smoke.MakeRequest(t, router, "PUT", "/roles/"+smoke.FakeUUID(), map[string]any{
		"display_name": "Updated Role",
		"description":  "Updated Description",
	})

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestRoleHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		DeleteRoleFunc: func(ctx context.Context, roleID uuidv7.UUID) error {
			return nil
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.DELETE("/roles/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/roles/"+smoke.FakeUUID(), nil)

	assert.Equal(t, 204, resp.Code)
}

func TestRoleHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockRoleUseCase{
		DeleteRoleFunc: func(ctx context.Context, roleID uuidv7.UUID) error {
			return role.ErrRoleNotFound
		},
	}

	handler := roleHTTP.NewRoleHandler(mockUC)
	router.DELETE("/roles/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/roles/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}
