package repository

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUserRepository defines the interface for user data access
type IUserRepository interface {
	// Basic CRUD
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)

	// Auth-specific operations
	UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.UserStatus) error
	UpdateLastLogin(ctx context.Context, id uuidv7.UUID, loginTime time.Time) error
	UpdatePassword(ctx context.Context, id uuidv7.UUID, hashedPassword string) error
	VerifyEmail(ctx context.Context, id uuidv7.UUID) error

	// Status management
	Suspend(ctx context.Context, id uuidv7.UUID, reason string, until *time.Time) error
	Ban(ctx context.Context, id uuidv7.UUID, reason string) error
	Reactivate(ctx context.Context, id uuidv7.UUID) error
	Deactivate(ctx context.Context, id uuidv7.UUID) error
}
