package user_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	userHTTP "github.com/basilex/promenade/internal/contexts/identity/user/adapter/http"
	userAggregate "github.com/basilex/promenade/internal/contexts/identity/user/aggregate"
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type MockUserUseCase struct {
	RegisterFunc       func(ctx context.Context, email, name, password string) (*userAggregate.User, error)
	AuthenticateFunc   func(ctx context.Context, email, password string) (*userAggregate.User, error)
	GetUserFunc        func(ctx context.Context, userID uuidv7.UUID) (*userAggregate.User, error)
	GetUserByEmailFunc func(ctx context.Context, email string) (*userAggregate.User, error)
	VerifyEmailFunc    func(ctx context.Context, userID uuidv7.UUID) error
	ChangePasswordFunc func(ctx context.Context, userID uuidv7.UUID, oldPassword, newPassword string) error
	SuspendUserFunc    func(ctx context.Context, userID uuidv7.UUID) error
	BanUserFunc        func(ctx context.Context, userID uuidv7.UUID) error
	ActivateUserFunc   func(ctx context.Context, userID uuidv7.UUID) error
	UnlockUserFunc     func(ctx context.Context, userID uuidv7.UUID) error
	ListUsersFunc      func(ctx context.Context, limit, offset int) ([]*userAggregate.User, int, error)
}

func (m *MockUserUseCase) Register(ctx context.Context, email, name, password string) (*userAggregate.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, name, password)
	}
	return nil, nil
}

func (m *MockUserUseCase) Authenticate(ctx context.Context, email, password string) (*userAggregate.User, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *MockUserUseCase) GetUser(ctx context.Context, userID uuidv7.UUID) (*userAggregate.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserUseCase) GetUserByEmail(ctx context.Context, email string) (*userAggregate.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserUseCase) VerifyEmail(ctx context.Context, userID uuidv7.UUID) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserUseCase) ChangePassword(ctx context.Context, userID uuidv7.UUID, oldPassword, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(ctx, userID, oldPassword, newPassword)
	}
	return nil
}

func (m *MockUserUseCase) SuspendUser(ctx context.Context, userID uuidv7.UUID) error {
	if m.SuspendUserFunc != nil {
		return m.SuspendUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserUseCase) BanUser(ctx context.Context, userID uuidv7.UUID) error {
	if m.BanUserFunc != nil {
		return m.BanUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserUseCase) ActivateUser(ctx context.Context, userID uuidv7.UUID) error {
	if m.ActivateUserFunc != nil {
		return m.ActivateUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserUseCase) UnlockUser(ctx context.Context, userID uuidv7.UUID) error {
	if m.UnlockUserFunc != nil {
		return m.UnlockUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserUseCase) ListUsers(ctx context.Context, limit, offset int) ([]*userAggregate.User, int, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func fakeUser() *userAggregate.User {
	u, _ := userAggregate.NewUser("test@example.com", "password123")
	return u
}

func TestUserHandler_Register_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		RegisterFunc: func(ctx context.Context, email, name, password string) (*userAggregate.User, error) {
			return fakeUser(), nil
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.POST("/users/register", handler.Register)

	resp := smoke.MakeRequest(t, router, "POST", "/users/register", map[string]any{
		"email":    "test@example.com",
		"name":     "Test User",
		"password": "password123",
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestUserHandler_Register_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.POST("/users/register", handler.Register)

	resp := smoke.MakeRequest(t, router, "POST", "/users/register", map[string]any{
		"email": "invalid-email",
	})

	smoke.AssertErrorResponse(t, resp, 400, "BAD_REQUEST")
}

func TestUserHandler_Login_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		AuthenticateFunc: func(ctx context.Context, email, password string) (*userAggregate.User, error) {
			return fakeUser(), nil
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.POST("/users/login", handler.Login)

	resp := smoke.MakeRequest(t, router, "POST", "/users/login", map[string]any{
		"email":    "test@example.com",
		"password": "password123",
	})

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestUserHandler_Login_Unauthorized(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		AuthenticateFunc: func(ctx context.Context, email, password string) (*userAggregate.User, error) {
			return nil, user.ErrInvalidCredentials
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.POST("/users/login", handler.Login)

	resp := smoke.MakeRequest(t, router, "POST", "/users/login", map[string]any{
		"email":    "test@example.com",
		"password": "wrongpassword",
	})

	smoke.AssertErrorResponse(t, resp, 401, "UNAUTHORIZED")
}

func TestUserHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		GetUserFunc: func(ctx context.Context, userID uuidv7.UUID) (*userAggregate.User, error) {
			return fakeUser(), nil
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.GET("/users/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/users/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		GetUserFunc: func(ctx context.Context, userID uuidv7.UUID) (*userAggregate.User, error) {
			return nil, user.ErrUserNotFound
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.GET("/users/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/users/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}

func TestUserHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		ListUsersFunc: func(ctx context.Context, limit, offset int) ([]*userAggregate.User, int, error) {
			return []*userAggregate.User{fakeUser()}, 1, nil
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.GET("/users", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/users?page=1&page_size=20", nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestUserHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockUserUseCase{
		ListUsersFunc: func(ctx context.Context, limit, offset int) ([]*userAggregate.User, int, error) {
			return []*userAggregate.User{}, 0, nil
		},
	}

	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  900,
		RefreshTokenDuration: 604800,
		Issuer:               "test",
	})

	handler := userHTTP.NewUserHandler(mockUC, jwtManager, nil)
	router.GET("/users", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/users?page=1&page_size=20", nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}
