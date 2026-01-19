package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	permissionerrors "github.com/basilex/promenade/internal/contexts/identity/permission"
	"github.com/basilex/promenade/internal/contexts/identity/permission/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock repository for testing
type mockPermissionRepository struct {
	permissions  map[string]*aggregate.Permission
	existsByName bool
	createErr    error
	getErr       error
	updateErr    error
	deleteErr    error
	listErr      error
	getRoleErr   error
}

func newMockPermissionRepository() *mockPermissionRepository {
	return &mockPermissionRepository{
		permissions: make(map[string]*aggregate.Permission),
	}
}

func (m *mockPermissionRepository) Create(ctx context.Context, perm *aggregate.Permission) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.permissions[perm.ID.String()] = perm
	return nil
}

func (m *mockPermissionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Permission, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	perm, ok := m.permissions[id.String()]
	if !ok {
		return nil, permissionerrors.ErrPermissionNotFound
	}
	return perm, nil
}

func (m *mockPermissionRepository) GetByName(ctx context.Context, name string) (*aggregate.Permission, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, perm := range m.permissions {
		if perm.Name == name {
			return perm, nil
		}
	}
	return nil, permissionerrors.ErrPermissionNotFound
}

func (m *mockPermissionRepository) Update(ctx context.Context, perm *aggregate.Permission) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.permissions[perm.ID.String()] = perm
	return nil
}

func (m *mockPermissionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.permissions, id.String())
	return nil
}

func (m *mockPermissionRepository) ListPermissions(ctx context.Context, limit, offset int) ([]*aggregate.Permission, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	perms := make([]*aggregate.Permission, 0, len(m.permissions))
	for _, perm := range m.permissions {
		perms = append(perms, perm)
	}
	return perms, len(perms), nil
}

func (m *mockPermissionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	return m.existsByName, nil
}

func (m *mockPermissionRepository) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*aggregate.Permission, error) {
	if m.getRoleErr != nil {
		return nil, m.getRoleErr
	}
	return []*aggregate.Permission{}, nil
}

func (m *mockPermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	return nil
}

func (m *mockPermissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	return nil
}

func TestUseCase_CreatePermission(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		perm, err := uc.CreatePermission(ctx, "users", "create", "Create users")

		require.NoError(t, err)
		assert.NotNil(t, perm)
		assert.Equal(t, "users", perm.Resource)
		assert.Equal(t, "create", perm.Action)
		assert.Equal(t, "Create users", perm.Description)
		assert.Equal(t, "users:create", perm.Name)
	})

	t.Run("resource required", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		_, err := uc.CreatePermission(ctx, "", "create", "Description")

		assert.ErrorIs(t, err, permissionerrors.ErrResourceRequired)
	})

	t.Run("action required", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		_, err := uc.CreatePermission(ctx, "users", "", "Description")

		assert.ErrorIs(t, err, permissionerrors.ErrActionRequired)
	})

	t.Run("duplicate name", func(t *testing.T) {
		repo := newMockPermissionRepository()
		repo.existsByName = true
		uc := NewPermissionUseCase(repo)

		_, err := uc.CreatePermission(ctx, "users", "create", "Description")

		assert.ErrorIs(t, err, permissionerrors.ErrPermissionNameExists)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := newMockPermissionRepository()
		repo.createErr = errors.New("database error")
		uc := NewPermissionUseCase(repo)

		_, err := uc.CreatePermission(ctx, "users", "create", "Description")

		assert.Error(t, err)
	})
}

func TestUseCase_GetPermission(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		// Create permission first
		created, _ := uc.CreatePermission(ctx, "users", "read", "Read users")

		// Retrieve it
		found, err := uc.GetPermission(ctx, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "users", found.Resource)
		assert.Equal(t, "read", found.Action)
	})

	t.Run("permission not found", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		_, err := uc.GetPermission(ctx, uuidv7.New())

		assert.ErrorIs(t, err, permissionerrors.ErrPermissionNotFound)
	})
}

func TestUseCase_GetPermissionByName(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		// Create permission first
		_, _ = uc.CreatePermission(ctx, "users", "delete", "Delete users")

		// Retrieve by name
		found, err := uc.GetPermissionByName(ctx, "users:delete")

		require.NoError(t, err)
		assert.Equal(t, "users:delete", found.Name)
	})

	t.Run("name required", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		_, err := uc.GetPermissionByName(ctx, "")

		assert.ErrorIs(t, err, permissionerrors.ErrResourceRequired)
	})

	t.Run("permission not found", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		_, err := uc.GetPermissionByName(ctx, "nonexistent:action")

		assert.ErrorIs(t, err, permissionerrors.ErrPermissionNotFound)
	})
}

func TestUseCase_UpdatePermission(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		// Create permission first
		created, _ := uc.CreatePermission(ctx, "users", "update", "Old description")

		// Update it
		updated, err := uc.UpdatePermission(ctx, created.ID, "New description")

		require.NoError(t, err)
		assert.Equal(t, "New description", updated.Description)
	})

	t.Run("permission not found", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		_, err := uc.UpdatePermission(ctx, uuidv7.New(), "New description")

		assert.ErrorIs(t, err, permissionerrors.ErrPermissionNotFound)
	})
}

func TestUseCase_DeletePermission(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		// Create permission first
		created, _ := uc.CreatePermission(ctx, "users", "archive", "Archive users")

		// Delete it
		err := uc.DeletePermission(ctx, created.ID)

		assert.NoError(t, err)
	})

	t.Run("permission not found", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		err := uc.DeletePermission(ctx, uuidv7.New())

		assert.ErrorIs(t, err, permissionerrors.ErrPermissionNotFound)
	})
}

func TestUseCase_ListPermissions(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		// Create some permissions
		_, _ = uc.CreatePermission(ctx, "users", "create", "Create users")
		_, _ = uc.CreatePermission(ctx, "users", "read", "Read users")

		// List them
		perms, total, err := uc.ListPermissions(ctx, 20, 0)

		require.NoError(t, err)
		assert.Len(t, perms, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("default limit", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		perms, total, err := uc.ListPermissions(ctx, 0, 0)

		require.NoError(t, err)
		assert.NotNil(t, perms)
		assert.Equal(t, 0, total)
	})
}

func TestUseCase_GetRolePermissions(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := newMockPermissionRepository()
		uc := NewPermissionUseCase(repo)

		roleID := uuidv7.New()
		perms, err := uc.GetRolePermissions(ctx, roleID)

		require.NoError(t, err)
		assert.NotNil(t, perms)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := newMockPermissionRepository()
		repo.getRoleErr = errors.New("database error")
		uc := NewPermissionUseCase(repo)

		_, err := uc.GetRolePermissions(ctx, uuidv7.New())

		assert.Error(t, err)
	})
}
