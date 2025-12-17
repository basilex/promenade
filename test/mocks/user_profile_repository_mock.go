package mocks

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/mock"
)

// MockUserProfileRepository is a mock implementation of repository.UserProfileRepository
type MockUserProfileRepository struct {
	mock.Mock
}

func (m *MockUserProfileRepository) Create(ctx context.Context, profile *entity.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserProfileRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) GetByNickname(ctx context.Context, nickname string) (*entity.UserProfile, error) {
	args := m.Called(ctx, nickname)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) Update(ctx context.Context, profile *entity.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserProfileRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserProfileRepository) List(ctx context.Context, limit, offset int, isPublic *bool) ([]*entity.UserProfile, error) {
	args := m.Called(ctx, limit, offset, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) UpdateLastSeen(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserProfileRepository) IncrementProfileViews(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserProfileRepository) Ban(ctx context.Context, id uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error {
	args := m.Called(ctx, id, reason, bannedBy)
	return args.Error(0)
}

func (m *MockUserProfileRepository) Unban(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserProfileRepository) SetVerified(ctx context.Context, id uuidv7.UUID, verified bool) error {
	args := m.Called(ctx, id, verified)
	return args.Error(0)
}

func (m *MockUserProfileRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserProfile), args.Error(1)
}
