package usecase

import (
	"context"
	"fmt"

	contacterrors "github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/internal/contexts/identity/contact/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/contact/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IContactUseCase defines the business logic interface for Contact operations
type IContactUseCase interface {
	// CreateEmailContact creates a new email contact
	CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*aggregate.Contact, error)

	// CreatePhoneContact creates a new phone contact
	CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*aggregate.Contact, error)

	// CreateAddressContact creates a new address contact
	CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*aggregate.Contact, error)

	// GetContact retrieves a contact by ID
	GetContact(ctx context.Context, contactID uuidv7.UUID) (*aggregate.Contact, error)

	// GetUserContacts retrieves all contacts for a user
	GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Contact, error)

	// GetUserContactsByType retrieves contacts for a user by type
	GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType aggregate.ContactType) ([]*aggregate.Contact, error)

	// UpdateContact updates a contact
	UpdateContact(ctx context.Context, contact *aggregate.Contact) error

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, contactID uuidv7.UUID) error

	// SetAsPrimary sets a contact as primary
	SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error

	// VerifyContact marks a contact as verified
	VerifyContact(ctx context.Context, contactID uuidv7.UUID) error

	// UpdateVisibility updates contact visibility
	UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error
}

// ContactUseCase implements IContactUseCase
type ContactUseCase struct {
	repo repository.IContactRepository
}

// NewContactUseCase creates a new ContactUseCase
func NewContactUseCase(repo repository.IContactRepository) IContactUseCase {
	return &ContactUseCase{
		repo: repo,
	}
}

// CreateEmailContact creates a new email contact
func (u *ContactUseCase) CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*aggregate.Contact, error) {
	contact, err := aggregate.NewEmailContact(userID, email, label)
	if err != nil {
		return nil, err
	}

	if isPrimary {
		// Check if primary already exists
		exists, err := u.repo.ExistsPrimaryForUserAndType(ctx, userID, aggregate.ContactTypeEmail)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, contacterrors.ErrPrimaryContactExists
		}
		contact.SetAsPrimary()
	}

	if err := contact.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := u.repo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	return contact, nil
}

// CreatePhoneContact creates a new phone contact
func (u *ContactUseCase) CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*aggregate.Contact, error) {
	contact, err := aggregate.NewPhoneContact(userID, phone, label)
	if err != nil {
		return nil, err
	}

	if isPrimary {
		exists, err := u.repo.ExistsPrimaryForUserAndType(ctx, userID, aggregate.ContactTypePhone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, contacterrors.ErrPrimaryContactExists
		}
		contact.SetAsPrimary()
	}

	if err := contact.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := u.repo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to save contact: %w", err)
	}

	return contact, nil
}

// CreateAddressContact creates a new address contact
func (u *ContactUseCase) CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*aggregate.Contact, error) {
	contact, err := aggregate.NewAddressContact(userID, street, city, country, postalCode, label)
	if err != nil {
		return nil, err
	}

	if isPrimary {
		exists, err := u.repo.ExistsPrimaryForUserAndType(ctx, userID, aggregate.ContactTypeAddress)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, contacterrors.ErrPrimaryContactExists
		}
		contact.SetAsPrimary()
	}

	if err := contact.Validate(); err != nil {
		return nil, err
	}

	if err := u.repo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return contact, nil
}

// GetContact retrieves a contact by ID
func (u *ContactUseCase) GetContact(ctx context.Context, contactID uuidv7.UUID) (*aggregate.Contact, error) {
	contact, err := u.repo.GetByID(ctx, contactID)
	if err != nil {
		return nil, err
	}
	return contact, nil
}

// GetUserContacts retrieves all contacts for a user
func (u *ContactUseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Contact, error) {
	contacts, err := u.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// GetUserContactsByType retrieves contacts for a user by type
func (u *ContactUseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType aggregate.ContactType) ([]*aggregate.Contact, error) {
	contacts, err := u.repo.GetByUserIDAndType(ctx, userID, contactType)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// UpdateContact updates a contact
func (u *ContactUseCase) UpdateContact(ctx context.Context, contact *aggregate.Contact) error {
	if err := contact.Validate(); err != nil {
		return err
	}

	if err := u.repo.Update(ctx, contact); err != nil {
		return err
	}

	return nil
}

// DeleteContact deletes a contact
func (u *ContactUseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID) error {
	if err := u.repo.Delete(ctx, contactID); err != nil {
		return err
	}
	return nil
}

// SetAsPrimary sets a contact as primary
func (u *ContactUseCase) SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error {
	// Get the contact first
	contact, err := u.repo.GetByID(ctx, contactID)
	if err != nil {
		return err
	}

	// Use repository method that handles unsetting other primary contacts
	if err := u.repo.SetPrimary(ctx, contactID); err != nil {
		return err
	}

	// Update the contact entity
	contact.SetAsPrimary()

	return nil
}

// VerifyContact marks a contact as verified
func (u *ContactUseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	contact, err := u.repo.GetByID(ctx, contactID)
	if err != nil {
		return err
	}

	contact.Verify()

	if err := u.repo.Update(ctx, contact); err != nil {
		return err
	}

	return nil
}

// UpdateVisibility updates contact visibility
func (u *ContactUseCase) UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error {
	contact, err := u.repo.GetByID(ctx, contactID)
	if err != nil {
		return err
	}

	if isPublic {
		contact.MakePublic()
	} else {
		contact.MakePrivate()
	}

	if err := u.repo.Update(ctx, contact); err != nil {
		return err
	}

	return nil
}
