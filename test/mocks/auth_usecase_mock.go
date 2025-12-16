package mocks

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/mock"
)

// MockAuthUseCase is a mock implementation of usecase.AuthUseCase
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Register(ctx context.Context, email, name, password string) (*entity.User, error) {
	args := m.Called(ctx, email, name, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockAuthUseCase) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, *entity.User, error) {
	args := m.Called(ctx, email, password, userAgent, ipAddress)
	var user *entity.User
	if args.Get(2) != nil {
		user = args.Get(2).(*entity.User)
	}
	return args.String(0), args.String(1), user, args.Error(3)
}

func (m *MockAuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthUseCase) GetMe(ctx context.Context, userID uuidv7.UUID) (*entity.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockAuthUseCase) GetUserSessions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Session), args.Error(1)
}

func (m *MockAuthUseCase) SuspendUser(ctx context.Context, userID uuidv7.UUID, reason string, until *time.Time) error {
	args := m.Called(ctx, userID, reason, until)
	return args.Error(0)
}

func (m *MockAuthUseCase) BanUser(ctx context.Context, userID uuidv7.UUID, reason string) error {
	args := m.Called(ctx, userID, reason)
	return args.Error(0)
}

func (m *MockAuthUseCase) ReactivateUser(ctx context.Context, userID uuidv7.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
