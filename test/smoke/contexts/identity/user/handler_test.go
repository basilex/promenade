package user_test

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

	"github.com/basilex/promenade/internal/contexts/identity/user"
	userHTTP "github.com/basilex/promenade/internal/contexts/identity/user/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserUseCase is a mock implementation of user.IUseCase for smoke tests
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) Register(ctx context.Context, email, name, password string) (*user.User, error) {
	args := m.Called(ctx, email, name, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserUseCase) Authenticate(ctx context.Context, email, password string) (*user.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserUseCase) GetUser(ctx context.Context, userID uuidv7.UUID) (*user.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserUseCase) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserUseCase) VerifyEmail(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) ChangePassword(ctx context.Context, userID uuidv7.UUID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserUseCase) SuspendUser(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) BanUser(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) ActivateUser(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) UnlockUser(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Int(1), args.Error(2)
}

func setupUserRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestUserHandler_Smoke - smoke tests for User handler (mock-based, no database)
func TestUserHandler_Smoke(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := userHTTP.NewUserHandler(mockUC)
	router := setupUserRouter()

	// Register routes
	router.POST("/users/register", handler.Register)
	router.POST("/users/login", handler.Login)
	router.GET("/users/:id", handler.GetByID)
	router.GET("/users/email/:email", handler.GetByEmail)
	router.POST("/users/:id/verify-email", handler.VerifyEmail)
	router.POST("/users/:id/change-password", handler.ChangePassword)
	router.POST("/users/:id/suspend", handler.Suspend)
	router.POST("/users/:id/ban", handler.Ban)
	router.POST("/users/:id/activate", handler.Activate)
	router.POST("/users/:id/unlock", handler.Unlock)
	router.GET("/users", handler.List)

	t.Run("Register returns 201", func(t *testing.T) {
		userID := uuidv7.New()
		email := "newuser@example.com"
		name := "New User"

		// Create mock user (no real hashing needed for smoke test)
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("Register", mock.Anything, email, name, "password123").Return(u, nil).Once()

		reqBody := userHTTP.RegisterRequest{
			Email:    email,
			Name:     name,
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Register should return 201")
		mockUC.AssertExpectations(t)
	})

	t.Run("Login returns 200", func(t *testing.T) {
		email := "user@example.com"
		password := "password123"

		// Mock user for authentication
		u := &user.User{
			ID:           uuidv7.New(),
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("Authenticate", mock.Anything, email, password).Return(u, nil).Once()

		reqBody := userHTTP.LoginRequest{
			Email:    email,
			Password: password,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Login should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("GetByID returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByID should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("GetByEmail returns 200", func(t *testing.T) {
		email := "user@example.com"
		u := &user.User{
			ID:           uuidv7.New(),
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("GetUserByEmail", mock.Anything, email).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users/email/"+email, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByEmail should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("VerifyEmail returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("VerifyEmail", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/verify-email", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "VerifyEmail should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("ChangePassword returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "newHashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("ChangePassword", mock.Anything, mock.AnythingOfType("uuid.UUID"), "oldPassword", "newPassword").Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		reqBody := userHTTP.ChangePasswordRequest{
			OldPassword: "oldPassword",
			NewPassword: "newPassword",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "ChangePassword should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("Suspend returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusSuspended,
		}

		mockUC.On("SuspendUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/suspend", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Suspend should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("Ban returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusBanned,
		}

		mockUC.On("BanUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/ban", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Ban should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("Activate returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("ActivateUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Activate should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("Unlock returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("UnlockUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/unlock", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Unlock should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("List returns 200", func(t *testing.T) {
		users := []*user.User{
			{ID: uuidv7.New()},
			{ID: uuidv7.New()},
		}
		totalCount := 2

		// Handler uses page-based pagination: page=1, page_size=20
		mockUC.On("ListUsers", mock.Anything, 1, 20).Return(users, totalCount, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "List should return 200")
		mockUC.AssertExpectations(t)
	})
}
