package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IProfileRepository defines the interface for Profile persistence
type IProfileRepository interface {
	// Create creates a new profile
	Create(ctx context.Context, profile *aggregate.Profile) error

	// GetByID retrieves a profile by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Profile, error)

	// GetByUserID retrieves a profile by user ID
	GetByUserID(ctx context.Context, userID uuidv7.UUID) (*aggregate.Profile, error)

	// Update updates a profile
	Update(ctx context.Context, profile *aggregate.Profile) error

	// Delete deletes a profile
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListPublicProfiles retrieves all public profiles (for discovery)
	ListPublicProfiles(ctx context.Context, limit, offset int) ([]*aggregate.Profile, error)

	// ExistsForUser checks if a profile exists for a user
	ExistsForUser(ctx context.Context, userID uuidv7.UUID) (bool, error)
}
