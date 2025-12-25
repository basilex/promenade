package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUserPreferenceRepository defines user preference data access operations
type IUserPreferenceRepository interface {
	// Create creates new user preferences
	Create(ctx context.Context, preference *entity.UserPreference) error

	// GetByUserID retrieves user preferences by user ID
	GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error)

	// Update updates existing user preferences
	Update(ctx context.Context, preference *entity.UserPreference) error

	// Delete deletes user preferences
	Delete(ctx context.Context, userID uuidv7.UUID) error

	// Exists checks if preferences exist for a user
	Exists(ctx context.Context, userID uuidv7.UUID) (bool, error)
}
