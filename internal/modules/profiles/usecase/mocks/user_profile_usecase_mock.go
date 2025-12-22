package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserProfileUseCase is a mock implementation of UserProfileUseCase
//  Module-independent: imports only module types, no core dependencies
type MockUserProfileUseCase struct {
	mock.Mock
}

// Compile-time check to ensure MockUserProfileUseCase implements UserProfileUseCase interface
var _ usecase.UserProfileUseCase = (*MockUserProfileUseCase)(nil)

func (m *MockUserProfileUseCase) CreateProfile(ctx context.Context, userID uuidv7.UUID, profile *entity.UserProfile) (*entity.UserProfile, error) {
	args := m.Called(ctx, userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) GetProfile(ctx context.Context, profileID uuidv7.UUID, viewerID *uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, profileID, viewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) GetProfileByUserID(ctx context.Context, userID uuidv7.UUID, viewerID *uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, userID, viewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) GetProfileByNickname(ctx context.Context, nickname string, viewerID *uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, nickname, viewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) UpdateProfile(ctx context.Context, profileID uuidv7.UUID, userID uuidv7.UUID, updates *entity.UserProfile) (*entity.UserProfile, error) {
	args := m.Called(ctx, profileID, userID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) DeleteProfile(ctx context.Context, profileID uuidv7.UUID, userID uuidv7.UUID) error {
	args := m.Called(ctx, profileID, userID)
	return args.Error(0)
}

func (m *MockUserProfileUseCase) ListProfiles(ctx context.Context, limit, offset int, publicOnly bool) ([]*entity.UserProfile, error) {
	args := m.Called(ctx, limit, offset, publicOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) SearchProfiles(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileUseCase) UpdateLastSeen(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockUserProfileUseCase) IncrementViews(ctx context.Context, profileID uuidv7.UUID, viewerID *uuidv7.UUID) error {
	args := m.Called(ctx, profileID, viewerID)
	return args.Error(0)
}

func (m *MockUserProfileUseCase) BanProfile(ctx context.Context, profileID uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error {
	args := m.Called(ctx, profileID, reason, bannedBy)
	return args.Error(0)
}

func (m *MockUserProfileUseCase) UnbanProfile(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockUserProfileUseCase) VerifyProfile(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockUserProfileUseCase) UnverifyProfile(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}
