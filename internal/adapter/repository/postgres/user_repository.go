package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type userRepository struct {
	*BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, name, password, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Password,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, name, password, status, email_verified_at, 
		       suspended_reason, suspended_until, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user entity.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, name, password, status, email_verified_at, 
		       suspended_reason, suspended_until, last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user entity.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, status = $3, email_verified_at = $4,
		    suspended_reason = $5, suspended_until = $6, last_login_at = $7, updated_at = $8
		WHERE id = $9
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Name,
		user.Status,
		user.EmailVerifiedAt,
		user.SuspendedReason,
		user.SuspendedUntil,
		user.LastLoginAt,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
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

func (r *userRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
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

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT id, email, name, password, status, email_verified_at, 
		       suspended_reason, suspended_until, last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	var users []*entity.User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.UserStatus) error {
	query := `UPDATE users SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
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

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuidv7.UUID, loginTime time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, loginTime, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
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

func (r *userRepository) UpdatePassword(ctx context.Context, id uuidv7.UUID, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
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

func (r *userRepository) VerifyEmail(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE users 
		SET status = $1, email_verified_at = $2, updated_at = $3 
		WHERE id = $4
	`
	result, err := r.db.ExecContext(ctx, query, entity.UserStatusActive, time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
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

func (r *userRepository) Suspend(ctx context.Context, id uuidv7.UUID, reason string, until *time.Time) error {
	query := `
		UPDATE users 
		SET status = $1, suspended_reason = $2, suspended_until = $3, updated_at = $4 
		WHERE id = $5
	`
	result, err := r.db.ExecContext(ctx, query, entity.UserStatusSuspended, reason, until, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
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

func (r *userRepository) Ban(ctx context.Context, id uuidv7.UUID, reason string) error {
	query := `
		UPDATE users 
		SET status = $1, suspended_reason = $2, updated_at = $3 
		WHERE id = $4
	`
	result, err := r.db.ExecContext(ctx, query, entity.UserStatusBanned, reason, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
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

func (r *userRepository) Reactivate(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE users 
		SET status = $1, suspended_reason = NULL, suspended_until = NULL, updated_at = $2 
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, entity.UserStatusActive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
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

func (r *userRepository) Deactivate(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE users SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, entity.UserStatusInactive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
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
