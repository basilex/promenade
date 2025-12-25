package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/internal/modules/notifications/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// NotificationRepository implements INotificationRepository for PostgreSQL
type NotificationRepository struct {
	*BaseRepository
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sqlx.DB) repository.INotificationRepository {
	return &NotificationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	query := `
		INSERT INTO notifications_notifications (
			id, user_id, type, channel, status, template, subject, content, data, 
			sent_at, opened_at, clicked_at, failed_at, error_msg, created_at, updated_at
		) VALUES (
			:id, :user_id, :type, :channel, :status, :template, :subject, :content, :data,
			:sent_at, :opened_at, :clicked_at, :failed_at, :error_msg, :created_at, :updated_at
		)
	`
	_, err := r.NamedExec(ctx, query, notification)
	return err
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Notification, error) {
	var notification entity.Notification
	query := `SELECT * FROM notifications_notifications WHERE id = $1`
	err := r.Get(ctx, &notification, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &notification, err
}

// GetByUserID retrieves notifications for a user with pagination
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Notification, error) {
	var notifications []*entity.Notification
	query := `
		SELECT * FROM notifications_notifications 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &notifications, query, userID, limit, offset)
	return notifications, err
}

// GetPendingNotifications retrieves pending notifications for processing
func (r *NotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*entity.Notification, error) {
	var notifications []*entity.Notification
	query := `
		SELECT * FROM notifications_notifications 
		WHERE status = 'pending' 
		ORDER BY created_at ASC 
		LIMIT $1
	`
	err := r.Select(ctx, &notifications, query, limit)
	return notifications, err
}

// Update updates an existing notification
func (r *NotificationRepository) Update(ctx context.Context, notification *entity.Notification) error {
	query := `
		UPDATE notifications_notifications SET
			status = :status,
			sent_at = :sent_at,
			opened_at = :opened_at,
			clicked_at = :clicked_at,
			failed_at = :failed_at,
			error_msg = :error_msg,
			updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.NamedExec(ctx, query, notification)
	return err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM notifications_notifications WHERE id = $1`
	_, err := r.Exec(ctx, query, id)
	return err
}

// CountByUserID counts total notifications for a user
func (r *NotificationRepository) CountByUserID(ctx context.Context, userID uuidv7.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications_notifications WHERE user_id = $1`
	err := r.Get(ctx, &count, query, userID)
	return count, err
}

// GetUnreadCount counts unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuidv7.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM notifications_notifications 
		WHERE user_id = $1 AND status IN ('pending', 'sent', 'delivered')
	`
	err := r.Get(ctx, &count, query, userID)
	return count, err
}
