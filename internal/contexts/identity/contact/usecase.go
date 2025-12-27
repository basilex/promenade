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

// UseCase implements IUseCase
type UseCase struct {
	repo IRepository
}

// NewUseCase creates a new UseCase
func NewUseCase(repo IRepository) IUseCase {
	return &UseCase{
		repo: repo,
	}
}

// CreateEmailContact creates a new email contact
func (uc *UseCase) CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*Contact, error) {
	contact, err := NewEmailContact(userID, email, label)
	if err != nil {
		return nil, fmt.Errorf("failed to create email contact: %w", err)
	}

	if isPrimary {
		// Check if primary already exists
		exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypeEmail)
		if err != nil {
			return nil, fmt.Errorf("failed to check primary contact: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("primary email contact already exists for user")
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
func (uc *UseCase) CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*Contact, error) {
	contact, err := NewPhoneContact(userID, phone, label)
	if err != nil {
		return nil, fmt.Errorf("failed to create phone contact: %w", err)
	}

	if isPrimary {
		exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypePhone)
		if err != nil {
			return nil, fmt.Errorf("failed to check primary contact: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("primary phone contact already exists for user")
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
func (uc *UseCase) CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*Contact, error) {
	contact, err := NewAddressContact(userID, street, city, country, postalCode, label)
	if err != nil {
		return nil, fmt.Errorf("failed to create address contact: %w", err)
	}

	if isPrimary {
		exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to check primary contact: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("primary address contact already exists for user")
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

// GetContact retrieves a contact by ID
func (uc *UseCase) GetContact(ctx context.Context, contactID uuidv7.UUID) (*Contact, error) {
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	return contact, nil
}

// GetUserContacts retrieves all contacts for a user
func (uc *UseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*Contact, error) {
	contacts, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user contacts: %w", err)
	}
	return contacts, nil
}

// GetUserContactsByType retrieves contacts for a user by type
func (uc *UseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) ([]*Contact, error) {
	contacts, err := uc.repo.GetByUserIDAndType(ctx, userID, contactType)
	if err != nil {
		return nil, fmt.Errorf("failed to get user contacts by type: %w", err)
	}
	return contacts, nil
}

// UpdateContact updates a contact
func (uc *UseCase) UpdateContact(ctx context.Context, contact *Contact) error {
	if err := contact.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := uc.repo.Update(ctx, contact); err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	return nil
}

// DeleteContact deletes a contact
func (uc *UseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, contactID); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	return nil
}

// SetAsPrimary sets a contact as primary
func (uc *UseCase) SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error {
	// Get the contact first
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Use repository method that handles unsetting other primary contacts
	if err := uc.repo.SetPrimary(ctx, contactID); err != nil {
		return fmt.Errorf("failed to set primary: %w", err)
	}

	// Update the contact entity
	contact.SetAsPrimary()

	return nil
}

// VerifyContact marks a contact as verified
func (uc *UseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}

	contact.Verify()

	if err := uc.repo.Update(ctx, contact); err != nil {
		return fmt.Errorf("failed to verify contact: %w", err)
	}

	return nil
}

// UpdateVisibility updates contact visibility
func (uc *UseCase) UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error {
	contact, err := uc.repo.GetByID(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}

	if isPublic {
		contact.MakePublic()
	} else {
		contact.MakePrivate()
	}

	if err := uc.repo.Update(ctx, contact); err != nil {
		return fmt.Errorf("failed to update visibility: %w", err)
	}

	return nil
}
