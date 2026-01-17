package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IInteractionRepository defines the interface for interaction repository
type IInteractionRepository interface {
	// Create creates a new interaction
	Create(ctx context.Context, interaction *aggregate.Interaction) error

	// GetByID retrieves an interaction by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Interaction, error)

	// Update updates an existing interaction
	Update(ctx context.Context, interaction *aggregate.Interaction) error

	// Delete soft deletes an interaction
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListByCustomer lists all interactions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListByCompany lists all interactions for a company
	ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListByType lists all interactions of a specific type
	ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListByCreatedBy lists all interactions created by a user
	ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListPendingFollowUps lists all interactions with pending follow-ups
	ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*aggregate.Interaction, int64, error)
}
