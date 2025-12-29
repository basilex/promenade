package role

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock repository for testing
type mockRoleRepository struct {
	roles         map[string]*Role
	existsByName  bool
	createErr     error
	getErr        error
	updateErr     error
	deleteErr     error
	listErr       error
	getUserErr    error
}

func newMockRoleRepository() *mockRoleRepository {
	return &mockRoleRepository{
		roles: make(map[string]*Role),
	}
}

func (m *mockRoleRepository) Create(ctx context.Context, role *Role) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.roles[role.ID.String()] = role
	return nil
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Role, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	role, ok := m.roles[id.String()]
	if !ok {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*Role, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, role := range m.roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, ErrRoleNotFound
}

func (m *mockRoleRepository) Update(ctx context.Context, role *Role) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.roles[role.ID.String()] = role
	return nil
}

func (m *mockRoleRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.roles, id.String())
	return nil
}

func (m *mockRoleRepository) ListRoles(ctx context.Context, limit, offset int) ([]*Role, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	roles := make([]*Role, 0, len(m.roles))
	for _, role := range m.roles {
		roles = append(roles, role)
	}
	return roles, len(roles), nil
}

func (m *mockRoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	return m.existsByName, nil
}

func (m *mockRoleRepository) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*Role, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	return []*Role{}, nil
}

func (m *mockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuidv7.UUID, assignedBy *uuidv7.UUID) error {
	return nil
}

func (m *mockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error {
	return nil
}

func TestUseCase_CreateRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		role, err := uc.CreateRole(ctx, "manager", "Manager", "Manages team")

		require.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "manager", role.Name)
		assert.Equal(t, "Manager", role.DisplayName)
		assert.Equal(t, "Manages team", role.Description)
		assert.False(t, role.IsSystem)
	})

	t.Run("name required", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		_, err := uc.CreateRole(ctx, "", "Display", "Description")

		assert.ErrorIs(t, err, ErrRoleNameRequired)
	})

	t.Run("display name required", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		_, err := uc.CreateRole(ctx, "manager", "", "Description")

		assert.ErrorIs(t, err, ErrRoleDisplayRequired)
	})

	t.Run("duplicate name", func(t *testing.T) {
		repo := newMockRoleRepository()
		repo.existsByName = true
		uc := NewUseCase(repo)

		_, err := uc.CreateRole(ctx, "manager", "Manager", "Description")

		assert.ErrorIs(t, err, ErrRoleNameExists)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := newMockRoleRepository()
		repo.createErr = errors.New("database error")
		uc := NewUseCase(repo)

		_, err := uc.CreateRole(ctx, "manager", "Manager", "Description")

		assert.Error(t, err)
	})
}

func TestUseCase_GetRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		// Create role first
		created, _ := uc.CreateRole(ctx, "admin", "Administrator", "Admin role")

		// Retrieve it
		found, err := uc.GetRole(ctx, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "admin", found.Name)
	})

	t.Run("role not found", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		_, err := uc.GetRole(ctx, uuidv7.New())

		assert.ErrorIs(t, err, ErrRoleNotFound)
	})
}

func TestUseCase_GetRoleByName(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		// Create role first
		_, _ = uc.CreateRole(ctx, "admin", "Administrator", "Admin role")

		// Retrieve by name
		found, err := uc.GetRoleByName(ctx, "admin")

		require.NoError(t, err)
		assert.Equal(t, "admin", found.Name)
	})

	t.Run("name required", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		_, err := uc.GetRoleByName(ctx, "")

		assert.ErrorIs(t, err, ErrRoleNameRequired)
	})

	t.Run("role not found", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		_, err := uc.GetRoleByName(ctx, "nonexistent")

		assert.ErrorIs(t, err, ErrRoleNotFound)
	})
}

func TestUseCase_UpdateRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		// Create role first
		created, _ := uc.CreateRole(ctx, "manager", "Manager", "Old description")

		// Update it
		updated, err := uc.UpdateRole(ctx, created.ID, "Senior Manager", "New description")

		require.NoError(t, err)
		assert.Equal(t, "Senior Manager", updated.DisplayName)
		assert.Equal(t, "New description", updated.Description)
	})

	t.Run("display name required", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		created, _ := uc.CreateRole(ctx, "manager", "Manager", "Description")

		_, err := uc.UpdateRole(ctx, created.ID, "", "New description")

		assert.ErrorIs(t, err, ErrRoleDisplayRequired)
	})

	t.Run("role not found", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		_, err := uc.UpdateRole(ctx, uuidv7.New(), "Display", "Description")

		assert.ErrorIs(t, err, ErrRoleNotFound)
	})
}

func TestUseCase_DeleteRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		// Create role first
		created, _ := uc.CreateRole(ctx, "manager", "Manager", "Description")

		// Delete it
		err := uc.DeleteRole(ctx, created.ID)

		assert.NoError(t, err)
	})

	t.Run("cannot delete system role", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		// Create system role
		systemRole, _ := NewSystemRole("admin", "Administrator", "System admin role")
		_ = repo.Create(ctx, systemRole)

		// Try to delete
		err := uc.DeleteRole(ctx, systemRole.ID)

		assert.ErrorIs(t, err, ErrCannotDeleteSystem)
	})

	t.Run("role not found", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		err := uc.DeleteRole(ctx, uuidv7.New())

		assert.ErrorIs(t, err, ErrRoleNotFound)
	})
}

func TestUseCase_ListRoles(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		// Create some roles
		_, _ = uc.CreateRole(ctx, "admin", "Administrator", "Admin role")
		_, _ = uc.CreateRole(ctx, "user", "User", "Regular user")

		// List them
		roles, total, err := uc.ListRoles(ctx, 20, 0)

		require.NoError(t, err)
		assert.Len(t, roles, 2)
		assert.Equal(t, 2, total)
	})

	t.Run("default limit", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		roles, total, err := uc.ListRoles(ctx, 0, 0)

		require.NoError(t, err)
		assert.NotNil(t, roles)
		assert.Equal(t, 0, total)
	})
}

func TestUseCase_GetUserRoles(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := newMockRoleRepository()
		uc := NewUseCase(repo)

		userID := uuidv7.New()
		roles, err := uc.GetUserRoles(ctx, userID)

		require.NoError(t, err)
		assert.NotNil(t, roles)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := newMockRoleRepository()
		repo.getUserErr = errors.New("database error")
		uc := NewUseCase(repo)

		_, err := uc.GetUserRoles(ctx, uuidv7.New())

		assert.Error(t, err)
	})
}
