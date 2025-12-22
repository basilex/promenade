package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserContactRepository is a mock implementation of UserContactRepository
//  Module-independent: imports only module types, no core dependencies
type MockUserContactRepository struct {
	mock.Mock
}

// Compile-time check to ensure MockUserContactRepository implements UserContactRepository interface
var _ repository.UserContactRepository = (*MockUserContactRepository)(nil)

func (m *MockUserContactRepository) Create(ctx context.Context, contact *entity.UserContact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockUserContactRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserContact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *MockUserContactRepository) GetUserContacts(ctx context.Context, userID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID, includeInactive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *MockUserContactRepository) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *MockUserContactRepository) GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) (*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *MockUserContactRepository) GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *MockUserContactRepository) Update(ctx context.Context, contact *entity.UserContact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockUserContactRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserContactRepository) SetPrimary(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, id uuidv7.UUID) error {
	args := m.Called(ctx, userID, contactType, id)
	return args.Error(0)
}

func (m *MockUserContactRepository) VerifyContact(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
