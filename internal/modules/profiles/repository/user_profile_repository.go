package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserProfileRepository defines the interface for user profile data access
type UserProfileRepository interface {
	// Create creates a new user profile
	Create(ctx context.Context, profile *entity.UserProfile) error

	// GetByID retrieves a profile by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserProfile, error)

	// GetByUserID retrieves a profile by user ID (one-to-one relationship)
	GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserProfile, error)

	// GetByNickname retrieves a profile by nickname
	GetByNickname(ctx context.Context, nickname string) (*entity.UserProfile, error)

	// Update updates an existing profile
	Update(ctx context.Context, profile *entity.UserProfile) error

	// Delete deletes a profile by ID
	Delete(ctx context.Context, id uuidv7.UUID) error

	// List retrieves profiles with optional filters
	List(ctx context.Context, limit, offset int, isPublic *bool) ([]*entity.UserProfile, error)

	// UpdateLastSeen updates the last seen timestamp
	UpdateLastSeen(ctx context.Context, profileID uuidv7.UUID) error

	// IncrementProfileViews increments profile view counter
	IncrementProfileViews(ctx context.Context, profileID uuidv7.UUID) error

	// Ban bans a profile
	Ban(ctx context.Context, profileID uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error

	// Unban removes ban from profile
	Unban(ctx context.Context, profileID uuidv7.UUID) error

	// SetVerified sets profile verification status
	SetVerified(ctx context.Context, profileID uuidv7.UUID, verified bool) error

	// Search searches profiles by query (nickname, display name, bio)
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error)
}
