package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/internal/modules/notifications/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserPreferenceRepository implements IUserPreferenceRepository for PostgreSQL
type UserPreferenceRepository struct {
	*BaseRepository
}

// NewUserPreferenceRepository creates a new user preference repository
func NewUserPreferenceRepository(db *sqlx.DB) repository.IUserPreferenceRepository {
	return &UserPreferenceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates new user preferences
func (r *UserPreferenceRepository) Create(ctx context.Context, preference *entity.UserPreference) error {
	query := `
		INSERT INTO notifications_user_preferences (
			id, user_id, email_enabled, sms_enabled, push_enabled, in_app_enabled,
			system_enabled, security_enabled, marketing_enabled, product_enabled, social_enabled,
			quiet_hours_start, quiet_hours_end, timezone, created_at, updated_at
		) VALUES (
			:id, :user_id, :email_enabled, :sms_enabled, :push_enabled, :in_app_enabled,
			:system_enabled, :security_enabled, :marketing_enabled, :product_enabled, :social_enabled,
			:quiet_hours_start, :quiet_hours_end, :timezone, :created_at, :updated_at
		)
	`
	_, err := r.NamedExec(ctx, query, preference)
	return err
}

// GetByUserID retrieves user preferences by user ID
func (r *UserPreferenceRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error) {
	var preference entity.UserPreference
	query := `SELECT * FROM notifications_user_preferences WHERE user_id = $1`
	err := r.Get(ctx, &preference, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &preference, err
}

// Update updates existing user preferences
func (r *UserPreferenceRepository) Update(ctx context.Context, preference *entity.UserPreference) error {
	query := `
		UPDATE notifications_user_preferences SET
			email_enabled = :email_enabled,
			sms_enabled = :sms_enabled,
			push_enabled = :push_enabled,
			in_app_enabled = :in_app_enabled,
			system_enabled = :system_enabled,
			security_enabled = :security_enabled,
			marketing_enabled = :marketing_enabled,
			product_enabled = :product_enabled,
			social_enabled = :social_enabled,
			quiet_hours_start = :quiet_hours_start,
			quiet_hours_end = :quiet_hours_end,
			timezone = :timezone,
			updated_at = :updated_at
		WHERE user_id = :user_id
	`
	_, err := r.NamedExec(ctx, query, preference)
	return err
}

// Delete deletes user preferences
func (r *UserPreferenceRepository) Delete(ctx context.Context, userID uuidv7.UUID) error {
	query := `DELETE FROM notifications_user_preferences WHERE user_id = $1`
	_, err := r.Exec(ctx, query, userID)
	return err
}

// Exists checks if preferences exist for a user
func (r *UserPreferenceRepository) Exists(ctx context.Context, userID uuidv7.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications_user_preferences WHERE user_id = $1`
	err := r.Get(ctx, &count, query, userID)
	return count > 0, err
}
