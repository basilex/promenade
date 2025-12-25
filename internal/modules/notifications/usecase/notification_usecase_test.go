package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/internal/modules/notifications/domain/repository/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNotificationUseCase_SendNotification_Success(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create user with default preferences
	userID := uuidv7.New()
	pref := entity.NewUserPreference(userID)
	err := prefRepo.Create(context.Background(), pref)
	require.NoError(t, err)

	// Send notification
	notification, err := useCase.SendNotification(
		context.Background(),
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"welcome_email",
		"Welcome!",
		"Welcome to our platform",
		nil,
	)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, userID, notification.UserID)
	assert.Equal(t, entity.NotificationTypeSystem, notification.Type)
	assert.Equal(t, entity.NotificationChannelEmail, notification.Channel)
	assert.Equal(t, entity.NotificationStatusPending, notification.Status)

	// Check event published
	assert.Equal(t, 1, eventBus.GetPublishedEventCount())
}

func TestNotificationUseCase_SendNotification_ChannelDisabled(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create user with email disabled
	userID := uuidv7.New()
	pref := entity.NewUserPreference(userID)
	pref.EmailEnabled = false
	err := prefRepo.Create(context.Background(), pref)
	require.NoError(t, err)

	// Try to send email notification
	notification, err := useCase.SendNotification(
		context.Background(),
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"welcome_email",
		"Welcome!",
		"Welcome to our platform",
		nil,
	)

	// Assertions
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrChannelDisabled))
	assert.Nil(t, notification)
	assert.Equal(t, 0, eventBus.GetPublishedEventCount())
}

func TestNotificationUseCase_SendNotification_TypeDisabled(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create user with marketing disabled
	userID := uuidv7.New()
	pref := entity.NewUserPreference(userID)
	pref.MarketingEnabled = false
	err := prefRepo.Create(context.Background(), pref)
	require.NoError(t, err)

	// Try to send marketing notification
	notification, err := useCase.SendNotification(
		context.Background(),
		userID,
		entity.NotificationTypeMarketing,
		entity.NotificationChannelEmail,
		"promo_email",
		"Special Offer!",
		"50% off today",
		nil,
	)

	// Assertions
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTypeDisabled))
	assert.Nil(t, notification)
}

func TestNotificationUseCase_GetNotification_Success(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create notification
	userID := uuidv7.New()
	notification := entity.NewNotification(
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"test_template",
		"Test",
		"Test content",
	)
	err := notifRepo.Create(context.Background(), notification)
	require.NoError(t, err)

	// Get notification
	result, err := useCase.GetNotification(context.Background(), notification.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, notification.ID, result.ID)
}

func TestNotificationUseCase_GetNotification_NotFound(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Try to get non-existent notification
	result, err := useCase.GetNotification(context.Background(), uuidv7.New())

	// Assertions
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotificationNotFound))
	assert.Nil(t, result)
}

func TestNotificationUseCase_MarkAsOpened(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create notification
	userID := uuidv7.New()
	notification := entity.NewNotification(
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"test_template",
		"Test",
		"Test content",
	)
	err := notifRepo.Create(context.Background(), notification)
	require.NoError(t, err)

	// Mark as opened
	err = useCase.MarkAsOpened(context.Background(), notification.ID)

	// Assertions
	assert.NoError(t, err)

	// Verify status changed
	result, err := notifRepo.GetByID(context.Background(), notification.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.NotificationStatusOpened, result.Status)
}

func TestNotificationUseCase_CreateDefaultPreferences(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create default preferences
	userID := uuidv7.New()
	prefs, err := useCase.CreateDefaultPreferences(context.Background(), userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)
	assert.True(t, prefs.EmailEnabled)
	assert.True(t, prefs.SystemEnabled)
	assert.False(t, prefs.MarketingEnabled)
}

func TestNotificationUseCase_GetUserNotifications(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create notifications for user
	userID := uuidv7.New()
	for i := 0; i < 5; i++ {
		notification := entity.NewNotification(
			userID,
			entity.NotificationTypeSystem,
			entity.NotificationChannelEmail,
			"test_template",
			"Test",
			"Test content",
		)
		err := notifRepo.Create(context.Background(), notification)
		require.NoError(t, err)
	}

	// Get notifications with pagination
	notifications, total, err := useCase.GetUserNotifications(context.Background(), userID, 10, 0)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, notifications, 5)
	assert.Equal(t, 5, total)
}

func TestNotificationUseCase_GetUnreadCount(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create notifications (some pending, some opened)
	userID := uuidv7.New()

	// 3 pending (unread)
	for i := 0; i < 3; i++ {
		notification := entity.NewNotification(
			userID,
			entity.NotificationTypeSystem,
			entity.NotificationChannelEmail,
			"test_template",
			"Test",
			"Test content",
		)
		err := notifRepo.Create(context.Background(), notification)
		require.NoError(t, err)
	}

	// 2 opened (read)
	for i := 0; i < 2; i++ {
		notification := entity.NewNotification(
			userID,
			entity.NotificationTypeSystem,
			entity.NotificationChannelEmail,
			"test_template",
			"Test",
			"Test content",
		)
		notification.MarkAsOpened()
		err := notifRepo.Create(context.Background(), notification)
		require.NoError(t, err)
	}

	// Get unread count
	count, err := useCase.GetUnreadCount(context.Background(), userID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestNotificationUseCase_MarkAsClicked(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create notification
	userID := uuidv7.New()
	notification := entity.NewNotification(
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"test_template",
		"Test",
		"Test content",
	)
	err := notifRepo.Create(context.Background(), notification)
	require.NoError(t, err)

	// Mark as clicked
	err = useCase.MarkAsClicked(context.Background(), notification.ID)

	// Assertions
	assert.NoError(t, err)

	// Verify status changed
	result, err := notifRepo.GetByID(context.Background(), notification.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.NotificationStatusClicked, result.Status)
}

func TestNotificationUseCase_UpdateUserPreferences(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create initial preferences
	userID := uuidv7.New()
	prefs := entity.NewUserPreference(userID)
	err := prefRepo.Create(context.Background(), prefs)
	require.NoError(t, err)

	// Update preferences
	prefs.MarketingEnabled = true
	prefs.SMSEnabled = true
	err = useCase.UpdateUserPreferences(context.Background(), userID, prefs)

	// Assertions
	assert.NoError(t, err)

	// Verify changes persisted
	updated, err := prefRepo.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.True(t, updated.MarketingEnabled)
	assert.True(t, updated.SMSEnabled)
}

func TestNotificationUseCase_GetUserPreferences(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create preferences
	userID := uuidv7.New()
	prefs := entity.NewUserPreference(userID)
	err := prefRepo.Create(context.Background(), prefs)
	require.NoError(t, err)

	// Get preferences
	result, err := useCase.GetUserPreferences(context.Background(), userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
}

func TestNotificationUseCase_GetUserPreferences_NotFound(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Try to get non-existent preferences
	result, err := useCase.GetUserPreferences(context.Background(), uuidv7.New())

	// Assertions
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPreferenceNotFound))
	assert.Nil(t, result)
}

func TestNotificationUseCase_SendNotification_CreatesDefaultPreferences(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// User without preferences
	userID := uuidv7.New()

	// Send notification
	notification, err := useCase.SendNotification(
		context.Background(),
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"welcome_email",
		"Welcome!",
		"Welcome to our platform",
		nil,
	)

	// Should create default preferences and send notification
	assert.NoError(t, err)
	assert.NotNil(t, notification)

	// Verify preferences were created
	prefs, err := prefRepo.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, prefs)
}

func TestNotificationUseCase_SendNotification_SystemBypassesQuietHours(t *testing.T) {
	// Setup mocks
	notifRepo := mocks.NewMockNotificationRepository()
	prefRepo := mocks.NewMockUserPreferenceRepository()
	eventBus := mocks.NewMockEventBus()

	useCase := NewNotificationUseCase(notifRepo, prefRepo, eventBus)

	// Create user with preferences
	userID := uuidv7.New()
	pref := entity.NewUserPreference(userID)
	err := prefRepo.Create(context.Background(), pref)
	require.NoError(t, err)

	// Send system notification (should always work)
	notification, err := useCase.SendNotification(
		context.Background(),
		userID,
		entity.NotificationTypeSystem,
		entity.NotificationChannelEmail,
		"system_alert",
		"System Alert",
		"Important system notification",
		nil,
	)

	// System notifications should work
	assert.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, entity.NotificationTypeSystem, notification.Type)
}
