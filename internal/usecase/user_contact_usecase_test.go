package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
)

func TestUserContactUseCase_CreateContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		label := "Work"
		mockRepo.On("Create", ctx, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.UserID == userID && c.ContactType == entity.ContactTypeEmail
		})).Return(nil)

		contact, err := uc.CreateContact(ctx, userID, entity.ContactTypeEmail,
			"test@example.com", &label, true, nil, nil, nil, nil, nil)

		require.NoError(t, err)
		assert.NotNil(t, contact)
		assert.Equal(t, userID, contact.UserID)
		assert.Equal(t, entity.ContactTypeEmail, contact.ContactType)
		assert.Equal(t, "test@example.com", contact.ContactValue)
		assert.False(t, contact.IsVerified)
		assert.False(t, contact.IsPrimary)
		assert.True(t, contact.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		_, err := uc.CreateContact(ctx, userID, entity.ContactType("invalid"),
			"test@example.com", nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidContactType)
	})

	t.Run("invalid availability times", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		from := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)
		to := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)

		_, err := uc.CreateContact(ctx, userID, entity.ContactTypeEmail,
			"test@example.com", nil, true, &from, &to, nil, nil, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAvailability)
	})

	t.Run("empty contact value", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		_, err := uc.CreateContact(ctx, userID, entity.ContactTypeEmail,
			"", nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
	})
}

func TestUserContactUseCase_GetContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("get own contact", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		expectedContact := &entity.UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "test@example.com",
			IsPublic:     false,
			IsActive:     true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(expectedContact, nil)

		contact, err := uc.GetContact(ctx, contactID, userID)

		require.NoError(t, err)
		assert.Equal(t, contactID, contact.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get public contact of other user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		expectedContact := &entity.UserContact{
			ID:           contactID,
			UserID:       otherUserID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "test@example.com",
			IsPublic:     true,
			IsActive:     true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(expectedContact, nil)

		contact, err := uc.GetContact(ctx, contactID, userID)

		require.NoError(t, err)
		assert.Equal(t, contactID, contact.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized to view private contact", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		expectedContact := &entity.UserContact{
			ID:           contactID,
			UserID:       otherUserID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "test@example.com",
			IsPublic:     false,
			IsActive:     true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(expectedContact, nil)

		_, err := uc.GetContact(ctx, contactID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		mockRepo.On("GetByID", ctx, contactID).Return(nil, entity.ErrNotFound)

		_, err := uc.GetContact(ctx, contactID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrContactNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_GetUserContacts(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("get own contacts", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		expectedContacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, ContactType: entity.ContactTypeEmail, IsPublic: false, IsActive: true},
			{ID: uuidv7.New(), UserID: userID, ContactType: entity.ContactTypePhone, IsPublic: true, IsActive: true},
		}

		mockRepo.On("GetUserContacts", ctx, userID, false).Return(expectedContacts, nil)

		contacts, err := uc.GetUserContacts(ctx, userID, userID, false)

		require.NoError(t, err)
		assert.Len(t, contacts, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get public contacts of other user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		publicContacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: otherUserID, ContactType: entity.ContactTypeEmail, IsPublic: true, IsActive: true},
		}

		mockRepo.On("GetPublicContacts", ctx, otherUserID).Return(publicContacts, nil)

		contacts, err := uc.GetUserContacts(ctx, otherUserID, userID, false)

		require.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.True(t, contacts[0].IsPublic)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_UpdateContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "old@example.com",
			IsActive:     true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.ContactValue == "new@example.com"
		})).Return(nil)

		newLabel := "Updated"
		contact, err := uc.UpdateContact(ctx, contactID, userID, entity.ContactTypeEmail,
			"new@example.com", &newLabel, true, nil, nil, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, "new@example.com", contact.ContactValue)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized update", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:           contactID,
			UserID:       otherUserID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "old@example.com",
			IsActive:     true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)

		_, err := uc.UpdateContact(ctx, contactID, userID, entity.ContactTypeEmail,
			"new@example.com", nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		mockRepo.On("GetByID", ctx, contactID).Return(nil, entity.ErrNotFound)

		_, err := uc.UpdateContact(ctx, contactID, userID, entity.ContactTypeEmail,
			"new@example.com", nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrContactNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_DeleteContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsActive: true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)
		mockRepo.On("Delete", ctx, contactID).Return(nil)

		err := uc.DeleteContact(ctx, contactID, userID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:       contactID,
			UserID:   otherUserID,
			IsActive: true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)

		err := uc.DeleteContact(ctx, contactID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_SetPrimaryContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("successful set primary", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:          contactID,
			UserID:      userID,
			ContactType: entity.ContactTypeEmail,
			IsPrimary:   false,
			IsActive:    true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)
		mockRepo.On("SetPrimary", ctx, userID, entity.ContactTypeEmail, contactID).Return(nil)

		err := uc.SetPrimaryContact(ctx, contactID, userID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized set primary", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		otherUserID := uuidv7.New()
		existingContact := &entity.UserContact{
			ID:          contactID,
			UserID:      otherUserID,
			ContactType: entity.ContactTypeEmail,
			IsActive:    true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)

		err := uc.SetPrimaryContact(ctx, contactID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_ToggleContactActive(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("toggle active to inactive", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsActive: true,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.IsActive == false
		})).Return(nil)

		err := uc.ToggleContactActive(ctx, contactID, userID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("toggle inactive to active", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		existingContact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsActive: false,
		}

		mockRepo.On("GetByID", ctx, contactID).Return(existingContact, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.IsActive == true
		})).Return(nil)

		err := uc.ToggleContactActive(ctx, contactID, userID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_GetContactsByType(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("get own contacts by type", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		expectedContacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, ContactType: entity.ContactTypeEmail, IsPublic: false, IsActive: true},
			{ID: uuidv7.New(), UserID: userID, ContactType: entity.ContactTypeEmail, IsPublic: true, IsActive: true},
		}

		mockRepo.On("GetUserContactsByType", ctx, userID, entity.ContactTypeEmail).Return(expectedContacts, nil)

		contacts, err := uc.GetUserContactsByType(ctx, userID, entity.ContactTypeEmail, userID)

		require.NoError(t, err)
		assert.Len(t, contacts, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("filter public contacts for other user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		allContacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: otherUserID, ContactType: entity.ContactTypeEmail, IsPublic: true},
			{ID: uuidv7.New(), UserID: otherUserID, ContactType: entity.ContactTypeEmail, IsPublic: false},
		}

		mockRepo.On("GetUserContactsByType", ctx, otherUserID, entity.ContactTypeEmail).Return(allContacts, nil)

		contacts, err := uc.GetUserContactsByType(ctx, otherUserID, entity.ContactTypeEmail, userID)

		require.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.True(t, contacts[0].IsPublic)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		_, err := uc.GetUserContactsByType(ctx, userID, entity.ContactType("invalid"), userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidContactType)
	})
}

func TestUserContactUseCase_GetPrimaryContact(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("get own primary contact", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		expectedContact := &entity.UserContact{
			ID:          contactID,
			UserID:      userID,
			ContactType: entity.ContactTypeEmail,
			IsPrimary:   true,
			IsPublic:    false,
		}

		mockRepo.On("GetPrimaryContact", ctx, userID, entity.ContactTypeEmail).Return(expectedContact, nil)

		contact, err := uc.GetPrimaryContact(ctx, userID, entity.ContactTypeEmail, userID)

		require.NoError(t, err)
		assert.Equal(t, contactID, contact.ID)
		assert.True(t, contact.IsPrimary)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get public primary contact of other user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		otherUserID := uuidv7.New()
		expectedContact := &entity.UserContact{
			ID:          contactID,
			UserID:      otherUserID,
			ContactType: entity.ContactTypeEmail,
			IsPrimary:   true,
			IsPublic:    true,
		}

		mockRepo.On("GetPrimaryContact", ctx, otherUserID, entity.ContactTypeEmail).Return(expectedContact, nil)

		contact, err := uc.GetPrimaryContact(ctx, otherUserID, entity.ContactTypeEmail, userID)

		require.NoError(t, err)
		assert.True(t, contact.IsPublic)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access to private primary contact", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		otherUserID := uuidv7.New()
		expectedContact := &entity.UserContact{
			ID:          contactID,
			UserID:      otherUserID,
			ContactType: entity.ContactTypeEmail,
			IsPrimary:   true,
			IsPublic:    false,
		}

		mockRepo.On("GetPrimaryContact", ctx, otherUserID, entity.ContactTypeEmail).Return(expectedContact, nil)

		_, err := uc.GetPrimaryContact(ctx, otherUserID, entity.ContactTypeEmail, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})

	t.Run("primary contact not found", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		mockRepo.On("GetPrimaryContact", ctx, userID, entity.ContactTypeEmail).Return(nil, entity.ErrNotFound)

		_, err := uc.GetPrimaryContact(ctx, userID, entity.ContactTypeEmail, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrContactNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserContactUseCase_VerifyContact(t *testing.T) {
	ctx := context.Background()
	contactID := uuidv7.New()

	t.Run("successful verification", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		mockRepo.On("VerifyContact", ctx, contactID).Return(nil)

		err := uc.VerifyContact(ctx, contactID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("verification fails", func(t *testing.T) {
		mockRepo := new(mocks.MockUserContactRepository)
		uc := NewUserContactUseCase(mockRepo)

		mockRepo.On("VerifyContact", ctx, contactID).Return(errors.New("database error"))

		err := uc.VerifyContact(ctx, contactID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
