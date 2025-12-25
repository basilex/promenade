//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/notifications/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/test/integration"
)

func TestNotificationRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	repo := postgres.NewNotificationRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user
	user := fixtures.CreateUser(t, "notification-user@example.com", "password123")

	t.Run("Create and GetByID", func(t *testing.T) {
		notification := entity.NewNotification(
			user.ID,
			entity.NotificationTypeSystem,
			entity.NotificationChannelEmail,
			"test_template",
			"Test Subject",
			"Test Content",
		)

		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, notification.UserID, retrieved.UserID)
		assert.Equal(t, notification.Type, retrieved.Type)
		assert.Equal(t, notification.Channel, retrieved.Channel)
		assert.Equal(t, notification.Subject, retrieved.Subject)
		assert.Equal(t, entity.NotificationStatusPending, retrieved.Status)
	})

	t.Run("GetByUserID returns user notifications", func(t *testing.T) {
		// Create additional notifications for the user
		for i := 0; i < 3; i++ {
			notification := entity.NewNotification(
				user.ID,
				entity.NotificationTypeProduct,
				entity.NotificationChannelInApp,
				"template",
				"Subject",
				"Content",
			)
			require.NoError(t, repo.Create(ctx, notification))
		}

		results, err := repo.GetByUserID(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3)
	})

	t.Run("Update notification status", func(t *testing.T) {
		notification := entity.NewNotification(
			user.ID,
			entity.NotificationTypeSystem,
			entity.NotificationChannelEmail,
			"template",
			"Subject",
			"Content",
		)
		require.NoError(t, repo.Create(ctx, notification))

		notification.MarkAsSent()
		err := repo.Update(ctx, notification)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.NotificationStatusSent, retrieved.Status)
		assert.NotNil(t, retrieved.SentAt)
	})

	t.Run("CountByUserID returns correct count", func(t *testing.T) {
		count, err := repo.CountByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 4)
	})

	t.Run("GetUnreadCount returns pending notifications", func(t *testing.T) {
		count, err := repo.GetUnreadCount(ctx, user.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("GetPendingNotifications returns pending only", func(t *testing.T) {
		results, err := repo.GetPendingNotifications(ctx, 10)
		require.NoError(t, err)
		for _, n := range results {
			assert.Equal(t, entity.NotificationStatusPending, n.Status)
		}
	})

	t.Run("Delete notification", func(t *testing.T) {
		notification := entity.NewNotification(
			user.ID,
			entity.NotificationTypeSystem,
			entity.NotificationChannelEmail,
			"template",
			"Subject",
			"Content",
		)
		require.NoError(t, repo.Create(ctx, notification))

		err := repo.Delete(ctx, notification.ID)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Nil(t, result) // Should not exist
	})
}

func TestUserPreferenceRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	repo := postgres.NewUserPreferenceRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user
	user := fixtures.CreateUser(t, "preference-user@example.com", "password123")

	t.Run("Create and GetByUserID", func(t *testing.T) {
		preference := entity.NewUserPreference(user.ID)

		err := repo.Create(ctx, preference)
		require.NoError(t, err)

		retrieved, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.UserID)
		assert.True(t, retrieved.EmailEnabled)
		assert.True(t, retrieved.SystemEnabled)
		assert.False(t, retrieved.MarketingEnabled)
	})

	t.Run("Update user preference", func(t *testing.T) {
		preference, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)

		preference.SMSEnabled = true
		preference.MarketingEnabled = true
		preference.Timezone = "Europe/Kiev"
		preference.UpdatedAt = time.Now()

		err = repo.Update(ctx, preference)
		require.NoError(t, err)

		retrieved, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.SMSEnabled)
		assert.True(t, retrieved.MarketingEnabled)
		assert.Equal(t, "Europe/Kiev", retrieved.Timezone)
	})

	t.Run("Exists returns true for existing preference", func(t *testing.T) {
		exists, err := repo.Exists(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Delete user preference", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		result, err := repo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, result) // Should not exist
	})

	t.Run("Exists returns false for deleted preference", func(t *testing.T) {
		exists, err := repo.Exists(ctx, user.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestNotificationWorkflow_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	prefRepo := postgres.NewUserPreferenceRepository(testDB.DB)
	notifRepo := postgres.NewNotificationRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user
	user := fixtures.CreateUser(t, "workflow-user@example.com", "password123")

	t.Run("Complete workflow: preferences → notification → tracking", func(t *testing.T) {
		// 1. Create user preferences
		prefs := entity.NewUserPreference(user.ID)
		err := prefRepo.Create(ctx, prefs)
		require.NoError(t, err)

		// 2. Verify email is enabled
		assert.True(t, prefs.EmailEnabled)

		// 3. Create notification
		notification := entity.NewNotification(
			user.ID,
			entity.NotificationTypeProduct,
			entity.NotificationChannelEmail,
			"product_update",
			"New Feature Available",
			"Check out our latest feature!",
		)
		notification.Data = map[string]any{
			"feature_id": "feature-123",
			"url":        "https://app.promenade.com/features/123",
		}
		err = notifRepo.Create(ctx, notification)
		require.NoError(t, err)

		// 4. Mark as sent
		notification.MarkAsSent()
		err = notifRepo.Update(ctx, notification)
		require.NoError(t, err)

		// 5. Mark as delivered
		notification.MarkAsDelivered()
		err = notifRepo.Update(ctx, notification)
		require.NoError(t, err)

		// 6. Mark as opened
		notification.MarkAsOpened()
		err = notifRepo.Update(ctx, notification)
		require.NoError(t, err)

		// 7. Verify final state
		result, err := notifRepo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.NotificationStatusOpened, result.Status)
		assert.NotNil(t, result.SentAt)
		assert.NotNil(t, result.OpenedAt)
		assert.True(t, result.IsDelivered())

		// 8. Verify unread count decreased (should be 0 since we opened it)
		count, err := notifRepo.GetUnreadCount(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestQuietHoursLogic_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	prefRepo := postgres.NewUserPreferenceRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user
	user := fixtures.CreateUser(t, "quiethours-user@example.com", "password123")

	t.Run("Quiet hours logic with timezone", func(t *testing.T) {
		// Create preferences with quiet hours
		prefs := entity.NewUserPreference(user.ID)

		// Set quiet hours: 22:00 - 08:00
		start, _ := time.Parse("15:04", "22:00")
		end, _ := time.Parse("15:04", "08:00")
		prefs.SetQuietHours(start, end)
		prefs.Timezone = "Europe/Kiev"

		err := prefRepo.Create(ctx, prefs)
		require.NoError(t, err)

		// Retrieve and verify
		result, err := prefRepo.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, result.QuietHoursStart)
		require.NotNil(t, result.QuietHoursEnd)
		assert.Equal(t, "Europe/Kiev", result.Timezone)

		// Test time at 23:00 (should be in quiet hours)
		testTime, _ := time.Parse("15:04", "23:00")
		assert.True(t, result.IsInQuietHours(testTime))

		// Test time at 12:00 (should NOT be in quiet hours)
		testTime2, _ := time.Parse("15:04", "12:00")
		assert.False(t, result.IsInQuietHours(testTime2))

		// Test time at 02:00 (should be in quiet hours - spans midnight)
		testTime3, _ := time.Parse("15:04", "02:00")
		assert.True(t, result.IsInQuietHours(testTime3))
	})
}
