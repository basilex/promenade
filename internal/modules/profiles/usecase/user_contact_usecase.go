package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	ErrContactNotFound      = errors.New("contact not found")
	ErrContactAlreadyExists = errors.New("contact already exists")
	ErrInvalidContactType   = errors.New("invalid contact type")
	ErrInvalidAvailability  = errors.New("invalid availability: 'from' time must be before 'to' time")
	ErrUnauthorized         = errors.New("unauthorized to access this contact")
)

// UserContactUseCase defines business logic for user contacts
type UserContactUseCase interface {
	// CreateContact creates a new contact for a user
	CreateContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, contactValue string,
		label *string, isPublic bool, availableFrom, availableTo *time.Time, availableDays []string, timezone *string, notes *string) (*entity.UserContact, error)

	// GetContact retrieves a contact by ID
	GetContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) (*entity.UserContact, error)

	// GetUserContacts retrieves all contacts for a user
	GetUserContacts(ctx context.Context, userID uuidv7.UUID, requestingUserID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error)

	// GetUserContactsByType retrieves user contacts filtered by type
	GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, requestingUserID uuidv7.UUID) ([]*entity.UserContact, error)

	// GetPrimaryContact retrieves the primary contact for a specific type
	GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, requestingUserID uuidv7.UUID) (*entity.UserContact, error)

	// GetPublicContacts retrieves all public contacts for a user
	GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error)

	// UpdateContact updates an existing contact
	UpdateContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID, contactType entity.ContactType,
		contactValue string, label *string, isPublic bool, availableFrom, availableTo *time.Time,
		availableDays []string, timezone *string, notes *string) (*entity.UserContact, error)

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error

	// SetPrimaryContact sets a contact as primary for its type
	SetPrimaryContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error

	// VerifyContact marks a contact as verified
	VerifyContact(ctx context.Context, contactID uuidv7.UUID) error

	// ToggleContactActive toggles the active status of a contact
	ToggleContactActive(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error
}

type userContactUseCase struct {
	contactRepo repository.UserContactRepository
}

// NewUserContactUseCase creates a new UserContactUseCase instance
func NewUserContactUseCase(contactRepo repository.UserContactRepository) UserContactUseCase {
	return &userContactUseCase{
		contactRepo: contactRepo,
	}
}

func (uc *userContactUseCase) CreateContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType,
	contactValue string, label *string, isPublic bool, availableFrom, availableTo *time.Time,
	availableDays []string, timezone *string, notes *string) (*entity.UserContact, error) {

	// Validate contact type
	if !contactType.IsValid() {
		return nil, ErrInvalidContactType
	}

	// Validate availability times
	if availableFrom != nil && availableTo != nil && !availableFrom.Before(*availableTo) {
		return nil, ErrInvalidAvailability
	}

	contact := &entity.UserContact{
		ID:            uuidv7.New(),
		UserID:        userID,
		ContactType:   contactType,
		ContactValue:  contactValue,
		Label:         label,
		IsVerified:    false, // New contacts are unverified by default
		IsPrimary:     false, // New contacts are not primary by default
		IsActive:      true,  // New contacts are active by default
		IsPublic:      isPublic,
		AvailableFrom: availableFrom,
		AvailableTo:   availableTo,
		AvailableDays: availableDays,
		Timezone:      timezone,
		Notes:         notes,
	}

	if err := contact.Validate(); err != nil {
		return nil, err
	}

	if err := uc.contactRepo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return contact, nil
}

func (uc *userContactUseCase) GetContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) (*entity.UserContact, error) {
	contact, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}

	// Check authorization: user can only see their own contacts or public contacts
	if contact.UserID != requestingUserID && !contact.IsPublic {
		return nil, ErrUnauthorized
	}

	return contact, nil
}

func (uc *userContactUseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID, requestingUserID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error) {
	// If requesting user is not the owner, only return public contacts
	if userID != requestingUserID {
		return uc.contactRepo.GetPublicContacts(ctx, userID)
	}

	return uc.contactRepo.GetUserContacts(ctx, userID, includeInactive)
}

func (uc *userContactUseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, requestingUserID uuidv7.UUID) ([]*entity.UserContact, error) {
	if !contactType.IsValid() {
		return nil, ErrInvalidContactType
	}

	contacts, err := uc.contactRepo.GetUserContactsByType(ctx, userID, contactType)
	if err != nil {
		return nil, err
	}

	// Filter out non-public contacts if requesting user is not the owner
	if userID != requestingUserID {
		publicContacts := make([]*entity.UserContact, 0)
		for _, contact := range contacts {
			if contact.IsPublic {
				publicContacts = append(publicContacts, contact)
			}
		}
		return publicContacts, nil
	}

	return contacts, nil
}

func (uc *userContactUseCase) GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, requestingUserID uuidv7.UUID) (*entity.UserContact, error) {
	if !contactType.IsValid() {
		return nil, ErrInvalidContactType
	}

	contact, err := uc.contactRepo.GetPrimaryContact(ctx, userID, contactType)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}

	// Check authorization
	if contact.UserID != requestingUserID && !contact.IsPublic {
		return nil, ErrUnauthorized
	}

	return contact, nil
}

func (uc *userContactUseCase) GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error) {
	return uc.contactRepo.GetPublicContacts(ctx, userID)
}

func (uc *userContactUseCase) UpdateContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID,
	contactType entity.ContactType, contactValue string, label *string, isPublic bool,
	availableFrom, availableTo *time.Time, availableDays []string, timezone *string, notes *string) (*entity.UserContact, error) {

	// Get existing contact to check ownership
	existing, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}

	// Check authorization
	if existing.UserID != requestingUserID {
		return nil, ErrUnauthorized
	}

	// Validate contact type
	if !contactType.IsValid() {
		return nil, ErrInvalidContactType
	}

	// Validate availability times
	if availableFrom != nil && availableTo != nil && !availableFrom.Before(*availableTo) {
		return nil, ErrInvalidAvailability
	}

	// Update fields
	existing.ContactType = contactType
	existing.ContactValue = contactValue
	existing.Label = label
	existing.IsPublic = isPublic
	existing.AvailableFrom = availableFrom
	existing.AvailableTo = availableTo
	existing.AvailableDays = availableDays
	existing.Timezone = timezone
	existing.Notes = notes

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := uc.contactRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (uc *userContactUseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error {
	// Get contact to check ownership
	contact, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	// Check authorization
	if contact.UserID != requestingUserID {
		return ErrUnauthorized
	}

	return uc.contactRepo.Delete(ctx, contactID)
}

func (uc *userContactUseCase) SetPrimaryContact(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error {
	// Get contact to check ownership
	contact, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	// Check authorization
	if contact.UserID != requestingUserID {
		return ErrUnauthorized
	}

	return uc.contactRepo.SetPrimary(ctx, contact.UserID, contact.ContactType, contactID)
}

func (uc *userContactUseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	// This would typically be called by an admin or verification system
	// For now, we'll just mark it as verified
	return uc.contactRepo.VerifyContact(ctx, contactID)
}

func (uc *userContactUseCase) ToggleContactActive(ctx context.Context, contactID uuidv7.UUID, requestingUserID uuidv7.UUID) error {
	// Get contact to check ownership
	contact, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	// Check authorization
	if contact.UserID != requestingUserID {
		return ErrUnauthorized
	}

	// Toggle active status
	contact.IsActive = !contact.IsActive

	return uc.contactRepo.Update(ctx, contact)
}
