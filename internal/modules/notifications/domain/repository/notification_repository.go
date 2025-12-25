package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// INotificationRepository defines notification data access operations
type INotificationRepository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *entity.Notification) error

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Notification, error)

	// GetByUserID retrieves notifications for a user with pagination
	GetByUserID(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Notification, error)

	// GetPendingNotifications retrieves pending notifications for processing
	GetPendingNotifications(ctx context.Context, limit int) ([]*entity.Notification, error)

	// Update updates an existing notification
	Update(ctx context.Context, notification *entity.Notification) error

	// Delete deletes a notification
	Delete(ctx context.Context, id uuidv7.UUID) error

	// CountByUserID counts total notifications for a user
	CountByUserID(ctx context.Context, userID uuidv7.UUID) (int, error)

	// GetUnreadCount counts unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuidv7.UUID) (int, error)
}
