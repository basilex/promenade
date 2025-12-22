package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUserContactUseCase is a mock implementation of UserContactUseCase
//  Module-independent: imports only module types, no core dependencies
type MockUserContactUseCase struct {
	mock.Mock
}

// Compile-time check to ensure MockUserContactUseCase implements UserContactUseCase interface
var _ usecase.UserContactUseCase = (*MockUserContactUseCase)(nil)

func (m *MockUserContactUseCase) CreateContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, contactValue string, label *string, isPublic bool, availableFrom, availableTo *time.Time, availableDays []string, timezone *string, notes *string) (*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType, contactValue, label, isPublic, availableFrom, availableTo, availableDays, timezone, notes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) GetContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) (*entity.UserContact, error) {
	args := m.Called(ctx, contactID, requestingUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID, requestingUserID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID, requestingUserID, includeInactive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, requestingUserID uuidv7.UUID) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType, requestingUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, requestingUserID uuidv7.UUID) (*entity.UserContact, error) {
	args := m.Called(ctx, userID, contactType, requestingUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) UpdateContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID, contactType entity.ContactType, contactValue string, label *string, isPublic bool, availableFrom, availableTo *time.Time, availableDays []string, timezone *string, notes *string) (*entity.UserContact, error) {
	args := m.Called(ctx, contactID, requestingUserID, contactType, contactValue, label, isPublic, availableFrom, availableTo, availableDays, timezone, notes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserContact), args.Error(1)
}

func (m *MockUserContactUseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error {
	args := m.Called(ctx, contactID, requestingUserID)
	return args.Error(0)
}

func (m *MockUserContactUseCase) SetPrimaryContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error {
	args := m.Called(ctx, contactID, requestingUserID)
	return args.Error(0)
}

func (m *MockUserContactUseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func (m *MockUserContactUseCase) ToggleContactActive(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error {
	args := m.Called(ctx, contactID, requestingUserID)
	return args.Error(0)
}
