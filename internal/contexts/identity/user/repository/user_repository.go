package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/identity/user/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUserRepository defines the interface for User repository
type IUserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *aggregate.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*aggregate.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *aggregate.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ListUsers lists users with pagination
	ListUsers(ctx context.Context, limit, offset int) ([]*aggregate.User, int, error)
}
