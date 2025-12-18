package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuidv7.UUID, lastLogin time.Time) error {
	args := m.Called(ctx, id, lastLogin)
	return args.Error(0)
}

func (m *MockUserRepository) Suspend(ctx context.Context, id uuidv7.UUID, reason string, until *time.Time) error {
	args := m.Called(ctx, id, reason, until)
	return args.Error(0)
}

func (m *MockUserRepository) Ban(ctx context.Context, id uuidv7.UUID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockUserRepository) Reactivate(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Deactivate(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.UserStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuidv7.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) VerifyEmail(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
