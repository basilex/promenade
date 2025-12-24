package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock IUserRepository
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.UserStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id uuidv7.UUID, loginTime time.Time) error {
	args := m.Called(ctx, id, loginTime)
	return args.Error(0)
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uuidv7.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *mockUserRepository) VerifyEmail(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) Suspend(ctx context.Context, id uuidv7.UUID, reason string, until *time.Time) error {
	args := m.Called(ctx, id, reason, until)
	return args.Error(0)
}

func (m *mockUserRepository) Ban(ctx context.Context, id uuidv7.UUID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockUserRepository) Reactivate(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) Deactivate(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock ISessionRepository
type mockSessionRepository struct {
	mock.Mock
}

func (m *mockSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *mockSessionRepository) GetByRefreshToken(ctx context.Context, hashedToken string) (*entity.Session, error) {
	args := m.Called(ctx, hashedToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *mockSessionRepository) Update(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSessionRepository) GetUserSessions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Session), args.Error(1)
}

func (m *mockSessionRepository) CountUserSessions(ctx context.Context, userID uuidv7.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockSessionRepository) GetOldestSession(ctx context.Context, userID uuidv7.UUID) (*entity.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test helpers
func setupAuthUseCase(t *testing.T) (*authUseCase, *mockUserRepository, *mockSessionRepository) {
	userRepo := new(mockUserRepository)
	sessionRepo := new(mockSessionRepository)
	jwtManager := jwtpkg.NewJWTManager(
		"test-secret-key-32-bytes-long!",
		15*time.Minute,
		7*24*time.Hour,
	)
	eventBus := memory.NewMemoryBus(bus.BusConfig{
		WorkerPoolSize: 4,
		BufferSize:     100,
	})

	uc := NewAuthUseCase(userRepo, sessionRepo, jwtManager, eventBus).(*authUseCase)
	return uc, userRepo, sessionRepo
}

func createTestUser(email, name string, status entity.UserStatus) *entity.User {
	user := &entity.User{
		ID:        uuidv7.New(),
		Email:     email,
		Name:      name,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = user.HashPassword("password123")
	return user
}

// Tests for Register
func TestAuthUseCase_Register_Success(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	email := "test@example.com"
	name := "Test User"
	password := "password123"

	// Mock: user doesn't exist
	userRepo.On("GetByEmail", ctx, email).Return(nil, entity.ErrNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

	user, err := uc.Register(ctx, email, name, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, entity.UserStatusUnverified, user.Status)
	assert.NotEmpty(t, user.Password)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Register_EmailAlreadyExists(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	email := "existing@example.com"
	existingUser := createTestUser(email, "Existing User", entity.UserStatusActive)

	userRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	user, err := uc.Register(ctx, email, "Test User", "password123")

	assert.Error(t, err)
	assert.Equal(t, ErrEmailAlreadyExists, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Register_RepositoryError(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, entity.ErrNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).
		Return(errors.New("database error"))

	user, err := uc.Register(ctx, "test@example.com", "Test User", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")
	userRepo.AssertExpectations(t)
}

// Tests for Login
func TestAuthUseCase_Login_Success(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusActive)
	userAgent := "Mozilla/5.0"
	ipAddress := "192.168.1.1"

	userRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)
	sessionRepo.On("CountUserSessions", ctx, user.ID).Return(0, nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID, mock.AnythingOfType("time.Time")).Return(nil)

	accessToken, refreshToken, returnedUser, err := uc.Login(ctx, user.Email, "password123", userAgent, ipAddress)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, user.ID, returnedUser.ID)
	assert.Equal(t, user.Email, returnedUser.Email)
	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, entity.ErrNotFound)

	accessToken, refreshToken, user, err := uc.Login(ctx, "test@example.com", "wrongpass", "agent", "ip")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusActive)
	userRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)

	accessToken, refreshToken, returnedUser, err := uc.Login(ctx, user.Email, "wrongpassword", "agent", "ip")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, returnedUser)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_SuspendedUser(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusSuspended)
	userRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)

	accessToken, refreshToken, returnedUser, err := uc.Login(ctx, user.Email, "password123", "agent", "ip")

	assert.Error(t, err)
	assert.Equal(t, ErrUserSuspended, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, returnedUser)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_BannedUser(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusBanned)
	userRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)

	accessToken, refreshToken, returnedUser, err := uc.Login(ctx, user.Email, "password123", "agent", "ip")

	assert.Error(t, err)
	assert.Equal(t, ErrUserBanned, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, returnedUser)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_MaxSessionsReached(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusActive)
	oldestSession := &entity.Session{
		ID:        uuidv7.New(),
		UserID:    user.ID,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	userRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)
	sessionRepo.On("CountUserSessions", ctx, user.ID).Return(MaxConcurrentSessions, nil)
	sessionRepo.On("GetOldestSession", ctx, user.ID).Return(oldestSession, nil)
	sessionRepo.On("Delete", ctx, oldestSession.ID).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID, mock.AnythingOfType("time.Time")).Return(nil)

	accessToken, refreshToken, returnedUser, err := uc.Login(ctx, user.Email, "password123", "agent", "ip")

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, user.ID, returnedUser.ID)
	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_SessionCreationError(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusActive)

	userRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)
	sessionRepo.On("CountUserSessions", ctx, user.ID).Return(0, nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Session")).
		Return(errors.New("database error"))

	accessToken, refreshToken, returnedUser, err := uc.Login(ctx, user.Email, "password123", "agent", "ip")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create session")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, returnedUser)
	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

// Tests for Logout
func TestAuthUseCase_Logout_Success(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	refreshToken := "test-refresh-token"
	hashedToken := uc.hashToken(refreshToken)
	session := &entity.Session{
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		RefreshToken: hashedToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	sessionRepo.On("GetByRefreshToken", ctx, hashedToken).Return(session, nil)
	sessionRepo.On("Delete", ctx, session.ID).Return(nil)

	err := uc.Logout(ctx, refreshToken)

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout_InvalidToken(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	refreshToken := "invalid-token"
	hashedToken := uc.hashToken(refreshToken)

	sessionRepo.On("GetByRefreshToken", ctx, hashedToken).Return(nil, entity.ErrNotFound)

	err := uc.Logout(ctx, refreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	sessionRepo.AssertExpectations(t)
}

// Tests for RefreshToken
func TestAuthUseCase_RefreshToken_Success(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusActive)
	oldRefreshToken := "old-refresh-token"
	hashedOldToken := uc.hashToken(oldRefreshToken)
	session := &entity.Session{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		RefreshToken: hashedOldToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	sessionRepo.On("GetByRefreshToken", ctx, hashedOldToken).Return(session, nil)
	userRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	sessionRepo.On("Update", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)

	newAccessToken, newRefreshToken, err := uc.RefreshToken(ctx, oldRefreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)
	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_ExpiredSession(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	oldRefreshToken := "old-refresh-token"
	hashedOldToken := uc.hashToken(oldRefreshToken)
	expiredSession := &entity.Session{
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		RefreshToken: hashedOldToken,
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}

	sessionRepo.On("GetByRefreshToken", ctx, hashedOldToken).Return(expiredSession, nil)
	sessionRepo.On("Delete", ctx, expiredSession.ID).Return(nil)

	newAccessToken, newRefreshToken, err := uc.RefreshToken(ctx, oldRefreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_UserNotActive(t *testing.T) {
	uc, userRepo, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	user := createTestUser("test@example.com", "Test User", entity.UserStatusSuspended)
	oldRefreshToken := "old-refresh-token"
	hashedOldToken := uc.hashToken(oldRefreshToken)
	session := &entity.Session{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		RefreshToken: hashedOldToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	sessionRepo.On("GetByRefreshToken", ctx, hashedOldToken).Return(session, nil)
	userRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	newAccessToken, newRefreshToken, err := uc.RefreshToken(ctx, oldRefreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotActive, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

// Tests for GetMe
func TestAuthUseCase_GetMe_Success(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	expectedUser := createTestUser("test@example.com", "Test User", entity.UserStatusActive)
	userRepo.On("GetByID", ctx, expectedUser.ID).Return(expectedUser, nil)

	user, err := uc.GetMe(ctx, expectedUser.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Email, user.Email)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_GetMe_UserNotFound(t *testing.T) {
	uc, userRepo, _ := setupAuthUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	userRepo.On("GetByID", ctx, userID).Return(nil, entity.ErrNotFound)

	user, err := uc.GetMe(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

// Tests for GetUserSessions
func TestAuthUseCase_GetUserSessions_Success(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	expectedSessions := []*entity.Session{
		{ID: uuidv7.New(), UserID: userID, CreatedAt: time.Now()},
		{ID: uuidv7.New(), UserID: userID, CreatedAt: time.Now()},
	}

	sessionRepo.On("GetUserSessions", ctx, userID).Return(expectedSessions, nil)

	sessions, err := uc.GetUserSessions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, expectedSessions[0].ID, sessions[0].ID)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_GetUserSessions_Empty(t *testing.T) {
	uc, _, sessionRepo := setupAuthUseCase(t)
	ctx := context.Background()

	userID := uuidv7.New()
	sessionRepo.On("GetUserSessions", ctx, userID).Return([]*entity.Session{}, nil)

	sessions, err := uc.GetUserSessions(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, sessions)
	sessionRepo.AssertExpectations(t)
}
