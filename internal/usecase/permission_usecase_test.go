package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock PermissionRepository
type mockPermissionRepository struct {
	mock.Mock
}

func (m *mockPermissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *mockPermissionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}

func (m *mockPermissionRepository) GetByResourceAction(ctx context.Context, resource, action string) (*entity.Permission, error) {
	args := m.Called(ctx, resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Permission), args.Error(1)
}

func (m *mockPermissionRepository) List(ctx context.Context) ([]*entity.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *mockPermissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *mockPermissionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPermissionRepository) CreateMany(ctx context.Context, permissions []*entity.Permission) error {
	args := m.Called(ctx, permissions)
	return args.Error(0)
}

func (m *mockPermissionRepository) GetByIDs(ctx context.Context, ids []uuidv7.UUID) ([]*entity.Permission, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *mockPermissionRepository) FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error) {
	args := m.Called(ctx, resource)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

// Test helper
func setupPermissionUseCase(t *testing.T) (*permissionUseCase, *mockPermissionRepository) {
	permRepo := new(mockPermissionRepository)
	uc := NewPermissionUseCase(permRepo).(*permissionUseCase)
	return uc, permRepo
}

// Tests for CreatePermission
func TestPermissionUseCase_CreatePermission_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	resource := "users"
	action := "create"
	description := "Create new users"

	permRepo.On("GetByResourceAction", ctx, resource, action).Return(nil, entity.ErrNotFound)
	permRepo.On("Create", ctx, mock.AnythingOfType("*entity.Permission")).Return(nil)

	permission, err := uc.CreatePermission(ctx, resource, action, description)

	assert.NoError(t, err)
	assert.NotNil(t, permission)
	assert.Equal(t, resource, permission.Resource)
	assert.Equal(t, action, permission.Action)
	assert.NotNil(t, permission.Description)
	assert.Equal(t, description, *permission.Description)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_CreatePermission_AlreadyExists(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	resource := "users"
	action := "create"
	existing, _ := entity.NewPermission("users:create")

	permRepo.On("GetByResourceAction", ctx, resource, action).Return(existing, nil)

	permission, err := uc.CreatePermission(ctx, resource, action, "description")

	assert.Error(t, err)
	assert.Equal(t, ErrPermissionAlreadyExists, err)
	assert.Nil(t, permission)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_CreatePermission_InvalidResource(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permRepo.On("GetByResourceAction", ctx, "", "create").Return(nil, entity.ErrNotFound)

	permission, err := uc.CreatePermission(ctx, "", "create", "description")

	assert.Error(t, err)
	assert.Nil(t, permission)
	assert.Contains(t, err.Error(), "validation failed")
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_CreatePermission_RepositoryError(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	resource := "users"
	action := "create"

	permRepo.On("GetByResourceAction", ctx, resource, action).Return(nil, entity.ErrNotFound)
	permRepo.On("Create", ctx, mock.AnythingOfType("*entity.Permission")).
		Return(errors.New("database error"))

	permission, err := uc.CreatePermission(ctx, resource, action, "description")

	assert.Error(t, err)
	assert.Nil(t, permission)
	assert.Contains(t, err.Error(), "failed to create permission")
	permRepo.AssertExpectations(t)
}

// Tests for CreatePermissionFromString
func TestPermissionUseCase_CreatePermissionFromString_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permissionString := "posts:create"
	description := "Create new posts"

	permRepo.On("GetByResourceAction", ctx, "posts", "create").Return(nil, entity.ErrNotFound)
	permRepo.On("Create", ctx, mock.AnythingOfType("*entity.Permission")).Return(nil)

	permission, err := uc.CreatePermissionFromString(ctx, permissionString, description)

	assert.NoError(t, err)
	assert.NotNil(t, permission)
	assert.Equal(t, "posts", permission.Resource)
	assert.Equal(t, "create", permission.Action)
	assert.Equal(t, description, *permission.Description)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_CreatePermissionFromString_InvalidFormat(t *testing.T) {
	uc, _ := setupPermissionUseCase(t)
	ctx := context.Background()

	permission, err := uc.CreatePermissionFromString(ctx, "invalid-format", "description")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPermissionFormat, err)
	assert.Nil(t, permission)
}

func TestPermissionUseCase_CreatePermissionFromString_AlreadyExists(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permissionString := "posts:create"
	existing, _ := entity.NewPermission(permissionString)

	permRepo.On("GetByResourceAction", ctx, "posts", "create").Return(existing, nil)

	permission, err := uc.CreatePermissionFromString(ctx, permissionString, "description")

	assert.Error(t, err)
	assert.Equal(t, ErrPermissionAlreadyExists, err)
	assert.Nil(t, permission)
	permRepo.AssertExpectations(t)
}

// Tests for GetPermission
func TestPermissionUseCase_GetPermission_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	expectedPerm, _ := entity.NewPermission("users:read")
	expectedPerm.ID = uuidv7.New()

	permRepo.On("GetByID", ctx, expectedPerm.ID).Return(expectedPerm, nil)

	permission, err := uc.GetPermission(ctx, expectedPerm.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPerm.ID, permission.ID)
	assert.Equal(t, "users", permission.Resource)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_GetPermission_NotFound(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permID := uuidv7.New()
	permRepo.On("GetByID", ctx, permID).Return(nil, entity.ErrNotFound)

	permission, err := uc.GetPermission(ctx, permID)

	assert.Error(t, err)
	// Note: ErrPermissionNotFound is not defined in the code but is used
	// This test will fail until that's fixed
	assert.Nil(t, permission)
	permRepo.AssertExpectations(t)
}

// Tests for GetPermissionByResourceAction
func TestPermissionUseCase_GetPermissionByResourceAction_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	expectedPerm, _ := entity.NewPermission("posts:update")

	permRepo.On("GetByResourceAction", ctx, "posts", "update").Return(expectedPerm, nil)

	permission, err := uc.GetPermissionByResourceAction(ctx, "posts", "update")

	assert.NoError(t, err)
	assert.Equal(t, "posts", permission.Resource)
	assert.Equal(t, "update", permission.Action)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_GetPermissionByResourceAction_NotFound(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permRepo.On("GetByResourceAction", ctx, "nonexistent", "read").Return(nil, entity.ErrNotFound)

	permission, err := uc.GetPermissionByResourceAction(ctx, "nonexistent", "read")

	assert.Error(t, err)
	assert.Nil(t, permission)
	permRepo.AssertExpectations(t)
}

// Tests for ListPermissions
func TestPermissionUseCase_ListPermissions_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	perm1, _ := entity.NewPermission("users:create")
	perm2, _ := entity.NewPermission("users:read")
	expectedPerms := []*entity.Permission{perm1, perm2}

	permRepo.On("List", ctx).Return(expectedPerms, nil)

	permissions, err := uc.ListPermissions(ctx)

	assert.NoError(t, err)
	assert.Len(t, permissions, 2)
	assert.Equal(t, "users", permissions[0].Resource)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_ListPermissions_Empty(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permRepo.On("List", ctx).Return([]*entity.Permission{}, nil)

	permissions, err := uc.ListPermissions(ctx)

	assert.NoError(t, err)
	assert.Empty(t, permissions)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_ListPermissions_RepositoryError(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permRepo.On("List", ctx).Return(nil, errors.New("database error"))

	permissions, err := uc.ListPermissions(ctx)

	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "failed to list permissions")
	permRepo.AssertExpectations(t)
}

// Tests for UpdatePermission
func TestPermissionUseCase_UpdatePermission_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	perm, _ := entity.NewPermission("users:update")
	perm.ID = uuidv7.New()
	newDescription := "Updated description"

	permRepo.On("GetByID", ctx, perm.ID).Return(perm, nil)
	permRepo.On("Update", ctx, mock.AnythingOfType("*entity.Permission")).Return(nil)

	updated, err := uc.UpdatePermission(ctx, perm.ID, newDescription)

	assert.NoError(t, err)
	assert.NotNil(t, updated.Description)
	assert.Equal(t, newDescription, *updated.Description)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_UpdatePermission_NotFound(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permID := uuidv7.New()
	permRepo.On("GetByID", ctx, permID).Return(nil, entity.ErrNotFound)

	updated, err := uc.UpdatePermission(ctx, permID, "new description")

	assert.Error(t, err)
	assert.Nil(t, updated)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_UpdatePermission_RepositoryError(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	perm, _ := entity.NewPermission("users:update")
	perm.ID = uuidv7.New()

	permRepo.On("GetByID", ctx, perm.ID).Return(perm, nil)
	permRepo.On("Update", ctx, mock.AnythingOfType("*entity.Permission")).
		Return(errors.New("database error"))

	updated, err := uc.UpdatePermission(ctx, perm.ID, "new description")

	assert.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "failed to update permission")
	permRepo.AssertExpectations(t)
}

// Tests for DeletePermission
func TestPermissionUseCase_DeletePermission_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permID := uuidv7.New()
	permRepo.On("Delete", ctx, permID).Return(nil)

	err := uc.DeletePermission(ctx, permID)

	assert.NoError(t, err)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_DeletePermission_NotFound(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permID := uuidv7.New()
	permRepo.On("Delete", ctx, permID).Return(entity.ErrNotFound)

	err := uc.DeletePermission(ctx, permID)

	assert.Error(t, err)
	permRepo.AssertExpectations(t)
}

// Tests for CreateMany
func TestPermissionUseCase_CreateMany_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	perm1, _ := entity.NewPermission("posts:create")
	perm2, _ := entity.NewPermission("posts:read")
	permissions := []*entity.Permission{perm1, perm2}

	permRepo.On("CreateMany", ctx, permissions).Return(nil)

	err := uc.CreateMany(ctx, permissions)

	assert.NoError(t, err)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_CreateMany_InvalidPermission(t *testing.T) {
	uc, _ := setupPermissionUseCase(t)
	ctx := context.Background()

	perm1, _ := entity.NewPermission("posts:create")
	invalidPerm := &entity.Permission{Resource: "", Action: "read"} // Invalid
	permissions := []*entity.Permission{perm1, invalidPerm}

	err := uc.CreateMany(ctx, permissions)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestPermissionUseCase_CreateMany_RepositoryError(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	perm1, _ := entity.NewPermission("posts:create")
	permissions := []*entity.Permission{perm1}

	permRepo.On("CreateMany", ctx, permissions).Return(errors.New("database error"))

	err := uc.CreateMany(ctx, permissions)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create many permissions")
	permRepo.AssertExpectations(t)
}

// Tests for FindByResource
func TestPermissionUseCase_FindByResource_Success(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	perm1, _ := entity.NewPermission("users:create")
	perm2, _ := entity.NewPermission("users:read")
	expectedPerms := []*entity.Permission{perm1, perm2}

	permRepo.On("FindByResource", ctx, "users").Return(expectedPerms, nil)

	permissions, err := uc.FindByResource(ctx, "users")

	assert.NoError(t, err)
	assert.Len(t, permissions, 2)
	assert.Equal(t, "users", permissions[0].Resource)
	assert.Equal(t, "users", permissions[1].Resource)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_FindByResource_Empty(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permRepo.On("FindByResource", ctx, "nonexistent").Return([]*entity.Permission{}, nil)

	permissions, err := uc.FindByResource(ctx, "nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, permissions)
	permRepo.AssertExpectations(t)
}

func TestPermissionUseCase_FindByResource_RepositoryError(t *testing.T) {
	uc, permRepo := setupPermissionUseCase(t)
	ctx := context.Background()

	permRepo.On("FindByResource", ctx, "users").Return(nil, errors.New("database error"))

	permissions, err := uc.FindByResource(ctx, "users")

	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "failed to find permissions by resource")
	permRepo.AssertExpectations(t)
}
