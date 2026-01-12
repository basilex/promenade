package contact

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the business logic interface for Contact operations
type IUseCase interface {
	// CreateEmailContact creates a new email contact
	CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*Contact, error)

	// CreatePhoneContact creates a new phone contact
	CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*Contact, error)

	// CreateAddressContact creates a new address contact
	CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*Contact, error)

	// GetContact retrieves a contact by ID
	GetContact(ctx context.Context, contactID uuidv7.UUID) (*Contact, error)

	// GetUserContacts retrieves all contacts for a user
	GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*Contact, error)

	// GetUserContactsByType retrieves contacts for a user by type
	GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) ([]*Contact, error)

	// UpdateContact updates a contact
	UpdateContact(ctx context.Context, contact *Contact) error

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, contactID uuidv7.UUID) error

	// SetAsPrimary sets a contact as primary
	SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error

	// VerifyContact marks a contact as verified
	VerifyContact(ctx context.Context, contactID uuidv7.UUID) error

	// UpdateVisibility updates contact visibility
	UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error
}

// useCase implements IUseCase
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new useCase
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateEmailContact creates a new email contact
func (uc *useCase) CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*Contact, error) {
	contact, err := NewEmailContact(userID, email, label)
	if err != nil {
		return nil, err
	}

	if isPrimary {
		// Check if primary already exists
		exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypeEmail)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPrimaryContactExists
		}
		contact.SetAsPrimary()
	}

	if err := contact.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := uc.repo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	return contact, nil
}

// CreatePhoneContact creates a new phone contact
func (uc *useCase) CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*Contact, error) {
	contact, err := NewPhoneContact(userID, phone, label)
	if err != nil {
		return nil, err
	}

	if isPrimary {
		exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypePhone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPrimaryContactExists
		}
		contact.SetAsPrimary()
	}

	if err := contact.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := uc.repo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	return contact, nil
}

// CreateAddressContact creates a new address contact
func (uc *useCase) CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*Contact, error) {
	contact, err := NewAddressContact(userID, street, city, country, postalCode, label)
	if err != nil {
		return nil, err
	}

	if isPrimary {
		exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypeAddress)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPrimaryContactExists
		}
		contact.SetAsPrimary()
	}

	if err := contact.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return contact, nil
}

// GetContact retrieves a contact by ID
func (uc *useCase) GetContact(ctx context.Context, contactID uuidv7.UUID) (*Contact, error) {
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return nil, err
	}
	return contact, nil
}

// GetUserContacts retrieves all contacts for a user
func (uc *useCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*Contact, error) {
	contacts, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// GetUserContactsByType retrieves contacts for a user by type
func (uc *useCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) ([]*Contact, error) {
	contacts, err := uc.repo.GetByUserIDAndType(ctx, userID, contactType)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// UpdateContact updates a contact
func (uc *useCase) UpdateContact(ctx context.Context, contact *Contact) error {
	if err := contact.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, contact); err != nil {
		return err
	}

	return nil
}

// DeleteContact deletes a contact
func (uc *useCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, contactID); err != nil {
		return err
	}
	return nil
}

// SetAsPrimary sets a contact as primary
func (uc *useCase) SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error {
	// Get the contact first
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return err
	}

	// Use repository method that handles unsetting other primary contacts
	if err := uc.repo.SetPrimary(ctx, contactID); err != nil {
		return err
	}

	// Update the contact entity
	contact.SetAsPrimary()

	return nil
}

// VerifyContact marks a contact as verified
func (uc *useCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return err
	}

	contact.Verify()

	if err := uc.repo.Update(ctx, contact); err != nil {
		return err
	}

	return nil
}

// UpdateVisibility updates contact visibility
func (uc *useCase) UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error {
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return err
	}

	if isPublic {
		contact.MakePublic()
	} else {
		contact.MakePrivate()
	}

	if err := uc.repo.Update(ctx, contact); err != nil {
		return err
	}

	return nil
}
