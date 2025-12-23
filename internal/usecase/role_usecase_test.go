package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock RoleRepository
type mockRoleRepository struct {
	mock.Mock
}

func (m *mockRoleRepository) Create(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Role), args.Error(1)
}

func (m *mockRoleRepository) List(ctx context.Context) ([]*entity.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *mockRoleRepository) Update(ctx context.Context, role *entity.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *mockRoleRepository) GetPermissions(ctx context.Context, roleID uuidv7.UUID) ([]*entity.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

func (m *mockRoleRepository) SyncPermissions(ctx context.Context, roleID uuidv7.UUID, permissionIDs []uuidv7.UUID) error {
	args := m.Called(ctx, roleID, permissionIDs)
	return args.Error(0)
}

func (m *mockRoleRepository) AssignToUser(ctx context.Context, userRole *entity.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *mockRoleRepository) RemoveFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *mockRoleRepository) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *mockRoleRepository) GetUserActiveRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *mockRoleRepository) GetUsersWithRole(ctx context.Context, roleID uuidv7.UUID) ([]uuidv7.UUID, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuidv7.UUID), args.Error(1)
}

func (m *mockRoleRepository) CreateMany(ctx context.Context, roles []*entity.Role) error {
	args := m.Called(ctx, roles)
	return args.Error(0)
}

// Test helpers
func setupRoleUseCase(t *testing.T) (*roleUseCase, *mockRoleRepository, *mockPermissionRepository) {
	roleRepo := new(mockRoleRepository)
	permRepo := new(mockPermissionRepository)
	uc := NewRoleUseCase(roleRepo, permRepo).(*roleUseCase)
	return uc, roleRepo, permRepo
}

func createTestRole(name, displayName string, isSystem bool) *entity.Role {
	desc := "Test role"
	return &entity.Role{
		ID:          uuidv7.New(),
		Name:        name,
		DisplayName: displayName,
		Description: &desc,
		IsSystem:    isSystem,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Tests for CreateRole
func TestRoleUseCase_CreateRole_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	name := "editor"
	displayName := "Editor"
	description := "Content editor role"

	roleRepo.On("GetByName", ctx, name).Return(nil, entity.ErrNotFound)
	roleRepo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(nil)

	role, err := uc.CreateRole(ctx, name, displayName, description)

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, name, role.Name)
	assert.Equal(t, displayName, role.DisplayName)
	assert.False(t, role.IsSystem)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_CreateRole_AlreadyExists(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	name := "editor"
	existingRole := createTestRole(name, "Editor", false)

	roleRepo.On("GetByName", ctx, name).Return(existingRole, nil)

	role, err := uc.CreateRole(ctx, name, "Editor", "description")

	assert.Error(t, err)
	assert.Equal(t, ErrRoleAlreadyExists, err)
	assert.Nil(t, role)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_CreateRole_InvalidName(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleRepo.On("GetByName", ctx, "").Return(nil, entity.ErrNotFound)

	role, err := uc.CreateRole(ctx, "", "Display", "description")

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "validation failed")
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_CreateRole_RepositoryError(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	name := "editor"
	roleRepo.On("GetByName", ctx, name).Return(nil, entity.ErrNotFound)
	roleRepo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).
		Return(errors.New("database error"))

	role, err := uc.CreateRole(ctx, name, "Editor", "description")

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "failed to create role")
	roleRepo.AssertExpectations(t)
}

// Tests for GetRole
func TestRoleUseCase_GetRole_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	expectedRole := createTestRole("editor", "Editor", false)
	permissions := []*entity.Permission{}

	roleRepo.On("GetByID", ctx, expectedRole.ID).Return(expectedRole, nil)
	roleRepo.On("GetPermissions", ctx, expectedRole.ID).Return(permissions, nil)

	role, err := uc.GetRole(ctx, expectedRole.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRole.ID, role.ID)
	assert.Equal(t, expectedRole.Name, role.Name)
	assert.NotNil(t, role.Permissions)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_GetRole_NotFound(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	roleRepo.On("GetByID", ctx, roleID).Return(nil, entity.ErrNotFound)

	role, err := uc.GetRole(ctx, roleID)

	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, role)
	roleRepo.AssertExpectations(t)
}

// Tests for GetRoleByName
func TestRoleUseCase_GetRoleByName_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	expectedRole := createTestRole("admin", "Administrator", true)
	permissions := []*entity.Permission{}

	roleRepo.On("GetByName", ctx, "admin").Return(expectedRole, nil)
	roleRepo.On("GetPermissions", ctx, expectedRole.ID).Return(permissions, nil)

	role, err := uc.GetRoleByName(ctx, "admin")

	assert.NoError(t, err)
	assert.Equal(t, "admin", role.Name)
	assert.True(t, role.IsSystem)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_GetRoleByName_NotFound(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleRepo.On("GetByName", ctx, "nonexistent").Return(nil, entity.ErrNotFound)

	role, err := uc.GetRoleByName(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, role)
	roleRepo.AssertExpectations(t)
}

// Tests for ListRoles
func TestRoleUseCase_ListRoles_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	role1 := createTestRole("admin", "Administrator", true)
	role2 := createTestRole("user", "User", true)
	expectedRoles := []*entity.Role{role1, role2}
	permissions := []*entity.Permission{}

	roleRepo.On("List", ctx).Return(expectedRoles, nil)
	roleRepo.On("GetPermissions", ctx, role1.ID).Return(permissions, nil)
	roleRepo.On("GetPermissions", ctx, role2.ID).Return(permissions, nil)

	roles, err := uc.ListRoles(ctx)

	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "admin", roles[0].Name)
	assert.Equal(t, "user", roles[1].Name)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_ListRoles_Empty(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleRepo.On("List", ctx).Return([]*entity.Role{}, nil)

	roles, err := uc.ListRoles(ctx)

	assert.NoError(t, err)
	assert.Empty(t, roles)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_ListRoles_RepositoryError(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleRepo.On("List", ctx).Return(nil, errors.New("database error"))

	roles, err := uc.ListRoles(ctx)

	assert.Error(t, err)
	assert.Nil(t, roles)
	assert.Contains(t, err.Error(), "failed to list roles")
	roleRepo.AssertExpectations(t)
}

// Tests for UpdateRole
func TestRoleUseCase_UpdateRole_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	role := createTestRole("editor", "Editor", false)
	newDisplayName := "Content Editor"
	newDescription := "Updated description"

	roleRepo.On("GetByID", ctx, role.ID).Return(role, nil)
	roleRepo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).Return(nil)

	updated, err := uc.UpdateRole(ctx, role.ID, newDisplayName, newDescription)

	assert.NoError(t, err)
	assert.Equal(t, newDisplayName, updated.DisplayName)
	assert.Equal(t, newDescription, *updated.Description)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_UpdateRole_NotFound(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	roleRepo.On("GetByID", ctx, roleID).Return(nil, entity.ErrNotFound)

	updated, err := uc.UpdateRole(ctx, roleID, "New Name", "description")

	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, updated)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_UpdateRole_SystemRole(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	systemRole := createTestRole("admin", "Administrator", true)
	roleRepo.On("GetByID", ctx, systemRole.ID).Return(systemRole, nil)

	updated, err := uc.UpdateRole(ctx, systemRole.ID, "New Name", "description")

	assert.Error(t, err)
	assert.Equal(t, ErrCannotDeleteSystemRole, err)
	assert.Nil(t, updated)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_UpdateRole_RepositoryError(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	role := createTestRole("editor", "Editor", false)
	roleRepo.On("GetByID", ctx, role.ID).Return(role, nil)
	roleRepo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).
		Return(errors.New("database error"))

	updated, err := uc.UpdateRole(ctx, role.ID, "New Name", "description")

	assert.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "failed to update role")
	roleRepo.AssertExpectations(t)
}

// Tests for DeleteRole
func TestRoleUseCase_DeleteRole_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	role := createTestRole("editor", "Editor", false)
	roleRepo.On("GetByID", ctx, role.ID).Return(role, nil)
	roleRepo.On("Delete", ctx, role.ID).Return(nil)

	err := uc.DeleteRole(ctx, role.ID)

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_DeleteRole_NotFound(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	roleRepo.On("GetByID", ctx, roleID).Return(nil, entity.ErrNotFound)

	err := uc.DeleteRole(ctx, roleID)

	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_DeleteRole_SystemRole(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	systemRole := createTestRole("superadmin", "Superadmin", true)
	roleRepo.On("GetByID", ctx, systemRole.ID).Return(systemRole, nil)

	err := uc.DeleteRole(ctx, systemRole.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrCannotDeleteSystemRole, err)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_DeleteRole_RepositoryError(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	role := createTestRole("editor", "Editor", false)
	roleRepo.On("GetByID", ctx, role.ID).Return(role, nil)
	roleRepo.On("Delete", ctx, role.ID).Return(errors.New("database error"))

	err := uc.DeleteRole(ctx, role.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete role")
	roleRepo.AssertExpectations(t)
}

// Tests for AddPermissionToRole
func TestRoleUseCase_AddPermissionToRole_Success(t *testing.T) {
	uc, roleRepo, permRepo := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	permissionID := uuidv7.New()
	role := createTestRole("editor", "Editor", false)
	role.ID = roleID
	permission, _ := entity.NewPermission("posts:create")
	permission.ID = permissionID

	roleRepo.On("GetByID", ctx, roleID).Return(role, nil)
	permRepo.On("GetByID", ctx, permissionID).Return(permission, nil)
	roleRepo.On("AddPermission", ctx, roleID, permissionID).Return(nil)

	err := uc.AddPermissionToRole(ctx, roleID, permissionID)

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

func TestRoleUseCase_AddPermissionToRole_RoleNotFound(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	permissionID := uuidv7.New()
	roleRepo.On("GetByID", ctx, roleID).Return(nil, entity.ErrNotFound)

	err := uc.AddPermissionToRole(ctx, roleID, permissionID)

	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_AddPermissionToRole_PermissionNotFound(t *testing.T) {
	uc, roleRepo, permRepo := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	permissionID := uuidv7.New()
	role := createTestRole("editor", "Editor", false)
	role.ID = roleID

	roleRepo.On("GetByID", ctx, roleID).Return(role, nil)
	permRepo.On("GetByID", ctx, permissionID).Return(nil, entity.ErrNotFound)

	err := uc.AddPermissionToRole(ctx, roleID, permissionID)

	assert.Error(t, err)
	assert.Equal(t, ErrPermissionNotFound, err)
	roleRepo.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

// Tests for RemovePermissionFromRole
func TestRoleUseCase_RemovePermissionFromRole_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	permissionID := uuidv7.New()

	roleRepo.On("RemovePermission", ctx, roleID, permissionID).Return(nil)

	err := uc.RemovePermissionFromRole(ctx, roleID, permissionID)

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_RemovePermissionFromRole_RepositoryError(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	permissionID := uuidv7.New()

	roleRepo.On("RemovePermission", ctx, roleID, permissionID).
		Return(errors.New("database error"))

	err := uc.RemovePermissionFromRole(ctx, roleID, permissionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove permission")
	roleRepo.AssertExpectations(t)
}

// Tests for GetRolePermissions
func TestRoleUseCase_GetRolePermissions_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	perm1, _ := entity.NewPermission("posts:create")
	perm2, _ := entity.NewPermission("posts:read")
	expectedPerms := []*entity.Permission{perm1, perm2}

	roleRepo.On("GetPermissions", ctx, roleID).Return(expectedPerms, nil)

	permissions, err := uc.GetRolePermissions(ctx, roleID)

	assert.NoError(t, err)
	assert.Len(t, permissions, 2)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_GetRolePermissions_Empty(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	roleID := uuidv7.New()
	roleRepo.On("GetPermissions", ctx, roleID).Return([]*entity.Permission{}, nil)

	permissions, err := uc.GetRolePermissions(ctx, roleID)

	assert.NoError(t, err)
	assert.Empty(t, permissions)
	roleRepo.AssertExpectations(t)
}

// Tests for AssignRoleToUser
func TestRoleUseCase_AssignRoleToUser_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	roleID := uuidv7.New()
	assignedBy := uuidv7.New()
	role := createTestRole("editor", "Editor", false)
	role.ID = roleID

	roleRepo.On("GetByID", ctx, roleID).Return(role, nil)
	roleRepo.On("AssignToUser", ctx, mock.AnythingOfType("*entity.UserRole")).Return(nil)

	err := uc.AssignRoleToUser(ctx, userID, roleID, assignedBy, nil)

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_AssignRoleToUser_RoleNotFound(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	roleID := uuidv7.New()
	assignedBy := uuidv7.New()

	roleRepo.On("GetByID", ctx, roleID).Return(nil, entity.ErrNotFound)

	err := uc.AssignRoleToUser(ctx, userID, roleID, assignedBy, nil)

	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	roleRepo.AssertExpectations(t)
}

// Tests for GetUserRoles
func TestRoleUseCase_GetUserRoles_Success(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	role1 := createTestRole("admin", "Administrator", true)
	role2 := createTestRole("editor", "Editor", false)
	expectedRoles := []*entity.Role{role1, role2}
	permissions := []*entity.Permission{}

	roleRepo.On("GetUserRoles", ctx, userID).Return(expectedRoles, nil)
	roleRepo.On("GetPermissions", ctx, role1.ID).Return(permissions, nil)
	roleRepo.On("GetPermissions", ctx, role2.ID).Return(permissions, nil)

	roles, err := uc.GetUserRoles(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	roleRepo.AssertExpectations(t)
}

func TestRoleUseCase_GetUserRoles_Empty(t *testing.T) {
	uc, roleRepo, _ := setupRoleUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*entity.Role{}, nil)

	roles, err := uc.GetUserRoles(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, roles)
	roleRepo.AssertExpectations(t)
}
