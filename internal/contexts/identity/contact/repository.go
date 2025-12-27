package contact

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for Contact persistence
type IRepository interface {
	// Create creates a new contact
	Create(ctx context.Context, contact *Contact) error

	// GetByID retrieves a contact by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error)

	// GetByUserID retrieves all contacts for a user
	GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*Contact, error)

	// GetByUserIDAndType retrieves contacts for a user by type
	GetByUserIDAndType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) ([]*Contact, error)

	// GetPrimaryByUserIDAndType retrieves the primary contact for a user by type
	GetPrimaryByUserIDAndType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) (*Contact, error)

	// Update updates a contact
	Update(ctx context.Context, contact *Contact) error

	// Delete deletes a contact
	Delete(ctx context.Context, id uuidv7.UUID) error

	// SetPrimary sets a contact as primary and unsets all other primary contacts of the same type
	SetPrimary(ctx context.Context, id uuidv7.UUID) error

	// ExistsPrimaryForUserAndType checks if a primary contact exists for user and type
	ExistsPrimaryForUserAndType(ctx context.Context, userID uuidv7.UUID, contactType ContactType) (bool, error)
}
