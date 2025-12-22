package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserContactRepository defines the interface for user contact operations
type UserContactRepository interface {
	// Create creates a new user contact
	Create(ctx context.Context, contact *entity.UserContact) error

	// GetByID retrieves a contact by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserContact, error)

	// GetUserContacts retrieves all contacts for a user
	GetUserContacts(ctx context.Context, userID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error)

	// GetUserContactsByType retrieves user contacts filtered by type
	GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) ([]*entity.UserContact, error)

	// GetPrimaryContact retrieves the primary contact of a specific type for a user
	GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) (*entity.UserContact, error)

	// GetPublicContacts retrieves all public contacts for a user
	GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error)

	// Update updates an existing contact
	Update(ctx context.Context, contact *entity.UserContact) error

	// Delete deletes a contact
	Delete(ctx context.Context, id uuidv7.UUID) error

	// SetPrimary sets a contact as primary for its type (unsets others)
	SetPrimary(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, id uuidv7.UUID) error

	// VerifyContact marks a contact as verified
	VerifyContact(ctx context.Context, id uuidv7.UUID) error
}
