package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of IRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*User), args.Int(1), args.Error(2)
}

// MockRoleRepository is a mock implementation of role.IRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *role.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*role.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*role.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*role.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, r *role.Role) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepository) ListRoles(ctx context.Context) ([]*role.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*role.Role), args.Error(1)
}

func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuidv7.UUID, assignedBy *uuidv7.UUID) error {
	args := m.Called(ctx, userID, roleID, assignedBy)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*role.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*role.Role), args.Error(1)
}

// Test Register
func TestUseCase_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		
		// Mock role assignment
		userRole := &role.Role{ID: uuidv7.New(), Name: "user"}
		roleRepo.On("GetByName", ctx, "user").Return(userRole, nil)
		roleRepo.On("AssignRoleToUser", ctx, mock.Anything, userRole.ID, mock.Anything).Return(nil)
		
		// Mock GetByID for reloading user with roles after assignment
		// Create a user with roles that will be returned
		userWithRoles, _ := NewUser("test@example.com", "password123")
		userWithRoles.Roles = []string{"user"}
		repo.On("GetByID", ctx, mock.Anything).Return(userWithRoles, nil).Maybe()

		user, err := uc.Register(ctx, "test@example.com", "Test User", "password123")

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email.Value())
		assert.Equal(t, UserStatusActive, user.Status)
		assert.False(t, user.EmailVerified)
		repo.AssertExpectations(t)
	})

	t.Run("error - email already exists", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ExistsByEmail", ctx, "test@example.com").Return(true, nil)

		_, err := uc.Register(ctx, "test@example.com", "Test User", "password123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmailAlreadyExists)
		repo.AssertExpectations(t)
	})

	t.Run("error - invalid email", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ExistsByEmail", ctx, "invalid-email").Return(false, nil)

		_, err := uc.Register(ctx, "invalid-email", "Test User", "password123")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("error - weak password", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)

		_, err := uc.Register(ctx, "test@example.com", "Test User", "weak")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository failure on check", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ExistsByEmail", ctx, "test@example.com").Return(false, errors.New("db error"))

		_, err := uc.Register(ctx, "test@example.com", "Test User", "password123")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository failure on create", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.Register(ctx, "test@example.com", "Test User", "password123")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

// Test Authenticate
func TestUseCase_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.Activate()

		repo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		authUser, err := uc.Authenticate(ctx, "test@example.com", "password123")

		require.NoError(t, err)
		assert.NotNil(t, authUser)
		assert.Equal(t, user.ID, authUser.ID)
		assert.NotNil(t, authUser.LastLoginAt)
		assert.Equal(t, 0, authUser.FailedLoginCount)
		repo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("GetByEmail", ctx, "test@example.com").Return(nil, ErrUserNotFound)

		_, err := uc.Authenticate(ctx, "test@example.com", "password123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		repo.AssertExpectations(t)
	})

	t.Run("error - account locked", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.Activate()
		// Lock account by setting LockedUntil in the future
		future := time.Now().Add(30 * time.Minute)
		user.LockedUntil = &future

		repo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		_, err := uc.Authenticate(ctx, "test@example.com", "password123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountLocked)
		repo.AssertExpectations(t)
	})

	t.Run("error - account not active", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.Deactivate()

		repo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		_, err := uc.Authenticate(ctx, "test@example.com", "password123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountNotActive)
		repo.AssertExpectations(t)
	})

	t.Run("error - wrong password", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.Activate()

		repo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		_, err := uc.Authenticate(ctx, "test@example.com", "wrongpassword")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		// Failed login should be recorded
		assert.Equal(t, 1, user.FailedLoginCount)
		repo.AssertExpectations(t)
	})

	t.Run("error - account locks after 5 failed attempts", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.Activate()
		user.FailedLoginCount = 4 // One more will lock

		repo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		_, err := uc.Authenticate(ctx, "test@example.com", "wrongpassword")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.True(t, user.IsLocked())
		repo.AssertExpectations(t)
	})
}

// Test GetUser
func TestUseCase_GetUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		expectedUser, _ := NewUser("test@example.com", "password123")
		expectedUser.ID = userID

		repo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		user, err := uc.GetUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		repo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("GetByID", ctx, userID).Return(nil, ErrUserNotFound)

		_, err := uc.GetUser(ctx, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		repo.AssertExpectations(t)
	})
}

// Test GetUserByEmail
func TestUseCase_GetUserByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		expectedUser, _ := NewUser("test@example.com", "password123")

		repo.On("GetByEmail", ctx, "test@example.com").Return(expectedUser, nil)

		user, err := uc.GetUserByEmail(ctx, "test@example.com")

		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email.Value())
		repo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("GetByEmail", ctx, "test@example.com").Return(nil, ErrUserNotFound)

		_, err := uc.GetUserByEmail(ctx, "test@example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)
		repo.AssertExpectations(t)
	})
}

// Test VerifyEmail
func TestUseCase_VerifyEmail(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.ID = userID

		repo.On("GetByID", ctx, userID).Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.VerifyEmail(ctx, userID)

		require.NoError(t, err)
		assert.True(t, user.EmailVerified)
		assert.NotNil(t, user.EmailVerifiedAt)
		repo.AssertExpectations(t)
	})

	t.Run("success - already verified", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.ID = userID
		user.VerifyEmail()

		repo.On("GetByID", ctx, userID).Return(user, nil)

		err := uc.VerifyEmail(ctx, userID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
		repo.AssertNotCalled(t, "Update", ctx, mock.Anything)
	})

	t.Run("error - user not found", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("GetByID", ctx, userID).Return(nil, ErrUserNotFound)

		err := uc.VerifyEmail(ctx, userID)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

// Test ChangePassword
func TestUseCase_ChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "oldPassword123")
		user.ID = userID

		repo.On("GetByID", ctx, userID).Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.ChangePassword(ctx, userID, "oldPassword123", "newPassword456")

		require.NoError(t, err)
		// Verify new password works
		assert.NoError(t, user.CheckPassword("newPassword456"))
		repo.AssertExpectations(t)
	})

	t.Run("error - wrong old password", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "oldPassword123")
		user.ID = userID

		repo.On("GetByID", ctx, userID).Return(user, nil)

		err := uc.ChangePassword(ctx, userID, "wrongPassword", "newPassword456")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		repo.AssertExpectations(t)
	})

	t.Run("error - invalid new password", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "oldPassword123")
		user.ID = userID

		repo.On("GetByID", ctx, userID).Return(user, nil)

		err := uc.ChangePassword(ctx, userID, "oldPassword123", "weak")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

// Test SuspendUser
func TestUseCase_SuspendUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.ID = userID

		repo.On("GetByID", ctx, userID).Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.SuspendUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, UserStatusSuspended, user.Status)
		repo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("GetByID", ctx, userID).Return(nil, ErrUserNotFound)

		err := uc.SuspendUser(ctx, userID)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

// Test BanUser
func TestUseCase_BanUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.ID = userID

		repo.On("GetByID", ctx, userID).Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.BanUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, UserStatusBanned, user.Status)
		repo.AssertExpectations(t)
	})
}

// Test ActivateUser
func TestUseCase_ActivateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.ID = userID
		user.Deactivate()

		repo.On("GetByID", ctx, userID).Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.ActivateUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, UserStatusActive, user.Status)
		repo.AssertExpectations(t)
	})
}

// Test UnlockUser
func TestUseCase_UnlockUser(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user, _ := NewUser("test@example.com", "password123")
		user.ID = userID
		// Lock account
		future := time.Now().Add(30 * time.Minute)
		user.LockedUntil = &future
		user.FailedLoginCount = 5

		repo.On("GetByID", ctx, userID).Return(user, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.UnlockUser(ctx, userID)

		require.NoError(t, err)
		assert.False(t, user.IsLocked())
		assert.Equal(t, 0, user.FailedLoginCount)
		assert.Nil(t, user.LockedUntil)
		repo.AssertExpectations(t)
	})
}

// Test ListUsers
func TestUseCase_ListUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		user1, _ := NewUser("user1@example.com", "password123")
		user2, _ := NewUser("user2@example.com", "password123")
		expectedUsers := []*User{user1, user2}

		repo.On("ListUsers", ctx, 20, 0).Return(expectedUsers, 2, nil)

		users, total, err := uc.ListUsers(ctx, 20, 0)

		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, 2, total)
		repo.AssertExpectations(t)
	})

	t.Run("success - empty list", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ListUsers", ctx, 20, 0).Return([]*User{}, 0, nil)

		users, total, err := uc.ListUsers(ctx, 20, 0)

		require.NoError(t, err)
		assert.Empty(t, users)
		assert.Equal(t, 0, total)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository failure", func(t *testing.T) {
		repo := new(MockRepository)
		roleRepo := new(MockRoleRepository)
		uc := NewUseCase(repo, roleRepo)

		repo.On("ListUsers", ctx, 20, 0).Return(nil, 0, errors.New("db error"))

		_, _, err := uc.ListUsers(ctx, 20, 0)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}
