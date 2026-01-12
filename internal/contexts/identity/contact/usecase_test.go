package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of IRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, contact *Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Contact), args.Error(1)
}

func (m *MockRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*Contact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Contact), args.Error(1)
}

func (m *MockRepository) GetByUserIDAndType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) ([]*Contact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Contact), args.Error(1)
}

func (m *MockRepository) GetPrimaryByUserIDAndType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) (*Contact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Contact), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, contact *Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) SetPrimary(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ExistsPrimaryForUserAndType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) (bool, error) {
	args := m.Called(ctx, userID, contactType)
	return args.Bool(0), args.Error(1)
}

func TestUseCase_CreateEmailContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success - non-primary", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("Create", ctx, mock.Anything).Return(nil)

		contact, err := uc.CreateEmailContact(ctx, userID, "test@example.com", "Work", false)

		require.NoError(t, err)
		assert.NotNil(t, contact)
		assert.Equal(t, ContactTypeEmail, contact.Type)
		assert.Equal(t, "test@example.com", contact.Email.Value())
		assert.False(t, contact.IsPrimary)
		repo.AssertExpectations(t)
	})

	t.Run("success - primary", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsPrimaryForUserAndType", ctx, userID, ContactTypeEmail).Return(false, nil)
		repo.On("Create", ctx, mock.Anything).Return(nil)

		contact, err := uc.CreateEmailContact(ctx, userID, "test@example.com", "Work", true)

		require.NoError(t, err)
		assert.True(t, contact.IsPrimary)
		repo.AssertExpectations(t)
	})

	t.Run("error - primary already exists", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsPrimaryForUserAndType", ctx, userID, ContactTypeEmail).Return(true, nil)

		_, err := uc.CreateEmailContact(ctx, userID, "test@example.com", "Work", true)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPrimaryContactExists))
		repo.AssertExpectations(t)
	})

	t.Run("error - invalid email", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		_, err := uc.CreateEmailContact(ctx, userID, "invalid-email", "Work", false)

		assert.Error(t, err)
	})

	t.Run("error - repository failure", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.CreateEmailContact(ctx, userID, "test@example.com", "Work", false)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_CreatePhoneContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("Create", ctx, mock.Anything).Return(nil)

		contact, err := uc.CreatePhoneContact(ctx, userID, "+380501234567", "Mobile", false)

		require.NoError(t, err)
		assert.NotNil(t, contact)
		assert.Equal(t, ContactTypePhone, contact.Type)
		repo.AssertExpectations(t)
	})

	t.Run("error - primary already exists", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsPrimaryForUserAndType", ctx, userID, ContactTypePhone).Return(true, nil)

		_, err := uc.CreatePhoneContact(ctx, userID, "+380501234567", "Mobile", true)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_CreateAddressContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("Create", ctx, mock.Anything).Return(nil)

		contact, err := uc.CreateAddressContact(ctx, userID, "Street", "City", "US", "12345", "Home", false)

		require.NoError(t, err)
		assert.NotNil(t, contact)
		assert.Equal(t, ContactTypeAddress, contact.Type)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_GetContact(t *testing.T) {
	ctx := context.Background()
	contactID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		expectedContact, _ := NewEmailContact(uuidv7.New(), "test@example.com", "Work")
		expectedContact.ID = contactID

		repo.On("GetByID", ctx, contactID).Return(expectedContact, nil)

		contact, err := uc.GetContact(ctx, contactID)

		require.NoError(t, err)
		assert.Equal(t, contactID, contact.ID)
		repo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("GetByID", ctx, contactID).Return(nil, errors.New("not found"))

		_, err := uc.GetContact(ctx, contactID)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_GetUserContacts(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		contact1, _ := NewEmailContact(userID, "test1@example.com", "Work")
		contact2, _ := NewPhoneContact(userID, "+380501234567", "Mobile")
		expectedContacts := []*Contact{contact1, contact2}

		repo.On("GetByUserID", ctx, userID).Return(expectedContacts, nil)

		contacts, err := uc.GetUserContacts(ctx, userID)

		require.NoError(t, err)
		assert.Len(t, contacts, 2)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_SetAsPrimary(t *testing.T) {
	ctx := context.Background()
	contactID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		contact, _ := NewEmailContact(uuidv7.New(), "test@example.com", "Work")
		contact.ID = contactID

		repo.On("GetByID", ctx, contactID).Return(contact, nil)
		repo.On("SetPrimary", ctx, contactID).Return(nil)

		err := uc.SetAsPrimary(ctx, contactID)

		require.NoError(t, err)
		assert.True(t, contact.IsPrimary)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_VerifyContact(t *testing.T) {
	ctx := context.Background()
	contactID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		contact, _ := NewEmailContact(uuidv7.New(), "test@example.com", "Work")
		contact.ID = contactID

		repo.On("GetByID", ctx, contactID).Return(contact, nil)
		repo.On("Update", ctx, contact).Return(nil)

		err := uc.VerifyContact(ctx, contactID)

		require.NoError(t, err)
		assert.True(t, contact.IsVerified)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_UpdateVisibility(t *testing.T) {
	ctx := context.Background()
	contactID := uuidv7.New()

	t.Run("make public", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		contact, _ := NewEmailContact(uuidv7.New(), "test@example.com", "Work")
		contact.ID = contactID

		repo.On("GetByID", ctx, contactID).Return(contact, nil)
		repo.On("Update", ctx, contact).Return(nil)

		err := uc.UpdateVisibility(ctx, contactID, true)

		require.NoError(t, err)
		assert.True(t, contact.IsPublic)
		repo.AssertExpectations(t)
	})

	t.Run("make private", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		contact, _ := NewEmailContact(uuidv7.New(), "test@example.com", "Work")
		contact.ID = contactID
		contact.MakePublic()

		repo.On("GetByID", ctx, contactID).Return(contact, nil)
		repo.On("Update", ctx, contact).Return(nil)

		err := uc.UpdateVisibility(ctx, contactID, false)

		require.NoError(t, err)
		assert.False(t, contact.IsPublic)
		repo.AssertExpectations(t)
	})
}
