package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/repository/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

//  Module-independent test: imports only module and pkg, no core dependencies

func TestCreateContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	userID := uuidv7.New()
	contactType := entity.ContactTypeEmail
	contactValue := "test@example.com"

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.UserID == userID && c.ContactType == contactType
		})).Return(nil).Once()

		result, err := uc.CreateContact(context.Background(), userID, contactType, contactValue, nil, true, nil, nil, nil, nil, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, contactType, result.ContactType)
		assert.Equal(t, contactValue, result.ContactValue)
		assert.True(t, result.IsActive)
		assert.False(t, result.IsVerified)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		invalidType := entity.ContactType("invalid")

		result, err := uc.CreateContact(context.Background(), userID, invalidType, contactValue, nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidContactType, err)
		assert.Nil(t, result)
	})

	t.Run("invalid availability", func(t *testing.T) {
		from := time.Now().Add(2 * time.Hour)
		to := time.Now()

		result, err := uc.CreateContact(context.Background(), userID, contactType, contactValue, nil, true, &from, &to, nil, nil, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAvailability, err)
		assert.Nil(t, result)
	})
}

func TestGetContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	contactID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("success - own contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsPublic: false,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		result, err := uc.GetContact(context.Background(), contactID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, contactID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - public contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsPublic: true,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		result, err := uc.GetContact(context.Background(), contactID, otherUserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - private contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsPublic: false,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		result, err := uc.GetContact(context.Background(), contactID, otherUserID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, contactID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.GetContact(context.Background(), contactID, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrContactNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserContacts(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("success - own contacts", func(t *testing.T) {
		contacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, IsPublic: false},
			{ID: uuidv7.New(), UserID: userID, IsPublic: true},
		}
		mockRepo.On("GetUserContacts", mock.Anything, userID, false).Return(contacts, nil).Once()

		result, err := uc.GetUserContacts(context.Background(), userID, userID, false)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - other user contacts (public only)", func(t *testing.T) {
		contacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, IsPublic: true},
		}
		mockRepo.On("GetPublicContacts", mock.Anything, userID).Return(contacts, nil).Once()

		result, err := uc.GetUserContacts(context.Background(), userID, otherUserID, false)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserContactsByType(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactType := entity.ContactTypeEmail

	t.Run("success - own contacts", func(t *testing.T) {
		contacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, ContactType: contactType, IsPublic: false},
		}
		mockRepo.On("GetUserContactsByType", mock.Anything, userID, contactType).Return(contacts, nil).Once()

		result, err := uc.GetUserContactsByType(context.Background(), userID, contactType, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - other user contacts (public only)", func(t *testing.T) {
		contacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, ContactType: contactType, IsPublic: false},
			{ID: uuidv7.New(), UserID: userID, ContactType: contactType, IsPublic: true},
		}
		mockRepo.On("GetUserContactsByType", mock.Anything, userID, contactType).Return(contacts, nil).Once()

		result, err := uc.GetUserContactsByType(context.Background(), userID, contactType, otherUserID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.True(t, result[0].IsPublic)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		invalidType := entity.ContactType("invalid")

		result, err := uc.GetUserContactsByType(context.Background(), userID, invalidType, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidContactType, err)
		assert.Nil(t, result)
	})
}

func TestGetPrimaryContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactType := entity.ContactTypeEmail

	t.Run("success - own contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:        uuidv7.New(),
			UserID:    userID,
			IsPrimary: true,
			IsPublic:  false,
		}
		mockRepo.On("GetPrimaryContact", mock.Anything, userID, contactType).Return(contact, nil).Once()

		result, err := uc.GetPrimaryContact(context.Background(), userID, contactType, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - private contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:        uuidv7.New(),
			UserID:    userID,
			IsPrimary: true,
			IsPublic:  false,
		}
		mockRepo.On("GetPrimaryContact", mock.Anything, userID, contactType).Return(contact, nil).Once()

		result, err := uc.GetPrimaryContact(context.Background(), userID, contactType, otherUserID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetPrimaryContact", mock.Anything, userID, contactType).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.GetPrimaryContact(context.Background(), userID, contactType, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrContactNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		invalidType := entity.ContactType("invalid")

		result, err := uc.GetPrimaryContact(context.Background(), userID, invalidType, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidContactType, err)
		assert.Nil(t, result)
	})
}

func TestUpdateContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	contactID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactType := entity.ContactTypeEmail
	newValue := "new@example.com"

	t.Run("success", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.ID == contactID && c.ContactValue == newValue
		})).Return(nil).Once()

		result, err := uc.UpdateContact(context.Background(), contactID, userID, contactType, newValue, nil, true, nil, nil, nil, nil, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newValue, result.ContactValue)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, contactID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.UpdateContact(context.Background(), contactID, userID, contactType, newValue, nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrContactNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		result, err := uc.UpdateContact(context.Background(), contactID, otherUserID, contactType, newValue, nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		invalidType := entity.ContactType("invalid")
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		result, err := uc.UpdateContact(context.Background(), contactID, userID, invalidType, newValue, nil, true, nil, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidContactType, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	contactID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()
		mockRepo.On("Delete", mock.Anything, contactID).Return(nil).Once()

		err := uc.DeleteContact(context.Background(), contactID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, contactID).Return(nil, entity.ErrNotFound).Once()

		err := uc.DeleteContact(context.Background(), contactID, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrContactNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		err := uc.DeleteContact(context.Background(), contactID, otherUserID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSetPrimaryContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	contactID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	contactType := entity.ContactTypeEmail

	t.Run("success", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:          contactID,
			UserID:      userID,
			ContactType: contactType,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()
		mockRepo.On("SetPrimary", mock.Anything, userID, contactType, contactID).Return(nil).Once()

		err := uc.SetPrimaryContact(context.Background(), contactID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, contactID).Return(nil, entity.ErrNotFound).Once()

		err := uc.SetPrimaryContact(context.Background(), contactID, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrContactNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		err := uc.SetPrimaryContact(context.Background(), contactID, otherUserID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestVerifyContact(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	contactID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("VerifyContact", mock.Anything, contactID).Return(nil).Once()

		err := uc.VerifyContact(context.Background(), contactID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		dbErr := errors.New("db error")
		mockRepo.On("VerifyContact", mock.Anything, contactID).Return(dbErr).Once()

		err := uc.VerifyContact(context.Background(), contactID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestToggleContactActive(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	contactID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("success - activate", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsActive: false,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.ID == contactID && c.IsActive == true
		})).Return(nil).Once()

		err := uc.ToggleContactActive(context.Background(), contactID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - deactivate", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:       contactID,
			UserID:   userID,
			IsActive: true,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.UserContact) bool {
			return c.ID == contactID && c.IsActive == false
		})).Return(nil).Once()

		err := uc.ToggleContactActive(context.Background(), contactID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, contactID).Return(nil, entity.ErrNotFound).Once()

		err := uc.ToggleContactActive(context.Background(), contactID, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrContactNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:     contactID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, contactID).Return(contact, nil).Once()

		err := uc.ToggleContactActive(context.Background(), contactID, otherUserID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPublicContacts(t *testing.T) {
	mockRepo := new(mocks.MockUserContactRepository)
	uc := NewUserContactUseCase(mockRepo)

	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		contacts := []*entity.UserContact{
			{ID: uuidv7.New(), UserID: userID, IsPublic: true},
			{ID: uuidv7.New(), UserID: userID, IsPublic: true},
		}
		mockRepo.On("GetPublicContacts", mock.Anything, userID).Return(contacts, nil).Once()

		result, err := uc.GetPublicContacts(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})
}
