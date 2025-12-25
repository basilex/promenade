package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// mockUserContactRepository implements IUserContactRepository for testing
type mockUserContactRepository struct {
	mock.Mock
}

func (m *mockUserContactRepository) Create(ctx context.Context, contact *entity.UserContact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *mockUserContactRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserContact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *mockUserContactRepository) GetUserContacts(ctx context.Context, userID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID, includeInactive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *mockUserContactRepository) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *mockUserContactRepository) GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) (*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *mockUserContactRepository) GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *mockUserContactRepository) Update(ctx context.Context, contact *entity.UserContact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *mockUserContactRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserContactRepository) SetPrimary(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, id uuidv7.UUID) error {
	args := m.Called(ctx, userID, contactType, id)
	return args.Error(0)
}

func (m *mockUserContactRepository) VerifyContact(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockUserProfileRepository implements IUserProfileRepository for testing
type mockUserProfileRepository struct {
	mock.Mock
}

func (m *mockUserProfileRepository) Create(ctx context.Context, profile *entity.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockUserProfileRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *mockUserProfileRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *mockUserProfileRepository) GetByNickname(ctx context.Context, nickname string) (*entity.UserProfile, error) {
	args := m.Called(ctx, nickname)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *mockUserProfileRepository) Update(ctx context.Context, profile *entity.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockUserProfileRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserProfileRepository) List(ctx context.Context, limit, offset int, isPublic *bool) ([]*entity.UserProfile, error) {
	args := m.Called(ctx, limit, offset, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserProfile), args.Error(1)
}

func (m *mockUserProfileRepository) UpdateLastSeen(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *mockUserProfileRepository) IncrementProfileViews(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *mockUserProfileRepository) Ban(ctx context.Context, profileID uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error {
	args := m.Called(ctx, profileID, reason, bannedBy)
	return args.Error(0)
}

func (m *mockUserProfileRepository) Unban(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *mockUserProfileRepository) SetVerified(ctx context.Context, profileID uuidv7.UUID, verified bool) error {
	args := m.Called(ctx, profileID, verified)
	return args.Error(0)
}

func (m *mockUserProfileRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserProfile), args.Error(1)
}
