package permission_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	permissionHTTP "github.com/basilex/promenade/internal/contexts/identity/permission/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/stretchr/testify/assert"
)

// MockPermissionUseCase mocks permission.IUseCase
type MockPermissionUseCase struct {
	CreatePermissionFunc     func(ctx context.Context, resource, action, description string) (*permission.Permission, error)
	GetPermissionFunc        func(ctx context.Context, permissionID uuidv7.UUID) (*permission.Permission, error)
	GetPermissionByNameFunc  func(ctx context.Context, name string) (*permission.Permission, error)
	UpdatePermissionFunc     func(ctx context.Context, permissionID uuidv7.UUID, description string) (*permission.Permission, error)
	DeletePermissionFunc     func(ctx context.Context, permissionID uuidv7.UUID) error
	ListPermissionsFunc      func(ctx context.Context, limit, offset int) ([]*permission.Permission, int, error)
	GetRolePermissionsFunc   func(ctx context.Context, roleID uuidv7.UUID) ([]*permission.Permission, error)
}

func (m *MockPermissionUseCase) CreatePermission(ctx context.Context, resource, action, description string) (*permission.Permission, error) {
	if m.CreatePermissionFunc != nil {
		return m.CreatePermissionFunc(ctx, resource, action, description)
	}
	return nil, nil
}

func (m *MockPermissionUseCase) GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*permission.Permission, error) {
	if m.GetPermissionFunc != nil {
		return m.GetPermissionFunc(ctx, permissionID)
	}
	return nil, nil
}

func (m *MockPermissionUseCase) GetPermissionByName(ctx context.Context, name string) (*permission.Permission, error) {
	if m.GetPermissionByNameFunc != nil {
		return m.GetPermissionByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockPermissionUseCase) UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*permission.Permission, error) {
	if m.UpdatePermissionFunc != nil {
		return m.UpdatePermissionFunc(ctx, permissionID, description)
	}
	return nil, nil
}

func (m *MockPermissionUseCase) DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error {
	if m.DeletePermissionFunc != nil {
		return m.DeletePermissionFunc(ctx, permissionID)
	}
	return nil
}

func (m *MockPermissionUseCase) ListPermissions(ctx context.Context, limit, offset int) ([]*permission.Permission, int, error) {
	if m.ListPermissionsFunc != nil {
		return m.ListPermissionsFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockPermissionUseCase) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*permission.Permission, error) {
	if m.GetRolePermissionsFunc != nil {
		return m.GetRolePermissionsFunc(ctx, roleID)
	}
	return nil, nil
}

func fakePermission() *permission.Permission {
	p, _ := permission.NewPermission("users", "read", "Read users")
	return p
}

func TestPermissionHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		CreatePermissionFunc: func(ctx context.Context, resource, action, description string) (*permission.Permission, error) {
			return fakePermission(), nil
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.POST("/permissions", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/permissions", map[string]any{
		"resource":    "users",
		"action":      "read",
		"description": "Read users",
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestPermissionHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()
	mockUC := &MockPermissionUseCase{}
	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.POST("/permissions", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/permissions", map[string]any{
		// Missing required fields
	})

	smoke.AssertErrorResponse(t, resp, 400, "BAD_REQUEST")
}

func TestPermissionHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		GetPermissionFunc: func(ctx context.Context, permissionID uuidv7.UUID) (*permission.Permission, error) {
			return fakePermission(), nil
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.GET("/permissions/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/permissions/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestPermissionHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		GetPermissionFunc: func(ctx context.Context, permissionID uuidv7.UUID) (*permission.Permission, error) {
			return nil, permission.ErrPermissionNotFound
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.GET("/permissions/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/permissions/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}

func TestPermissionHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		ListPermissionsFunc: func(ctx context.Context, limit, offset int) ([]*permission.Permission, int, error) {
			return []*permission.Permission{fakePermission()}, 1, nil
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.GET("/permissions", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/permissions?limit=10&offset=0", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestPermissionHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		ListPermissionsFunc: func(ctx context.Context, limit, offset int) ([]*permission.Permission, int, error) {
			return []*permission.Permission{}, 0, nil
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.GET("/permissions", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/permissions?limit=10&offset=0", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestPermissionHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		UpdatePermissionFunc: func(ctx context.Context, permissionID uuidv7.UUID, description string) (*permission.Permission, error) {
			return fakePermission(), nil
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.PUT("/permissions/:id", handler.Update)

	resp := smoke.MakeRequest(t, router, "PUT", "/permissions/"+smoke.FakeUUID(), map[string]any{
		"description": "Updated description",
	})

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestPermissionHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		DeletePermissionFunc: func(ctx context.Context, permissionID uuidv7.UUID) error {
			return nil
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.DELETE("/permissions/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/permissions/"+smoke.FakeUUID(), nil)

	assert.Equal(t, 204, resp.Code)
}

func TestPermissionHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPermissionUseCase{
		DeletePermissionFunc: func(ctx context.Context, permissionID uuidv7.UUID) error {
			return permission.ErrPermissionNotFound
		},
	}

	handler := permissionHTTP.NewPermissionHandler(mockUC)
	router.DELETE("/permissions/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/permissions/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}
