package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/internal/modules/notifications/domain/repository"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	// ErrNotificationNotFound is returned when notification is not found
	ErrNotificationNotFound = errors.New("notification not found")

	// ErrPreferenceNotFound is returned when user preference is not found
	ErrPreferenceNotFound = errors.New("user preference not found")

	// ErrInvalidNotification is returned when notification data is invalid
	ErrInvalidNotification = errors.New("invalid notification data")

	// ErrChannelDisabled is returned when notification channel is disabled
	ErrChannelDisabled = errors.New("notification channel is disabled for user")

	// ErrTypeDisabled is returned when notification type is disabled
	ErrTypeDisabled = errors.New("notification type is disabled for user")

	// ErrQuietHours is returned when notification is blocked by quiet hours
	ErrQuietHours = errors.New("notification blocked by quiet hours")
)

// INotificationUseCase defines notification business logic operations
type INotificationUseCase interface {
	// SendNotification creates and sends a notification
	SendNotification(ctx context.Context, userID uuidv7.UUID, notifType entity.NotificationType, channel entity.NotificationChannel, template, subject, content string, data map[string]any) (*entity.Notification, error)

	// GetNotification retrieves a notification by ID
	GetNotification(ctx context.Context, notificationID uuidv7.UUID) (*entity.Notification, error)

	// GetUserNotifications retrieves notifications for a user with pagination
	GetUserNotifications(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Notification, int, error)

	// GetUnreadCount gets count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuidv7.UUID) (int, error)

	// MarkAsOpened marks notification as opened
	MarkAsOpened(ctx context.Context, notificationID uuidv7.UUID) error

	// MarkAsClicked marks notification as clicked
	MarkAsClicked(ctx context.Context, notificationID uuidv7.UUID) error

	// GetUserPreferences retrieves user notification preferences
	GetUserPreferences(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error)

	// UpdateUserPreferences updates user notification preferences
	UpdateUserPreferences(ctx context.Context, userID uuidv7.UUID, preference *entity.UserPreference) error

	// CreateDefaultPreferences creates default preferences for a new user
	CreateDefaultPreferences(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error)
}

type notificationUseCase struct {
	notifRepo repository.INotificationRepository
	prefRepo  repository.IUserPreferenceRepository
	eventBus  bus.IBus
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase(
	notifRepo repository.INotificationRepository,
	prefRepo repository.IUserPreferenceRepository,
	eventBus bus.IBus,
) INotificationUseCase {
	return &notificationUseCase{
		notifRepo: notifRepo,
		prefRepo:  prefRepo,
		eventBus:  eventBus,
	}
}

// SendNotification creates and sends a notification
func (uc *notificationUseCase) SendNotification(
	ctx context.Context,
	userID uuidv7.UUID,
	notifType entity.NotificationType,
	channel entity.NotificationChannel,
	template, subject, content string,
	data map[string]any,
) (*entity.Notification, error) {
	log := logger.FromContext(ctx)

	// Check user preferences
	prefs, err := uc.prefRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Create default preferences if not exist
	if prefs == nil {
		prefs = entity.NewUserPreference(userID)
		if err := uc.prefRepo.Create(ctx, prefs); err != nil {
			log.Error("Failed to create default preferences", "error", err)
		}
	}

	// Check if channel is enabled
	if !prefs.IsChannelEnabled(channel) {
		return nil, ErrChannelDisabled
	}

	// Check if notification type is enabled
	if !prefs.IsTypeEnabled(notifType) {
		return nil, ErrTypeDisabled
	}

	// Check quiet hours (skip for system and security notifications)
	if notifType != entity.NotificationTypeSystem && notifType != entity.NotificationTypeSecurity {
		if prefs.IsInQuietHours(time.Now()) {
			return nil, ErrQuietHours
		}
	}

	// Create notification
	notification := entity.NewNotification(userID, notifType, channel, template, subject, content)
	notification.Data = data

	if err := notification.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidNotification, err)
	}

	// Save notification
	if err := uc.notifRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Publish event for async processing
	event := &NotificationRequestedEvent{
		BaseEvent:      bus.NewBaseEvent("notification.requested", notification.ID),
		NotificationID: notification.ID,
		UserID:         userID,
		NotifType:      string(notifType),
		Channel:        string(channel),
	}
	if err := uc.eventBus.Publish(ctx, "notification.requested", event); err != nil {
		log.Error("Failed to publish notification.requested event", "error", err)
	}

	log.Info("Notification created", "id", notification.ID, "user_id", userID, "type", notifType, "channel", channel)

	return notification, nil
}

// GetNotification retrieves a notification by ID
func (uc *notificationUseCase) GetNotification(ctx context.Context, notificationID uuidv7.UUID) (*entity.Notification, error) {
	notification, err := uc.notifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	if notification == nil {
		return nil, ErrNotificationNotFound
	}
	return notification, nil
}

// GetUserNotifications retrieves notifications for a user with pagination
func (uc *notificationUseCase) GetUserNotifications(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	notifications, err := uc.notifRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user notifications: %w", err)
	}

	total, err := uc.notifRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user notifications: %w", err)
	}

	return notifications, total, nil
}

// GetUnreadCount gets count of unread notifications for a user
func (uc *notificationUseCase) GetUnreadCount(ctx context.Context, userID uuidv7.UUID) (int, error) {
	count, err := uc.notifRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// MarkAsOpened marks notification as opened
func (uc *notificationUseCase) MarkAsOpened(ctx context.Context, notificationID uuidv7.UUID) error {
	notification, err := uc.notifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}
	if notification == nil {
		return ErrNotificationNotFound
	}

	notification.MarkAsOpened()
	if err := uc.notifRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}

// MarkAsClicked marks notification as clicked
func (uc *notificationUseCase) MarkAsClicked(ctx context.Context, notificationID uuidv7.UUID) error {
	notification, err := uc.notifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}
	if notification == nil {
		return ErrNotificationNotFound
	}

	notification.MarkAsClicked()
	if err := uc.notifRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}

// GetUserPreferences retrieves user notification preferences
func (uc *notificationUseCase) GetUserPreferences(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error) {
	prefs, err := uc.prefRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}
	if prefs == nil {
		return nil, ErrPreferenceNotFound
	}
	return prefs, nil
}

// UpdateUserPreferences updates user notification preferences
func (uc *notificationUseCase) UpdateUserPreferences(ctx context.Context, userID uuidv7.UUID, preference *entity.UserPreference) error {
	// Ensure user_id matches
	preference.UserID = userID
	preference.UpdatedAt = time.Now()

	if err := preference.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidNotification, err)
	}

	// Check if preferences exist
	exists, err := uc.prefRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check preferences existence: %w", err)
	}

	if !exists {
		return ErrPreferenceNotFound
	}

	if err := uc.prefRepo.Update(ctx, preference); err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// CreateDefaultPreferences creates default preferences for a new user
func (uc *notificationUseCase) CreateDefaultPreferences(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error) {
	// Check if preferences already exist
	exists, err := uc.prefRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check preferences existence: %w", err)
	}

	if exists {
		return uc.prefRepo.GetByUserID(ctx, userID)
	}

	// Create default preferences
	prefs := entity.NewUserPreference(userID)
	if err := uc.prefRepo.Create(ctx, prefs); err != nil {
		return nil, fmt.Errorf("failed to create default preferences: %w", err)
	}

	return prefs, nil
}
