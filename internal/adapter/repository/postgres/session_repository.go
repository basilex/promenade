package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type sessionRepository struct {
	*BaseRepository
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB) repository.SessionRepository {
	return &sessionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO user_sessions (id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM user_sessions
		WHERE id = $1
	`
	var session entity.Session
	err := r.Get(ctx, &session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get session by id: %w", err)
	}
	return &session, nil
}

func (r *sessionRepository) GetByRefreshToken(ctx context.Context, hashedToken string) (*entity.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM user_sessions
		WHERE refresh_token = $1 AND expires_at > NOW()
	`
	var session entity.Session
	err := r.Get(ctx, &session, query, hashedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}
	return &session, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *entity.Session) error {
	query := `
		UPDATE user_sessions
		SET refresh_token = $2, expires_at = $3
		WHERE id = $1
	`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, session.ID, session.RefreshToken, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM user_sessions WHERE id = $1`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *sessionRepository) GetUserSessions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	var sessions []*entity.Session
	err := r.Select(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	return sessions, nil
}

func (r *sessionRepository) CountUserSessions(ctx context.Context, userID uuidv7.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND expires_at > NOW()`
	var count int
	err := r.Get(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user sessions: %w", err)
	}
	return count, nil
}

// GetOldestSession retrieves the oldest active session for a user (by created_at)
func (r *sessionRepository) GetOldestSession(ctx context.Context, userID uuidv7.UUID) (*entity.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at ASC
		LIMIT 1
	`
	var session entity.Session
	err := r.Get(ctx, &session, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get oldest session: %w", err)
	}
	return &session, nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuidv7.UUID) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`
	err := r.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at <= NOW()`
	err := r.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}
