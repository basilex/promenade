package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ISessionRepository defines the interface for session data access
type ISessionRepository interface {
	// Basic CRUD
	Create(ctx context.Context, session *entity.Session) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Session, error)
	GetByRefreshToken(ctx context.Context, hashedToken string) (*entity.Session, error)
	Update(ctx context.Context, session *entity.Session) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// User session management
	GetUserSessions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Session, error)
	CountUserSessions(ctx context.Context, userID uuidv7.UUID) (int, error)
	GetOldestSession(ctx context.Context, userID uuidv7.UUID) (*entity.Session, error)
	DeleteByUserID(ctx context.Context, userID uuidv7.UUID) error
	DeleteExpired(ctx context.Context) error
}
